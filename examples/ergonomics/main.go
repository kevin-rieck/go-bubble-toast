package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kevin-rieck/go-bubble-toast"
)

const syncStatusID = "sync-status"

type syncDoneMsg struct{}
type undoDeleteMsg struct{}

type model struct {
	toasts toast.Model
	ascii  bool
	count  int
}

func main() {
	m := newModel(false)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newModel(ascii bool) model {
	options := []toast.Option{
		toast.WithPlacement(toast.TopRight),
		toast.WithRendererPreset(toast.PresetIcon),
		toast.WithKindIcons(),
		toast.WithKindPriority(toast.KindError, 10),
		toast.WithMaxVisible(2),
		toast.WithMaxQueued(3),
		toast.WithQueueOverflowPolicy(toast.DropNewestToast),
		toast.WithDuplicateCoalescing(),
		toast.WithQueueIndicator(func(ctx toast.QueueIndicatorContext) string {
			return fmt.Sprintf("+%d waiting", ctx.Count)
		}),
	}
	if ascii {
		options = append(options, toast.WithASCIIOnly(), toast.WithNoColor())
	}
	return model{toasts: toast.New(options...), ascii: ascii}
}

func (m model) Init() tea.Cmd {
	return toast.Show(toast.Info("press keys to try Toast ergonomics", toast.WithTitle("Ergonomics example")))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var toastCmd tea.Cmd
	m.toasts, toastCmd = m.toasts.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, tea.Batch(toastCmd, toast.DismissNewest())
		case "s":
			return m, tea.Batch(toastCmd, startSync())
		case "b":
			m.count++
			return m, tea.Batch(toastCmd, burst(m.count))
		case "d":
			return m, tea.Batch(toastCmd, deleteFile())
		case "a":
			m = newModel(!m.ascii)
			return m, toast.Show(toast.Info("compatibility mode toggled"))
		}
	case syncDoneMsg:
		return m, tea.Batch(toastCmd, toast.Replace(syncStatusID, toast.Success("sync complete")))
	case undoDeleteMsg:
		return m, tea.Batch(toastCmd, toast.Show(toast.Success("delete undone")))
	}

	return m, toastCmd
}

func (m model) View() string {
	compat := "unicode/color"
	if m.ascii {
		compat = "ASCII/no-color"
	}
	base := strings.Join([]string{
		"Bubble Toast ergonomics example",
		"",
		"s: long-running status replacement",
		"b: burst queue, priority, and duplicate coalescing",
		"d: action Toast with undo key",
		"esc: dismiss newest visible Toast",
		"a: toggle compatibility mode (" + compat + ")",
		"q: quit",
	}, "\n")
	return m.toasts.Overlay(base)
}

func startSync() tea.Cmd {
	return tea.Sequence(
		toast.Replace(syncStatusID, toast.Info("syncing…", toast.WithPersistent())),
		tea.Tick(1200*time.Millisecond, func(time.Time) tea.Msg { return syncDoneMsg{} }),
	)
}

func burst(n int) tea.Cmd {
	return tea.Sequence(
		toast.Show(toast.Info("background refresh", toast.WithPriority(1))),
		toast.Show(toast.Warning("retrying network request", toast.WithPriority(2))),
		toast.Show(toast.Warning("retrying network request", toast.WithPriority(2))),
		toast.Show(toast.Error(fmt.Sprintf("critical failure #%d", n))),
	)
}

func deleteFile() tea.Cmd {
	return toast.Show(toast.Warning(
		"file deleted",
		toast.WithTitle("Undo available"),
		toast.WithAction("u", "undo", func() tea.Msg { return undoDeleteMsg{} }),
		toast.WithDuration(8*time.Second),
	))
}
