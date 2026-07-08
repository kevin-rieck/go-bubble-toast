package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kevin-rieck/go-bubble-toast"
)

const syncID = "sync-status"

type scene int

const (
	sceneReady scene = iota
	sceneSync
	sceneBurst
	sceneAction
	scenePlacement
)

type timedMsg string
type undoMsg struct{}

type model struct {
	toasts toast.Model
	scene  scene
	step   int
	logs   []string
}

func main() {
	m := model{scene: sceneReady}
	m.configureToasts(toast.TopRight, nil)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return tea.Sequence(
		toast.Show(toast.Info("dashboard online", toast.WithTitle("Bubble Toast"), toast.WithDuration(5*time.Second))),
		toast.Show(toast.Success("watch notifications float over the app", toast.WithDuration(5*time.Second))),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var toastCmd tea.Cmd
	m.toasts, toastCmd = m.toasts.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "s":
			m.scene = sceneSync
			m.step = 0
			m.logs = appendLog(m.logs, "sync workflow started")
			m.configureToasts(toast.TopCenter, nil)
			return m, tea.Batch(toastCmd, syncScene())
		case "b":
			m.scene = sceneBurst
			m.logs = appendLog(m.logs, "burst: queue + priority + coalescing")
			m.configureBurstToasts()
			return m, tea.Batch(toastCmd, burstScene())
		case "d":
			m.scene = sceneAction
			m.logs = appendLog(m.logs, "delete action waiting for undo")
			m.configureToasts(toast.TopRight, nil)
			return m, tea.Batch(toastCmd, deleteScene())
		case "p":
			m.scene = scenePlacement
			m.step = 0
			m.logs = appendLog(m.logs, "placement tour started")
			m.configureToasts(toast.TopRight, nil)
			return m, tea.Batch(toastCmd, placementSceneStart())
		case "u":
			return m, toastCmd
		}
	case timedMsg:
		return m.handleTimed(msg, toastCmd)
	case undoMsg:
		m.logs = appendLog(m.logs, "undo command fired")
		return m, tea.Batch(toastCmd, toast.Show(toast.Success("delete undone", toast.WithTitle("Action complete"))))
	}

	return m, toastCmd
}

func (m model) handleTimed(msg timedMsg, toastCmd tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg {
	case "sync-40":
		m.logs = appendLog(m.logs, "sync progress: 40%")
		return m, tea.Batch(toastCmd, toast.Replace(syncID, toast.Info("syncing records… 40%", toast.WithTitle("Stable Toast ID"), toast.WithPersistent())), after(750*time.Millisecond, "sync-85"))
	case "sync-85":
		m.logs = appendLog(m.logs, "sync progress: 85%")
		return m, tea.Batch(toastCmd, toast.Replace(syncID, toast.Warning("retrying slow endpoint", toast.WithTitle("Stable Toast ID"), toast.WithPersistent())), after(750*time.Millisecond, "sync-done"))
	case "sync-done":
		m.logs = appendLog(m.logs, "sync complete")
		return m, tea.Batch(toastCmd, toast.Replace(syncID, toast.Success("all records synced", toast.WithTitle("Stable Toast ID"), toast.WithDuration(5*time.Second))))
	case "place-center":
		m.configureToasts(toast.TopCenter, nil)
		m.logs = appendLog(m.logs, "placement: top-center")
		return m, tea.Batch(toastCmd, toast.Show(toast.Info("top-center status", toast.WithDuration(4*time.Second))), after(1100*time.Millisecond, "place-left"))
	case "place-left":
		m.configureToasts(toast.BottomLeft, nil)
		m.logs = appendLog(m.logs, "placement: bottom-left")
		return m, tea.Batch(toastCmd, toast.Show(toast.Warning("bottom-left alert", toast.WithDuration(4*time.Second))), after(1100*time.Millisecond, "place-custom"))
	case "place-custom":
		m.configureToasts(toast.BottomCenter, customRenderer)
		m.logs = appendLog(m.logs, "custom renderer finale")
		return m, tea.Batch(toastCmd, toast.Show(toast.NewToast("custom renderer", toast.WithTitle("Bottom center"), toast.WithDuration(5*time.Second))))
	default:
		return m, toastCmd
	}
}

func (m model) View() string {
	return m.toasts.Overlay(dashboard(m))
}

func (m *model) configureToasts(placement toast.Placement, renderer toast.Renderer) {
	options := []toast.Option{
		toast.WithPlacement(placement),
		toast.WithWidth(38),
		toast.WithMaxVisible(3),
		toast.WithMaxQueued(6),
		toast.WithOverlayMargin(2, 3, 2, 3),
		toast.WithRendererPreset(toast.PresetIcon),
		toast.WithKindIcons(),
		toast.WithProgress(true),
	}
	if renderer != nil {
		options = append(options, toast.WithRenderer(renderer))
	}
	m.toasts = toast.New(options...)
}

func (m *model) configureBurstToasts() {
	m.toasts = toast.New(
		toast.WithPlacement(toast.BottomRight),
		toast.WithWidth(38),
		toast.WithMaxVisible(2),
		toast.WithMaxQueued(4),
		toast.WithOverlayMargin(2, 3, 2, 3),
		toast.WithRendererPreset(toast.PresetIcon),
		toast.WithKindIcons(),
		toast.WithKindPriority(toast.KindError, 10),
		toast.WithDuplicateCoalescing(),
		toast.WithQueueIndicator(func(ctx toast.QueueIndicatorContext) string {
			return fmt.Sprintf("+%d waiting", ctx.Count)
		}),
	)
}

func syncScene() tea.Cmd {
	return tea.Batch(
		toast.Replace(syncID, toast.Info("sync queued…", toast.WithTitle("Stable Toast ID"), toast.WithPersistent())),
		after(750*time.Millisecond, "sync-40"),
	)
}

func burstScene() tea.Cmd {
	return tea.Sequence(
		toast.Show(toast.Info("index refreshed", toast.WithPersistent())),
		toast.Show(toast.Success("cache warmed", toast.WithPersistent())),
		toast.Show(toast.Warning("network retry", toast.WithPriority(2), toast.WithPersistent())),
		toast.Show(toast.Warning("network retry", toast.WithPriority(2), toast.WithPersistent())),
		toast.Show(toast.Warning("network retry", toast.WithPriority(2), toast.WithPersistent())),
		toast.Show(toast.Error("payment sync failed", toast.WithPersistent())),
	)
}

func deleteScene() tea.Cmd {
	return toast.Show(toast.Warning(
		"invoice deleted",
		toast.WithTitle("Undo available"),
		toast.WithAction("u", "undo", func() tea.Msg { return undoMsg{} }),
		toast.WithDuration(7*time.Second),
	))
}

func placementSceneStart() tea.Cmd {
	return tea.Batch(
		toast.Show(toast.Success("top-right notification", toast.WithDuration(4*time.Second))),
		after(1100*time.Millisecond, "place-center"),
	)
}

func after(d time.Duration, msg timedMsg) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return msg })
}

func customRenderer(t toast.Toast, ctx toast.RenderContext) string {
	box := lipgloss.NewStyle().
		Width(ctx.Width-2).
		Padding(0, 1).
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("99")).
		Foreground(lipgloss.Color("225"))
	return box.Render(fmt.Sprintf("✦ %s\n%s", t.Title, t.Message))
}

func dashboard(m model) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("Bubble Toast")
	subtitle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("workflow notifications for Bubble Tea")
	sceneLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(sceneName(m.scene))

	leftPanel := panel("Sync Dashboard", strings.Join([]string{
		"Status      online",
		"Region      eu-west-1",
		"Queue       12 jobs",
		"Latency     84ms",
		"Workers     8 active",
	}, "\n"), 34)

	rightPanel := panel("Event Log", strings.Join(paddedLogs(m.logs), "\n"), 42)

	chart := lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Render(strings.Join([]string{
		"Throughput",
		"▁▃▅▆▇▆▅▇▆▇▅▆▇",
		"Errors",
		"▁▁▂▁▁▃▁▁▁▂▁▁▁",
	}, "\n"))

	controls := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("s sync · b burst · d delete · u undo · p placements · q quit")

	return strings.Join([]string{
		fmt.Sprintf("%s  %s", title, subtitle),
		sceneLabel,
		"",
		lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel),
		"",
		panel("Live Metrics", chart, 80),
		"",
		controls,
		strings.Repeat("\n", 8),
	}, "\n")
}

func panel(title, body string, width int) string {
	return lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Render(lipgloss.NewStyle().Bold(true).Render(title) + "\n" + body)
}

func sceneName(s scene) string {
	switch s {
	case sceneSync:
		return "Scene 1 — replacement by Toast ID + progress"
	case sceneBurst:
		return "Scene 2 — queue indicator + priority + duplicate coalescing"
	case sceneAction:
		return "Scene 3 — keyboard action Toast"
	case scenePlacement:
		return "Scene 4 — placement + custom rendering"
	default:
		return "Ready — press s, b, d, or p"
	}
}

func appendLog(logs []string, line string) []string {
	logs = append(logs, time.Now().Format("15:04:05")+"  "+line)
	if len(logs) > 5 {
		return logs[len(logs)-5:]
	}
	return logs
}

func paddedLogs(logs []string) []string {
	if len(logs) == 0 {
		logs = []string{"waiting for demo input"}
	}
	out := append([]string(nil), logs...)
	for len(out) < 5 {
		out = append(out, "")
	}
	return out
}
