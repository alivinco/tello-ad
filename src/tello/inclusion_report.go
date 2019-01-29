package tello

import (
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/fimptype"
)

func (svc *TelloService) registerDrone() {

	outLvlSwitchInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.binary.set",
		ValueType: "bool",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.lvl.set",
		ValueType: "int",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.lvl.start",
		ValueType: "string",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.lvl.stop",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.lvl.report",
		ValueType: "int",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.binary.report",
		ValueType: "bool",
		Version:   "1",
	}}

	outBinSwitchInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.binary.set",
		ValueType: "bool",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.binary.get_report",
		ValueType: "int",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.binary.report",
		ValueType: "bool",
		Version:   "1",
	}}

	batteryInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.lvl.get_report",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.lvl.report",
		ValueType: "int",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.alarm.report",
		ValueType: "str_map",
		Version:   "1",
	}}

	cameraInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.camera.get_image",
		ValueType: "string",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.camera.image",
		ValueType: "string",
		Version:   "1",
	}}

	droneInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.position.start_move",
		ValueType: "str_map",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "cmd.position.change_to",
		ValueType: "str_map",
		Version:   "1",
	},{
		Type:      "out",
		MsgType:   "cmd.mode.set",
		ValueType: "string",
		Version:   "1",
	}}
	droneService := fimptype.Service{
		Name:    "drone",
		Alias:   "Drone control",
		Address: "/rt:dev/rn:tello/ad:1/sv:drone/ad:1_0",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{
			"sup_modes": []string{"take_off","throw_take_off","land","palm_land","stop_landing","right_flip","left_flip","set_home","back_flip","bounce","reconnect"},
			"sup_moves": []string{"right","left","up","down","forw","back","yaw"},
		},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       droneInterfaces,
	}

	outLvlSwitchService := fimptype.Service{
		Name:    "out_lvl_switch",
		Alias:   "Drone rotation and camera",
		Address: "/rt:dev/rn:tello/ad:1/sv:out_lvl_switch/ad:1_0",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{
			"max_lvl": 100,
			"min_lvl": 0,
		},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       outLvlSwitchInterfaces,
	}

	batteryService := fimptype.Service{
		Name:    "battery",
		Alias:   "Drone battery",
		Address: "/rt:dev/rn:tello/ad:1/sv:battery/ad:1_0",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{
		},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       batteryInterfaces,
	}

	outBinSwitchService := fimptype.Service{
		Name:    "out_bin_switch",
		Alias:   "Drone takeoff and landing",
		Address: "/rt:dev/rn:tello/ad:1/sv:out_bin_switch/ad:1_1",
		Enabled: true,
		Groups:  []string{"ch_1"},
		Props:  map[string]interface{}{},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       outBinSwitchInterfaces,
	}

	cameraService := fimptype.Service{
		Name:    "camera",
		Alias:   "Drone camera",
		Address: "/rt:dev/rn:tello/ad:1/sv:camera/ad:1_0",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props:  map[string]interface{}{},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       cameraInterfaces,
	}

	services := []fimptype.Service{outLvlSwitchService,outBinSwitchService,cameraService,droneService,batteryService}

	inclReport := fimptype.ThingInclusionReport{
		IntegrationId:     "",
		Address:           "1",
		Type:              "",
		ProductHash:       "TELLO-DR-1",
		Alias:             "Tello drone",
		CommTechnology:    "tello",
		ProductId:         "T1",
		ProductName:       "TELLO",
		ManufacturerId:    "rezen",
		DeviceId:          "dr1",
		HwVersion:         "1",
		SwVersion:         "1",
		PowerSource:       "battery",
		WakeUpInterval:    "-1",
		Security:          "",
		Tags:              nil,
		Groups:            []string{"ch_0","ch_1"},
		PropSets:          nil,
		TechSpecificProps: nil,
		Services:          services,
	}

	msg := fimpgo.NewMessage("evt.thing.inclusion_report", "tello", fimpgo.VTypeObject, inclReport, nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "tello", ResourceAddress: "1"}
	svc.mq.Publish(&adr, msg)
}