package stack

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/data"
)

type StackComponent struct {
	Data data.StackData
}

func NewStackComponent(stack data.StackData) StackComponent {
	return StackComponent{
		Data: stack,
	}
}

func (s StackComponent) Render() string {
	style := lipgloss.NewStyle().
		PaddingLeft(2).
		MarginRight(2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("62")) // Hellblau
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")) // Fast Weiß
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")) // Grau

	// Konstruieren Sie den darzustellenden Inhalt
	content := lipgloss.JoinHorizontal(lipgloss.Top,
		titleStyle.Render(s.Data.MetaName),
		statusStyle.Render(string(s.Data.Status)),
	)

	return style.Render(content)
}