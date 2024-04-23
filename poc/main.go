package main

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/browser"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Auth
type CredentialData struct {
	IDToken string `json:"id_token"`
}

func LoadCredentials(filepath string) (string, error) {
	var creds CredentialData
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", err
	}
	return creds.IDToken, nil
}

// Define Type for Stacks and Stacks API Response
type StackStatus string

const (
	StatusAll       StackStatus = "all"
	StatusUnhealthy StackStatus = "unhealthy"
	StatusHealthy   StackStatus = "healthy"
	StatusDrifted   StackStatus = "drifted"
	StatusFailed    StackStatus = "failed"
	StatusOK        StackStatus = "ok"
)

type DeploymentStatus string

const (
	DeploymentCanceled DeploymentStatus = "canceled"
	DeploymentFailed   DeploymentStatus = "failed"
	DeploymentOK       DeploymentStatus = "ok"
	DeploymentPending  DeploymentStatus = "pending"
	DeploymentRunning  DeploymentStatus = "running"
)

type DriftStatus string

const (
	DriftOK      DriftStatus = "ok"
	DriftDrifted DriftStatus = "drifted"
	DriftFailed  DriftStatus = "failed"
)

type Stack struct {
	StackID          int              `json:"stack_id"`
	Repository       string           `json:"repository"`
	Path             string           `json:"path"`
	DefaultBranch    string           `json:"default_branch"`
	MetaID           string           `json:"meta_id"`
	MetaName         string           `json:"meta_name"`
	MetaDescription  string           `json:"meta_description"`
	MetaTags         []string         `json:"meta_tags"`
	Status           StackStatus      `json:"status"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	SeenAt           time.Time        `json:"seen_at"`
	DeploymentStatus DeploymentStatus `json:"deployment_status"`
	DriftStatus      DriftStatus      `json:"drift_status"`
	Draft            bool             `json:"draft"`
}

type StackAPIResponse struct {
	Stacks          []Stack `json:"stacks"`
	PaginatedResult struct {
		Total   int `json:"total"`
		Page    int `json:"page"`
		PerPage int `json:"per_page"`
	} `json:"paginated_result"`
}

type KeyMap struct {
	Open key.Binding
}

func newKeyMap() KeyMap {
	return KeyMap{
		Open: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open in browser"),
		),
	}
}

//Styling

var (
	red    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render
	green  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render
	yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render

	pathStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF")).Render
	repoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#C0C0C0")).Render
)

func (s Stack) FilterValue() string {
	return fmt.Sprintf("%s %s %s", s.MetaName, s.Status, s.DriftStatus)
}

func (s Stack) Title() string {
	statusText := s.colorizeStatus(s.Status)
	driftStatusText := s.colorizeDriftStatus(s.DriftStatus)
	return fmt.Sprintf("%s\tStatus:%s\tDrift:%s", s.MetaName, statusText, driftStatusText)
}

func (s Stack) Description() string {
	trimmedRepo := strings.TrimPrefix(s.Repository, "github.com/")
	return fmt.Sprintf("Repo: %s Path: %s", repoStyle(trimmedRepo), pathStyle(s.Path))

}

func (s Stack) colorizeStatus(status StackStatus) string {
	switch status {
	case StatusOK:
		return green(string(status))
	case StatusFailed:
		return red(string(status))
	default:
		return yellow(string(status))

	}
}

func (s Stack) colorizeDriftStatus(status DriftStatus) string {
	switch status {
	case DriftOK:
		return green(string(status))
	case DriftDrifted:
		return red(string(status))
	default:
		return yellow(string(status))
	}
}

type Model struct {
	list list.Model
	err  error
	keys KeyMap
}

func New() *Model {
	return &Model{
		keys: newKeyMap(),
	}
}

func fetchStacks(client *http.Client, token string) ([]Stack, error) {
	request, err := http.NewRequest("GET", "http://api.terramate.io/v1/stacks/5fbadfe9-b35b-4352-aadf-b03ee7a0a0c0", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var apiResponse StackAPIResponse

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	sort.Slice(apiResponse.Stacks, func(i, j int) bool {
		return apiResponse.Stacks[i].UpdatedAt.After(apiResponse.Stacks[j].UpdatedAt)
	})

	return apiResponse.Stacks, nil

}

func (m *Model) initList(width, height int, token string) {
	m.keys = newKeyMap()
	client := &http.Client{}

	// Fetch Stacks
	stacks, err := fetchStacks(client, token)
	if err != nil {
		m.err = err
		return
	}

	//Convert Stacks to []list.Item

	var items []list.Item
	for _, stack := range stacks {
		items = append(items, stack)
	}

	m.list = list.New(items, list.NewDefaultDelegate(), width, height)
	m.list.Title = "Terramate Cloud Stacks"
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			m.keys.Open,
		}
	}
}

type CustomDelegate struct {
	list.DefaultDelegate
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Fehler beim Abrufen des Home-Verzeichnisses: %v", err)
		}

		// Construct the full path to the credentials file
		credentialPath := filepath.Join(homeDir, ".terramate.d", "credentials.tmrc.json")
		token, err := LoadCredentials(credentialPath)
		if err != nil {
			fmt.Printf("Could not load Credentials %v", err)
		}
		m.initList(msg.Width, msg.Height, token)

	case tea.KeyMsg:
		switch msg.String() {
		case "o":
			if selectedItem, ok := m.list.SelectedItem().(Stack); ok {
				url := fmt.Sprintf("https://cloud.terramate.io/o/terramate-demo/stacks/%d", selectedItem.StackID)
				if err := browser.OpenURL(url); err != nil {
					log.Printf("Failed to open URL: %v", err)
				}
			}
		}

	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return m.list.View()
}

func main() {
	m := New()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
