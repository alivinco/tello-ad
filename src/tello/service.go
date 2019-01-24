package tello

import (
	"github.com/futurehomeno/fimpgo"
	log "github.com/sirupsen/logrus"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
)

type TelloService struct {
	drone *tello.Driver
	mq *fimpgo.MqttTransport
	inboundMsgCh fimpgo.MessageCh
	robot *gobot.Robot
}

func NewTelloService(transport *fimpgo.MqttTransport) *TelloService {
	svc := &TelloService{mq:transport,inboundMsgCh: make(fimpgo.MessageCh,5)}
	svc.mq.RegisterChannel("ch1",svc.inboundMsgCh)
	svc.drone = tello.NewDriver("8899")
	svc.robot = gobot.NewRobot("tello",
		[]gobot.Connection{},
		[]gobot.Device{svc.drone},
		svc.onDroneReady,
	)
	svc.drone.Start()
	return svc
}

func(svc *TelloService) Start() {
	go func(msgChan fimpgo.MessageCh) {
		for  {
			select {
			case newMsg :=<- msgChan:
				svc.routeFimpMessage(newMsg)

			}
		}

	}(svc.inboundMsgCh)
}

func (svc *TelloService) routeFimpMessage(newMsg *fimpgo.Message) {
	log.Debug("New fimp msg")
	switch newMsg.Payload.Service {
	case "out_lvl_switch" :

	case "out_bin_switch":
		switch newMsg.Payload.Type {
		case "cmd.binary.set":
			// response evt.network.all_nodes_report
			val , _ := newMsg.Payload.GetBoolValue()
			if val {
				log.Debug("Take off command")
				svc.drone.TakeOff()
			}else {
				log.Debug("Land command")
				svc.drone.Land()
			}
		}
		log.Debug("Sending switch")


		//
	case "tello":
		switch newMsg.Payload.Type {
		case "cmd.network.get_all_nodes":
			// response evt.network.all_nodes_report
		case "cmd.thing.inclusion":
			// open/close network
		case "cmd.thing.remove":
			// remove device from network
		}
		//

	}

}

func (svc *TelloService) onDroneReady(){
	svc.drone.On(tello.ConnectedEvent, func(data interface{}) {
		log.Info("------- DRONE connected------------- ")
	})
	svc.drone.On(tello.FlightDataEvent, func(data interface{}) {
		log.Info("------- FLight data event------------- ")

	})

	log.Info("Drone is ready .")

}

func (svc *TelloService) droneEventRouter(){

}