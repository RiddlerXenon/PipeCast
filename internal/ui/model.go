package ui

import (
	"github.com/RiddlerXenon/PipeCast/internal/device"
)

type menuItem int

const (
	menuSelectDevices menuItem = iota
	menuRemoveLinks
	menuDevicesTab
	menuVersion
	menuQuit
)

var menuItems = []string{
	"Выбрать устройства для привязки",
	"Удалить ссылки (перезапуск PipeWire)",
	"Доступные устройства",
	"Инфо о версии",
	"Выход",
}

type tab int

const (
	tabMenu tab = iota
	tabDevices
)

type Model struct {
	cursor       int
	status       string
	json         string
	devices      []device.Device
	selected     map[int]struct{}
	inSelectMode bool
	quitting     bool
	confirm      bool
	activeTab    tab

	version   string
	buildTime string
	commit    string
}

func InitialModel(version, buildTime, commit string) Model {
	return Model{
		selected:  make(map[int]struct{}),
		activeTab: tabMenu,
		version:   version,
		buildTime: buildTime,
		commit:    commit,
	}
}
