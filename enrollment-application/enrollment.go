package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func register(address string, account string) (result map[string]string, err error) {

	if address == "" {
		err = fmt.Errorf("Address not informed")
		return
	}

	if account == "" {
		err = fmt.Errorf("Account not informed")
		return
	}

	if agentInstance != nil {
		agentInstance.SaveConfig(address, account)
	}

	logrus.Infof("Initializing registration with address:%s account:%s\n", address, account)

	schematicAddress := address
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		schematicAddress = "http://" + address
	}
	generateMachineIDURL := fmt.Sprintf("%s/account/%s/machine", schematicAddress, account)
	req, err := http.NewRequest("POST", generateMachineIDURL, nil)
	if err != nil {
		logrus.Errorf("Error preparing request: %s", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		logrus.Errorf("Error requesting: %s", err)
		return
	}
	json.NewDecoder(response.Body).Decode(&result)
	if response.StatusCode != 200 {
		err = fmt.Errorf("Error requesting: %d", response.StatusCode)
		return
	}

	logrus.Infof("SSH Information received. Host: %s | Port: %s | User: %s | Password: %s | TunnelPort: %s", result["sshHost"], result["sshPortInternal"], result["machineId"], account, result["tunnelPort"])
	logrus.Debugf("JWT Token: %s", result["token"])
	logrus.Debugf("TOTP Image: %s", result["totpUrl"])

	// err = estabelishSSHTunnel(result["sshHost"], result["sshPort"], result["machineId"], account, result["tunnelPort"])
	err = estabelishSSHTunnel(result["sshHost"], result["sshPortInternal"], result["machineId"], result["token"], result["tunnelPort"])
	logrus.Infof("Connection estabilished to SSH server tunneling to port %s", result["tunnelPort"])
	return
}

//estabelishSSHTunnel has the SSH logic with remote tunnel
func estabelishSSHTunnel(sshHost string, sshPort string, sshUser string, sshPassword string, tunnelPort string) error {
	sshPortInt, err := strconv.Atoi(sshPort)
	if err != nil {
		return err
	}
	tunnelPortInt, err := strconv.Atoi(tunnelPort)
	if err != nil {
		return err
	}
	go serve(sshHost, sshPortInt, sshUser, sshPassword, "127.0.0.1", 3389, "127.0.0.1", tunnelPortInt)
	return nil
}
