package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kevin-rieck/go-bubble-toast"
)

type model struct {
	toasts toast.Model
	count  int
}

func main() {
	if os.Getenv("BUBBLE_TOAST_DEBUG") != "" {
		f, err := tea.LogToFile("bubble-toast-debug.log", "bubble-toast")
		if err == nil {
			defer f.Close()
		}
	}
	// Enable the reverse progress bar
	m := model{toasts: toast.New(toast.WithProgress(true), toast.WithDefaultDuration(4*time.Second))}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return toast.Show(toast.Info("press t for a Toast, q to quit", toast.WithTitle("Progress Bar Enabled"), toast.WithDuration(5*time.Second)))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.toasts, cmd = m.toasts.Update(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "t":
			m.count++
			return m, tea.Batch(cmd, toast.Show(toast.Success(fmt.Sprintf("Toast with progress #%d", m.count))))
		}
	}
	return m, cmd
}

func (m model) View() string {
	base := "Bubble Toast progress example\n\nPress t to show a Toast. Press q to quit."
	return m.toasts.Overlay(base)
}
