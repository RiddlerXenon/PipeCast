package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/RiddlerXenon/PipeCast/internal/monitor"
	"github.com/RiddlerXenon/PipeCast/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	version   = "dev"
	buildTime = "unknown"
	commit    = "none"
)

func main() {
	fmt.Printf("Version: %s\nBuild time: %s\nCommit: %s\n", version, buildTime, commit)
	go monitor.Run()

	p := tea.NewProgram(ui.InitialModel(version, buildTime, commit))
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

	monitor.Stop()
	fmt.Println("Device monitor stopped. Exiting.")
}
