package stackssection

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/data"
)

// StackItem ist eine Wrapper-Struktur um data.StackData, die die list.Item Schnittstelle implementiert.
type StackItem struct {
	Data data.StackData
}

// Title gibt den Titel des StackItems zurück, der in der Liste angezeigt wird.
func (s StackItem) Title() string {
	return s.Data.MetaName // Oder eine andere repräsentative Zeichenkette
}

// Description gibt eine Beschreibung oder zusätzliche Details des StackItems zurück.
func (s StackItem) Description() string {
	return s.Data.MetaDescription // Oder eine andere Detailinformation
}

// FilterValue wird für die Suche innerhalb der Liste verwendet. Hier verwenden wir den Titel.
func (s StackItem) FilterValue() string {
	return s.Title()
}

// Model repräsentiert das Modell der Stacks-Sektion in der TUI.
type Model struct {
	list list.Model
}

// NewModel erstellt und initialisiert ein neues Model für die Stacks-Sektion.
func NewModel(stacks []data.StackData) Model {
	items := make([]list.Item, len(stacks))
	for i, stack := range stacks {
		items[i] = StackItem{Data: stack} // Erstellen von StackItem aus StackData
	}

	listModel := list.New(items, list.NewDefaultDelegate(), 0, 0)
	listModel.Title = "Stacks"

	return Model{
		list: listModel,
	}
}

// Init wird verwendet, um Befehle beim Initialisieren des Modells auszuführen. Hier gibt es keine Initialisierungsbefehle.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update behandelt eingehende Nachrichten und aktualisiert das Modell entsprechend.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.String() == "q" {
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View rendert die visuelle Darstellung der Stacks-Sektion.
func (m Model) View() string {
	return lipgloss.NewStyle().Render(m.list.View())
}