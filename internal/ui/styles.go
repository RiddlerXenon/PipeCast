package ui

import "github.com/charmbracelet/lipgloss"

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
