package device

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
)

type Device struct {
	Id         uint32                 `json:"id"`
	Type       string                 `json:"type"`
	Version    int                    `json:"version"`
	Properties map[string]interface{} `json:"properties"`
}

var virtualNodeName = "virtual-sink"

func LoadDevices() []Device {
	usr, err := user.Current()
	if err != nil {
		return nil
	}
	path := filepath.Join(usr.HomeDir, ".cache", "pipecast", "mon-list.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var arr []Device
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil
	}
	var list []Device
	for _, d := range arr {
		nodeName, _ := d.Properties["node.name"].(string)
		if nodeName == virtualNodeName {
			continue
		}
		list = append(list, d)
	}
	return list
}

func FindVirtualDeviceNodeName() string {
	usr, err := user.Current()
	if err != nil {
		return ""
	}
	path := filepath.Join(usr.HomeDir, ".cache", "pipecast", "mon-list.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var arr []Device
	if err := json.Unmarshal(raw, &arr); err != nil {
		return ""
	}
	for _, d := range arr {
		nodeName, _ := d.Properties["node.name"].(string)
		if nodeName == virtualNodeName {
			return nodeName
		}
	}
	return ""
}
