package ui

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/RiddlerXenon/PipeCast/internal/device"
	"github.com/RiddlerXenon/PipeCast/internal/monitor"
)

// Проверяет, существует ли связь между двумя портами
func isLinkExists(from, to string) bool {
	cmd := exec.Command("pw-link", "-l")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false
	}
	lines := strings.Split(out.String(), "\n")
	search := fmt.Sprintf("%s -> %s", from, to)
	for _, line := range lines {
		if strings.Contains(line, search) {
			return true
		}
	}
	return false
}

// Создаёт связь, если её нет
func (m Model) linkSelectedDevices() string {
	if len(m.selected) == 0 {
		return "Не выбрано ни одного устройства"
	}
	virtualName := device.FindVirtualDeviceNodeName()
	if virtualName == "" {
		return "Виртуальное устройство не найдено!"
	}
	var outMsgs []string
	virtualChannels := []string{"monitor_FL", "monitor_FR"}
	playbackChannels := []string{"playback_FL", "playback_FR"}

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
			if isLinkExists(vchan, schan) {
				outMsgs = append(outMsgs, fmt.Sprintf("Связь уже существует: %s → %s", vchan, schan))
				continue
			}
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

func (m Model) removeLinks() string {
	// Перезапуск PipeWire
	cmd := exec.Command("systemctl", "--user", "restart", "pipewire")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Ошибка перезапуска pipewire: %v\n%s", err, string(output))
	}

	// Запуск горутины для перезапуска мониторинга устройств с небольшой задержкой
	go func() {
		time.Sleep(2 * time.Second)
		monitor.RestartDeviceMonitor()
	}()

	return "PipeWire перезапущен, мониторинг устройств будет обновлён"
}
