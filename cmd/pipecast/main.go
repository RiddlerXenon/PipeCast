package main

/*
#cgo CFLAGS: -I../../internal/c_modules/source
#cgo LDFLAGS: -L../../internal/c_modules/lib -ldm -lpipewire-0.3 -ljson-c
#include "device_monitor.h"

extern struct global_data* init_device_monitor();
struct global_data* data;

void run() {
    data = init_device_monitor();

    if (data == NULL) {
        fprintf(stderr, "Failed to initialize device monitor\n");
        return;
    }
    run_device_monitor(data);
}

void stop() {
    stop_device_monitor(data);
}
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	version   = "dev"
	buildTime = "unknown"
	commit    = "none"
)

const virtualNodeName = "virtual-sink"

var (
	virtualChannels  = []string{"monitor_FL", "monitor_FR"}
	playbackChannels = []string{"playback_FL", "playback_FR"}
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

type model struct {
	cursor       int
	status       string
	json         string
	devices      []Device
	selected     map[int]struct{}
	inSelectMode bool
	quitting     bool
	confirm      bool
	activeTab    tab
}

type Device struct {
	Id         uint32                 `json:"id"`
	Type       string                 `json:"type"`
	Version    int                    `json:"version"`
	Properties map[string]interface{} `json:"properties"`
}

func initialModel() model {
	return model{
		selected:  make(map[int]struct{}),
		activeTab: tabMenu,
	}
}

var (
	menuStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF")).Bold(true)
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF")).SetString("▶")
	statusStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFD700"))
	pickedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")).Bold(true)
	confirmStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF3333")).Bold(true)
	tabActive     = lipgloss.NewStyle().Background(lipgloss.Color("#2222DD")).Foreground(lipgloss.Color("#FFF")).Bold(true)
	tabInactive   = lipgloss.NewStyle().Foreground(lipgloss.Color("#888"))
)

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.activeTab == tabDevices {
			switch msg.String() {
			case "tab", "esc":
				m.activeTab = tabMenu
				return m, nil
			case "r":
				m.status = "Обновление списка устройств..."
				m.devices = loadDevices()
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
				m.devices = loadDevices()
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
					m.devices = loadDevices()
					m.selected = make(map[int]struct{})
					m.inSelectMode = true
					m.cursor = 0
				case int(menuRemoveLinks):
					m.confirm = true
				case int(menuDevicesTab):
					m.activeTab = tabDevices
					m.devices = loadDevices()
					m.cursor = 0
				case int(menuVersion):
					m.status = fmt.Sprintf("Версия: %s | Сборка: %s | Commit: %s", version, buildTime, commit)
				case int(menuQuit):
					m.quitting = true
					return m, tea.Quit
				}
			}
		}
	}
	return m, nil
}

func (m model) View() string {
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

func deviceDisplayName(d Device) string {
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

func loadDevices() []Device {
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

func (m model) linkSelectedDevices() string {
	if len(m.selected) == 0 {
		return "Не выбрано ни одного устройства"
	}
	virtualName := findVirtualDeviceNodeName()
	if virtualName == "" {
		return "Виртуальное устройство не найдено!"
	}
	var outMsgs []string
	for i := range m.selected {
		dev := m.devices[i]
		sourceName, _ := dev.Properties["node.name"].(string)
		if sourceName == "" {
			outMsgs = append(outMsgs, fmt.Sprintf("Пропуск устройства без node.name (id %d)", dev.Id))
			continue
		}
		for idx := range virtualChannels {
			vchan := fmt.Sprintf("%s:%s", virtualName, virtualChannels[idx])
			schan := fmt.Sprintf("%s:%s", sourceName, playbackChannels[idx])
			cmd := exec.Command("pw-link", vchan, schan)
			output, err := cmd.CombinedOutput()
			if err != nil {
				outMsgs = append(outMsgs, fmt.Sprintf("Ошибка для %s: %v\n%s", schan, err, string(output)))
			} else {
				outMsgs = append(outMsgs, fmt.Sprintf("Связано: %s → %s", vchan, schan))
			}
		}
	}
	return strings.Join(outMsgs, "\n")
}

func findVirtualDeviceNodeName() string {
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

func (m model) removeLinks() string {
	cmd := exec.Command("systemctl", "--user", "restart", "pipewire")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Ошибка перезапуска pipewire: %v\n%s", err, string(output))
	}
	// После перезапуска pipewire нужно перезапустить мониторинг
	go restartDeviceMonitor()
	return "PipeWire перезапущен, ссылки удалены и мониторинг устройств перезапущен"
}

func restartDeviceMonitor() {
	// Ждем немного, чтобы pipewire успел подняться
	time.Sleep(2 * time.Second)
	C.stop()
	time.Sleep(1 * time.Second)
	go C.run()
}

func loadJson() string {
	usr, err := user.Current()
	if err != nil {
		return "Не удалось определить пользователя"
	}
	path := filepath.Join(usr.HomeDir, ".cache", "pipecast", "mon-list.json")
	out, err := os.ReadFile(path)
	if err != nil {
		return "Нет данных"
	}
	return string(out)
}

func main() {
	fmt.Printf("Version: %s\nBuild time: %s\nCommit: %s\n", version, buildTime, commit)

	go C.run()

	p := tea.NewProgram(initialModel())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		p.Quit()
	}()

	if err := p.Start(); err != nil {
		fmt.Printf("Ошибка запуска интерфейса: %v\n", err)
		os.Exit(1)
	}

	C.stop()
	fmt.Println("Device monitor stopped. Exiting.")
}
