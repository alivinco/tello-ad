package tello

import (
	"bufio"
	"encoding/base64"
	"github.com/SMerrony/tello"
	"github.com/disintegration/imaging"
	"github.com/futurehomeno/fimpgo"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"runtime"
	"time"
)

type TelloService struct {
	drone        *tello.Tello
	mq           *fimpgo.MqttTransport
	inboundMsgCh fimpgo.MessageCh
	batteryLevel int8
	connectionState string
	isDroneFlying bool
}

func NewTelloService(transport *fimpgo.MqttTransport) *TelloService {
	svc := &TelloService{mq: transport, inboundMsgCh: make(fimpgo.MessageCh, 5)}
	svc.mq.RegisterChannel("ch1", svc.inboundMsgCh)
	svc.drone = &tello.Tello{}

	return svc
}

func (svc *TelloService) Start() {
	svc.connectionState = "CONNECTING"
	go func(msgChan fimpgo.MessageCh) {
		for {
			select {
			case newMsg := <-msgChan:
				svc.routeFimpMessage(newMsg)

			}
		}

	}(svc.inboundMsgCh)
	//svc.sendImage()
	svc.connectToDrone()
	if runtime.GOOS == "linux" {
		svc.monitorWifiConnection()
	}


	log.Info("The system connected to the drone")
	// ask drone for version ever minute . the operatoin should prevent drone from going into power saving mode


	flightData , err :=  svc.drone.StreamFlightData(false,5000)
	if err != nil {
		log.Errorf("Error requesting flight data ",err)
		return
	}
	go func() {
		for {
			data :=<- flightData
			svc.onFlightData(data)

		}
	}()

	svc.startDroneAntiPowerDownProcess()

}

func (svc *TelloService) routeFimpMessage(newMsg *fimpgo.Message) {
	log.Debug("New fimp msg")
	switch newMsg.Payload.Service {
	case "out_lvl_switch":
		switch newMsg.Payload.Type {
		case "cmd.lvl.set":
			log.Debug("Turning")
			val, _ := newMsg.Payload.GetIntValue()
			origVal := val
			//
			val = val - 50 //
			angle := int16(float64(val) * 3.6)
			log.Debug("Angle = ",angle)
			done , _ := svc.drone.AutoTurnToYaw(angle)
			go func() {
				ready := <- done
				log.Debug("Rotation reached the target :",ready)
			}()

			msg := fimpgo.NewIntMessage("evt.lvl.report", "out_lvl_switch", origVal, nil, nil, newMsg.Payload)
			adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "tello", ResourceAddress: "1",ServiceName:"out_lvl_switch",ServiceAddress:"1_0"}
			svc.mq.Publish(&adr, msg)

		case "cmd.binary.set":
			svc.TakePicture()
			val , _ := newMsg.Payload.GetBoolValue()
			msg := fimpgo.NewBoolMessage("evt.binary.report", "out_lvl_switch", val, nil, nil, newMsg.Payload)
			adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "tello", ResourceAddress: "1",ServiceName:"out_lvl_switch",ServiceAddress:"1_0"}
			svc.mq.Publish(&adr, msg)

		case "cmd.lvl.start":
			val, _ := newMsg.Payload.GetStringValue()
			if val == "up" {
				log.Debug("Take off command")
				svc.drone.TakeOff()
				time.Sleep(10 * time.Second)
				svc.drone.AutoTurnToYaw(180)
				time.Sleep(10 * time.Second)
				svc.drone.AutoTurnToYaw(-180)
				log.Debug("Done")
			}else {
				log.Debug("Land command")
				svc.drone.Land()
			}
		case "cmd.lvl.stop":
			log.Debug("Hover command")
			svc.drone.Hover()


		}
	case "out_bin_switch":
		switch newMsg.Payload.Type {
		case "cmd.binary.set":
			// response evt.network.all_nodes_report
			val, _ := newMsg.Payload.GetBoolValue()
			if val {
				log.Debug("Take off command")
				svc.drone.TakeOff()
				svc.isDroneFlying = true
			} else {
				log.Debug("Land command")
				svc.drone.Land()
				svc.isDroneFlying = false
			}
		}

		val , _ := newMsg.Payload.GetBoolValue()
		msg := fimpgo.NewBoolMessage("evt.binary.report", "out_bin_switch", val, nil, nil, newMsg.Payload)
		adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "tello", ResourceAddress: "1",ServiceName:"out_bin_switch",ServiceAddress:"1_1"}
		svc.mq.Publish(&adr, msg)
		log.Debug("Sending switch")
	case "camera":
		switch newMsg.Payload.Type {
		case "cmd.camera.get_image":
			svc.sendImage()
		}
	case "drone":
		switch newMsg.Payload.Type {
		case "cmd.position.start_move":
			val , err := newMsg.Payload.GetIntMapValue()
			if err != nil {
				return
			}
			up , ok := val["up"]
			if ok {
				svc.drone.Up(int(up))
			}
			down , ok := val["down"]
			if ok {
				svc.drone.Down(int(down))
			}
			left , ok := val["left"]
			if ok {
				svc.drone.Left(int(left))
			}
			right , ok := val["right"]
			if ok {
				svc.drone.Right(int(right))
			}
			forward , ok := val["forw"]
			if ok {
				svc.drone.Forward(int(forward))
			}
			backward , ok := val["back"]
			if ok {
				svc.drone.Backward(int(backward))
			}
			yaw , ok := val["yaw"]
			if ok {
				if yaw >= 0 {
					svc.drone.Clockwise(int(yaw))
				}else {
					svc.drone.Anticlockwise(int(yaw))
				}

			}
		case "cmd.mode.set":
			val , err := newMsg.Payload.GetStringValue()
			if err != nil {
				return
			}

			//"take_off","throw_take_off","land","palm_land","stop_landing","right_flip","left_flip","set_home","back_flip","bounce"
			switch val{
			case "take_off":
				svc.drone.TakeOff()
				svc.isDroneFlying = true
			case "throw_take_off":
				svc.drone.ThrowTakeOff()
			case "land":
				svc.drone.Land()
				svc.isDroneFlying = false
			case "palm_land":
				svc.drone.PalmLand()
			case "stop_landing":
				svc.drone.StopLanding()
			case "right_flip":
				svc.drone.RightFlip()
			case "left_flip":
				svc.drone.LeftFlip()
			case "set_home":
				svc.drone.SetHome()
			case "back_flip":
				svc.drone.BackFlip()
			case "bounce":
				svc.drone.Bounce()
			case "reconnect":
				svc.drone.ControlDisconnect()
				time.Sleep(time.Second*1)
				err := svc.drone.ControlConnectDefault()
				if err != nil {
					log.Errorf("Can't connect to drone .Error: %v ", err)
					svc.reportDroneConnectionState("DOWN")
				}else {
					log.Info("Connected.")
					svc.reportDroneConnectionState("UP")
				}

			}
		}
		//
	case "tello":
		switch newMsg.Payload.Type {
		case "cmd.system.connect":
			svc.registerDrone()
			// response evt.network.all_nodes_report
		case "cmd.thing.inclusion":
			// open/close network
		case "cmd.thing.remove":
			// remove device from network
		}
		//

	}

}


func (svc *TelloService) onFlightData(data tello.FlightData) {
	log.Infof("FD height = %d Battery = %d", data.Height,data.BatteryPercentage)
	if data.Height == 0 {
		svc.isDroneFlying = false
	}else {
		svc.isDroneFlying = true
	}
	if svc.batteryLevel != data.BatteryPercentage {
		svc.batteryLevel = data.BatteryPercentage
		msg := fimpgo.NewIntMessage("evt.lvl.report", "battery", int64(svc.batteryLevel), nil, nil, nil)
		adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "tello", ResourceAddress: "1",ServiceName:"battery",ServiceAddress:"1_0"}
		svc.mq.Publish(&adr, msg)
	}
	log.Infof("FD X = %f Y = %f Z = %f" , data.MVO.PositionX,data.MVO.PositionY,data.MVO.PositionZ)

}

func (svc *TelloService) sendImage2() error{
	// Open a test image.
	src, err := imaging.Open("img_0.jpg")
	if err != nil {
		log.Error("failed to open image: %v", err)
	}

	// Crop the original image to 300x300px size using the center anchor.
	//src = imaging.CropAnchor(src, 300, 300, imaging.Center)

	// Resize the cropped image to width = 200px preserving the aspect ratio.
	src = imaging.Resize(src, 200, 0, imaging.Lanczos)

	// Save the resulting image as JPEG.
	err = imaging.Save(src, "img_0_mod.jpg")
	if err != nil {
		log.Fatalf("failed to save image: %v", err)
	}

	return nil
}

func (svc *TelloService) TakePicture() {
	for i:=0;i<10;i++ {
		log.Debug("Taking picture")
		err := svc.drone.TakePicture()
		if err != nil {
			log.Error("Can't take a picture . err:",err)
			return
		}
		count , err := svc.drone.SaveAllPics("img")
		if err != nil {
			log.Error("Can't save picture . err:",err)
			return
		}
		log.Infof("%d images transfered from drone",count)
		if count >0 {
			break
		}
		time.Sleep(1*time.Second)
	}
	svc.sendImage()
}

func (svc *TelloService) sendImage() error{
	// Open file on disk.
	f, err := os.Open("img_0.jpg")

	if err != nil {
		log.Errorf("Can't open image file. Err : ",err)
		return err
	}
	defer f.Close()
	// Read entire JPG into byte slice.
	reader := bufio.NewReader(f)

	img,err :=  ioutil.ReadAll(reader)
	if err != nil {
		log.Errorf("Can't open image file. Err : ",err)
		return err
	}
	encodedImg := base64.StdEncoding.EncodeToString(img)
	//srcImage, _, err := image.Decode(reader)
	//if err != nil {
	//	log.Errorf("Error decoding image . Err:",err)
	//	return err
	//}
	//dstImage := imaging.Resize(srcImage, 400, 0, imaging.Lanczos)
	//buf := new(bytes.Buffer)
	//err = jpeg.Encode(buf, dstImage, &jpeg.Options{100})
	// Encode as base64.
	//encodedImg := base64.StdEncoding.EncodeToString(buf.Bytes())


	msg := fimpgo.NewStringMessage("evt.camera.image", "camera", encodedImg, nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "tello", ResourceAddress: "1",ServiceName:"camera",ServiceAddress:"1_0"}
	svc.mq.Publish(&adr, msg)

	return nil
}



func (svc *TelloService) reportDroneConnectionState(state string) {
	if state == svc.connectionState {
		return
	}
	msg := fimpgo.NewStringMessage("evt.state.report", "dev_sys", state, nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "tello", ResourceAddress: "1",ServiceName:"dev_sys",ServiceAddress:"1_0"}
	svc.mq.Publish(&adr, msg)
	svc.connectionState = state
}

func (svc *TelloService) startDroneAntiPowerDownProcess() {
	go func() {
		for {
			log.Debug("Executing measures to keep drone powered.")
			svc.drone.ThrowTakeOff()
			time.Sleep(40*time.Second)
		}

	}()
}
