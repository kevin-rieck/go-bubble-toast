package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kevin-rieck/go-bubble-toast"
	"github.com/kevin-rieck/go-bubble-toast/examples/internal/layout"
)

const statusID = "interactive-status"

type replaceDoneMsg struct{}
type undoMsg struct{}

type model struct {
	toasts toast.Model
	stage  int
	log    string
	count  int
}

func main() {
	m := model{}
	m.configureStage()
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return tea.Sequence(
		toast.Show(toast.Info("press n to walk through every feature", toast.WithTitle("Interactive Bubble Toast demo"))),
		m.stageCmd(),
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
		case "n", "right":
			m.stage = (m.stage + 1) % len(stages)
			m.configureStage()
			return m, tea.Batch(toastCmd, m.stageCmd())
		case "p", "left":
			m.stage = (m.stage - 1 + len(stages)) % len(stages)
			m.configureStage()
			return m, tea.Batch(toastCmd, m.stageCmd())
		case "r":
			m.configureStage()
			return m, tea.Batch(toastCmd, m.stageCmd())
		case "esc":
			return m, tea.Batch(toastCmd, toast.DismissNewest())
		case "backspace":
			return m, tea.Batch(toastCmd, toast.DismissOldest())
		case "x":
			return m, tea.Batch(toastCmd, toast.DismissAll())
		case "g":
			var id toast.ID
			m.toasts, id, toastCmd = m.toasts.Push(toast.Success("pushed directly with Model.Push", toast.WithTitle("Direct API")))
			if t, ok := m.toasts.Get(string(id)); ok && m.toasts.Has(string(id)) {
				m.log = fmt.Sprintf("Get(%q) => %s", id, t.Message)
			}
			return m, toastCmd
		case "u":
			// Routed through toast.Model.Update first, so action Toasts handle this.
			return m, toastCmd
		}
	case replaceDoneMsg:
		return m, tea.Batch(toastCmd, toast.Replace(statusID, toast.Success("replacement finished", toast.WithTitle("Same Toast ID"))))
	case undoMsg:
		m.log = "undo action command fired and dismissed its Toast"
		return m, tea.Batch(toastCmd, toast.Show(toast.Success("undo complete")))
	}

	return m, toastCmd
}

func (m model) View() string {
	s := stages[m.stage]
	status := ""
	if m.log != "" {
		status = "\nLast API result: " + m.log + "\n"
	}

	base := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render("Bubble Toast interactive feature tour") + "\n\n" +
		lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Render(fmt.Sprintf("%d/%d  %s", m.stage+1, len(stages), s.title)) + "\n" +
		s.description + "\n" + status + "\n" +
		strings.Join([]string{
			"Controls:",
			"  n/right: next feature     p/left: previous     r: replay current",
			"  esc: dismiss newest      backspace: dismiss oldest     x: dismiss all",
			"  g: direct Model.Push/Get/Has demo     u: run undo action when shown",
			"  q: quit",
		}, "\n") + "\n\n" +
		"This tour covers Show/Replace/Dismiss messages, direct Model APIs, Toast kinds, titles, content, IDs,\n" +
		"lifetimes, persistence, progress, queueing, overflow policies, priorities, duplicate coalescing, actions, icons,\n" +
		"disabled affordances, compatibility modes, renderer presets, custom renderers, themes, placement, margins, wrapping, and truncation." +
		strings.Repeat("\n", 10)

	return layout.Columns(base, m.toasts.View())
}

type stage struct {
	title       string
	description string
}

var stages = []stage{
	{"Kinds and lifetimes", "Neutral, info, success, warning, and error Toasts. Defaults, per-kind durations, and progress bars are enabled."},
	{"Titles, content, and stable IDs", "A title+message Toast is followed by pre-rendered Content. The status Toast is replaced in-place by its caller-provided ID."},
	{"Queueing, priority, overflow, and coalescing", "Only two Toasts can be visible. Overflow queues, errors jump ahead by priority, duplicate warnings coalesce, and newest overflow is dropped."},
	{"Actions and dismissal", "An action Toast advertises [u] undo. The toast model consumes the key, runs the command, and dismisses the Toast."},
	{"Placement and layout", "Bottom-center placement demonstrates overlay margins, width wrapping, stack gap, and maximum-height truncation."},
	{"Icon preset and icon overrides", "The icon preset enables Kind icons, and this page overrides one Kind icon."},
	{"Compact preset and disabled affordances", "Compact rendering removes padding. This page disables Kind icons and the queued-count indicator."},
	{"Minimal preset and oldest-overflow queueing", "Minimal rendering removes borders. A full queue drops the oldest queued Toast to preserve the newest one."},
	{"Compatibility and animation", "ASCII-only, no-color, and no-animation modes for constrained terminals. One Toast opts out of progress with WithoutProgress."},
	{"Custom theme and renderer", "A custom theme/style is available to the renderer through RenderContext; the renderer takes ownership of presentation."},
}

func (m *model) configureStage() {
	m.log = ""
	switch m.stage {
	case 0:
		m.toasts = toast.New(
			toast.WithPlacement(toast.TopRight),
			toast.WithMaxVisible(5),
			toast.WithWidth(34),
			toast.WithProgress(true),
			toast.WithDefaultDuration(7*time.Second),
			toast.WithKindDuration(toast.KindError, 10*time.Second),
		)
	case 1:
		m.toasts = toast.New(toast.WithPlacement(toast.TopRight), toast.WithWidth(44), toast.WithMaxVisible(3))
	case 2:
		m.toasts = toast.New(
			toast.WithPlacement(toast.TopRight),
			toast.WithMaxVisible(2),
			toast.WithMaxQueued(3),
			toast.WithQueueOverflowPolicy(toast.DropNewestToast),
			toast.WithKindPriority(toast.KindError, 10),
			toast.WithDuplicateCoalescing(),
			toast.WithQueueIndicator(func(ctx toast.QueueIndicatorContext) string {
				return fmt.Sprintf("queue: %d waiting", ctx.Count)
			}),
		)
	case 3:
		m.toasts = toast.New(toast.WithPlacement(toast.TopRight), toast.WithWidth(42), toast.WithMaxVisible(3))
	case 4:
		m.toasts = toast.New(
			toast.WithPlacement(toast.BottomCenter),
			toast.WithWidth(32),
			toast.WithMaxHeight(4),
			toast.WithGap(0),
			toast.WithOverlayMargin(2, 2, 2, 2),
		)
	case 5:
		m.toasts = toast.New(
			toast.WithPlacement(toast.TopLeft),
			toast.WithMaxVisible(3),
			toast.WithWidth(34),
			toast.WithRendererPreset(toast.PresetIcon),
			toast.WithIcon(toast.KindSuccess, "OK"),
		)
	case 6:
		m.toasts = toast.New(
			toast.WithPlacement(toast.TopRight),
			toast.WithMaxVisible(2),
			toast.WithMaxQueued(2),
			toast.WithRendererPreset(toast.PresetCompact),
			toast.WithoutIcons(),
			toast.WithoutQueueIndicator(),
		)
	case 7:
		m.toasts = toast.New(
			toast.WithPlacement(toast.BottomLeft),
			toast.WithMaxVisible(1),
			toast.WithMaxQueued(2),
			toast.WithQueueOverflowPolicy(toast.DropOldestQueuedToast),
			toast.WithRendererPreset(toast.PresetMinimal),
		)
	case 8:
		m.toasts = toast.New(
			toast.WithPlacement(toast.BottomRight),
			toast.WithWidth(38),
			toast.WithProgress(true),
			toast.WithASCIIOnly(),
			toast.WithNoColor(),
			toast.WithNoAnimation(),
			toast.WithKindIcons(),
		)
	case 9:
		theme := toast.DefaultTheme()
		theme.Info = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)
		m.toasts = toast.New(
			toast.WithPlacement(toast.TopCenter),
			toast.WithWidth(46),
			toast.WithMaxVisible(3),
			toast.WithTheme(theme),
			toast.WithStyle(toast.KindSuccess, lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("42")).Padding(0, 1)),
			toast.WithRenderer(func(t toast.Toast, ctx toast.RenderContext) string {
				return ctx.Style.Render(fmt.Sprintf("CUSTOM %d/%d · %s", ctx.Index+1, ctx.Total, t.Message))
			}),
		)
	}
}

func (m model) stageCmd() tea.Cmd {
	switch m.stage {
	case 0:
		return tea.Sequence(
			toast.Show(toast.NewToast("neutral Toast", toast.WithDuration(7*time.Second))),
			toast.Show(toast.Info("info Toast", toast.WithDuration(7*time.Second))),
			toast.Show(toast.Success("success Toast", toast.WithDuration(7*time.Second))),
			toast.Show(toast.Warning("warning Toast", toast.WithDuration(7*time.Second))),
			toast.Show(toast.Error("error Toast uses a longer kind duration")),
		)
	case 1:
		content := lipgloss.NewStyle().Foreground(lipgloss.Color("213")).Bold(true).Render("pre-rendered Lip Gloss Content wins")
		return tea.Sequence(
			toast.Show(toast.Info("message body", toast.WithTitle("Separate title"), toast.WithDuration(9*time.Second))),
			toast.Show(toast.Success("ignored", toast.WithTitle("Ignored"), toast.WithContent(content), toast.WithDuration(9*time.Second))),
			toast.Replace(statusID, toast.Warning("working…", toast.WithTitle("Stable Toast ID"), toast.WithPersistent())),
			tea.Tick(1200*time.Millisecond, func(time.Time) tea.Msg { return replaceDoneMsg{} }),
		)
	case 2:
		return tea.Sequence(
			toast.Show(toast.Info("visible low priority", toast.WithPersistent())),
			toast.Show(toast.Success("visible normal priority", toast.WithPersistent())),
			toast.Show(toast.Warning("duplicate network retry", toast.WithPriority(2), toast.WithPersistent())),
			toast.Show(toast.Warning("duplicate network retry", toast.WithPriority(2), toast.WithPersistent())),
			toast.Show(toast.Error("error jumps ahead", toast.WithPersistent())),
			toast.Show(toast.Info("dropped newest when queue full", toast.WithPersistent())),
		)
	case 3:
		return toast.Show(toast.Warning(
			"file deleted",
			toast.WithTitle("Action Toast"),
			toast.WithAction("u", "undo", func() tea.Msg { return undoMsg{} }),
			toast.WithDuration(12*time.Second),
		))
	case 4:
		return tea.Sequence(
			toast.Show(toast.Info("A long Toast wraps to the configured width and is truncated to the configured maximum height with an ellipsis so it cannot take over the terminal.", toast.WithTitle("Bottom center layout"), toast.WithDuration(10*time.Second))),
			toast.Show(toast.Success("gap 0 keeps the stack tight", toast.WithDuration(10*time.Second))),
		)
	case 5:
		return tea.Sequence(
			toast.Show(toast.Info("icon preset adds Kind icons", toast.WithDuration(10*time.Second))),
			toast.Show(toast.Success("success icon overridden", toast.WithDuration(10*time.Second))),
			func() tea.Msg {
				return toast.ShowMsg{Toast: toast.NewToast(lipgloss.NewStyle().Italic(true).Render("minimal preset via separate model is shown in README"), toast.WithDuration(10*time.Second))}
			},
		)
	case 6:
		return tea.Sequence(
			toast.Show(toast.Info("compact preset, no icon", toast.WithPersistent())),
			toast.Show(toast.Success("indicator disabled; queued Toasts are hidden", toast.WithPersistent())),
			toast.Show(toast.Warning("queued without +N affordance", toast.WithPersistent())),
		)
	case 7:
		return tea.Sequence(
			toast.Show(toast.Info("visible minimal Toast", toast.WithPersistent())),
			toast.Show(toast.Warning("oldest queued; will be dropped", toast.WithPersistent())),
			toast.Show(toast.Success("newest queued; preserved", toast.WithPersistent())),
			toast.Show(toast.Error("new arrival drops oldest queued", toast.WithPersistent())),
		)
	case 8:
		return tea.Sequence(
			toast.Show(toast.Info("ASCII borders, ASCII icons, no color", toast.WithDuration(10*time.Second))),
			toast.Show(toast.Success("no animation disables progress", toast.WithDuration(10*time.Second))),
			toast.Show(toast.Warning("this Toast also opts out", toast.WithoutProgress(), toast.WithDuration(10*time.Second))),
		)
	case 9:
		return tea.Sequence(
			toast.Show(toast.Info("custom theme style reaches RenderContext", toast.WithDuration(10*time.Second))),
			toast.Show(toast.Success("WithStyle overrides the success style", toast.WithDuration(10*time.Second))),
			toast.Show(toast.NewToast("custom renderer owns this box", toast.WithDuration(10*time.Second))),
		)
	default:
		return nil
	}
}
