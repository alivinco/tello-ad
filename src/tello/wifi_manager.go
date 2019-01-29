package tello

import (
	"os/exec"
	"runtime"
	"strings"
	"time"
	log "github.com/sirupsen/logrus"
)

func (svc *TelloService) getWifiConnStatus()string {
	cmd := exec.Command("sudo","wpa_cli", "-i", "wlan0", "status")
	out ,_ := cmd.CombinedOutput()
	outStr := string(out)
	log.Debug("wifi status ",outStr)
	if strings.Contains(outStr,"wpa_state=COMPLETED") {
		return "UP"
	}else {
		if strings.Contains(outStr,"wlan0  error: No such file or directory") {
			log.Info("Interface is down, recovering")
			cmd := exec.Command("sudo","ifdown", "wlan0")
			out ,_ := cmd.CombinedOutput()
			outStr := string(out)
			log.Debug("wifi Recovery status (ifdown) ",outStr)
			time.Sleep(5*time.Second)
			cmd = exec.Command("sudo","ifup", "wlan0")
			out ,_ = cmd.CombinedOutput()
			outStr = string(out)
			log.Debug("wifi Recovery status (ifup) ",outStr)
		}
		return "DOWN"
	}
}

func (svc *TelloService)connectToDrone(){
	for {
		if runtime.GOOS == "linux" {
			if svc.getWifiConnStatus() == "DOWN" {
				log.Info("Host isn't connected via WiFi to drone , retrying .... ")
				time.Sleep(time.Second*5)
				continue
			}else {
				log.Info("Host is CONNECTED via wifi to drone  ")

			}
		}

		err := svc.drone.ControlConnectDefault()
		if err != nil && !svc.drone.ControlConnected() {
			log.Errorf("Error: %v reconnecting...", err)
			time.Sleep(time.Second*3)
		}else {
			svc.reportDroneConnectionState("UP")
			break
		}
	}
}


func (svc *TelloService) monitorWifiConnection() {
	// TODO : run  "wpa_cli -i wlan0 status" to check wifi connection
	log.Info("Starting WIFI connection monitoring")
	var isConnected bool
	if svc.connectionState == "UP" {
		isConnected = true
	}
	go func() {
		for{

			if svc.getWifiConnStatus() == "UP" {
				if !isConnected {
					// report only if state has changed
					log.Info("Connected to Drone wifi")
					svc.connectToDrone()
				}
				isConnected = true;
			}else {
				if isConnected {
					// report only if state has changed
					svc.reportDroneConnectionState("DOWN")
				}
				isConnected = false;
			}

			time.Sleep(time.Second*60)
		}
	}()
}

