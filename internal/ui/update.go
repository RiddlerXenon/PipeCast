package ui

import (
	"fmt"

	"github.com/RiddlerXenon/PipeCast/internal/device"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.activeTab == tabDevices {
			switch msg.String() {
			case "tab", "esc":
				m.activeTab = tabMenu
				return m, nil
			case "r":
				m.status = "Обновление списка устройств..."
				m.devices = device.LoadDevices()
				m.status = "Устройства обновлены"
				return m, nil
			}
		}
		if m.inSelectMode {
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.devices)-1 {
					m.cursor++
				}
			case " ":
				if _, ok := m.selected[m.cursor]; ok {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = struct{}{}
				}
			case "enter":
				m.status = m.linkSelectedDevices()
				m.inSelectMode = false
				m.cursor = 0
			case "esc", "tab":
				m.inSelectMode = false
				m.cursor = 0
			}
			return m, nil
		}
		if m.confirm {
			switch msg.String() {
			case "y", "Y":
				m.status = m.removeLinks()
				m.confirm = false
			case "n", "N", "esc":
				m.status = "Отменено"
				m.confirm = false
			}
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "tab":
			if m.activeTab == tabMenu {
				m.activeTab = tabDevices
				m.devices = device.LoadDevices()
				m.cursor = 0
			} else {
				m.activeTab = tabMenu
				m.cursor = 0
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.activeTab == tabMenu && m.cursor < len(menuItems)-1 {
				m.cursor++
			} else if m.activeTab == tabDevices && m.cursor < len(m.devices)-1 {
				m.cursor++
			}
		case "enter":
			if m.activeTab == tabMenu {
				switch m.cursor {
				case int(menuSelectDevices):
					m.status = ""
					m.devices = device.LoadDevices()
					m.selected = make(map[int]struct{})
					m.inSelectMode = true
					m.cursor = 0
				case int(menuRemoveLinks):
					m.confirm = true
				case int(menuDevicesTab):
					m.activeTab = tabDevices
					m.devices = device.LoadDevices()
					m.cursor = 0
				case int(menuVersion):
					m.status = fmt.Sprintf("Версия: %s | Сборка: %s | Commit: %s", m.version, m.buildTime, m.commit)
				case int(menuQuit):
					m.quitting = true
					return m, tea.Quit
				}
			}
		}
	}
	return m, nil
}
