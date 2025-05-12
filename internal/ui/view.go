package ui

import (
	"fmt"

	"github.com/RiddlerXenon/PipeCast/internal/device"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.quitting {
		return statusStyle.Render("\nЗавершение работы...\n")
	}
	if m.confirm {
		return confirmStyle.Render("\nУдалить все ссылки и перезапустить PipeWire? (y/n)\n")
	}
	// Вкладки
	tabBar := tabInactive.Render("Меню") + " | " + tabInactive.Render("Устройства")
	if m.activeTab == tabMenu {
		tabBar = tabActive.Render("Меню") + " | " + tabInactive.Render("Устройства")
	} else {
		tabBar = tabInactive.Render("Меню") + " | " + tabActive.Render("Устройства")
	}
	s := tabBar + "\n\n"

	if m.activeTab == tabDevices {
		s += lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#3AF")).
			Render("Доступные устройства (r — обновить, Tab/Esc — назад):") + "\n\n"
		devs := m.devices
		if len(devs) == 0 {
			s += "Нет устройств\n"
		}
		for i, d := range devs {
			line := deviceDisplayName(d)
			if m.cursor == i {
				line = cursorStyle.Render("▶ ") + selectedStyle.Render(line)
			}
			s += line + "\n"
		}
		s += "\n"
		if m.status != "" {
			s += statusStyle.Render(m.status) + "\n"
		}
		return s
	}

	// Меню выбора устройств для привязки
	if m.inSelectMode {
		s += lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#3AF")).
			Render("Выберите устройства для привязки (пробел — выбрать, Enter — подтвердить, Tab/Esc — отмена):") + "\n\n"
		for i, d := range m.devices {
			name := deviceDisplayName(d)
			if _, ok := m.selected[i]; ok {
				name = pickedStyle.Render("[Выбрано] ") + name
			}
			if m.cursor == i {
				line := cursorStyle.Render("▶ ") + name
				s += selectedStyle.Render(line) + "\n"
			} else {
				s += "  " + name + "\n"
			}
		}
		return s
	}

	// Главное меню
	s += lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#3AF")).Render("PipeCast\n") +
		"\nИспользуйте ↑/↓ или j/k для навигации, Enter — выбор, Tab — переключить вкладку:\n\n"
	for i, choice := range menuItems {
		cursor := "  "
		name := menuStyle.Render(choice)
		if m.cursor == i {
			cursor = cursorStyle.Render("▶ ")
			name = selectedStyle.Render(choice)
		}
		s += fmt.Sprintf("%s%s\n", cursor, name)
	}
	s += "\n"
	if m.status != "" {
		s += statusStyle.Render(m.status) + "\n"
	}
	if m.json != "" && m.cursor == int(menuDevicesTab) {
		s += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#0f0")).Render(m.json) + "\n"
	}
	return s
}

func deviceDisplayName(d device.Device) string {
	desc := ""
	if d.Properties != nil {
		if nodeDesc, ok := d.Properties["node.description"].(string); ok && nodeDesc != "" {
			desc = nodeDesc
		} else if nodeName, ok := d.Properties["node.name"].(string); ok && nodeName != "" {
			desc = nodeName
		}
	}
	if desc == "" {
		desc = fmt.Sprintf("id:%d", d.Id)
	}
	return fmt.Sprintf("[%d] %s", d.Id, desc)
}
