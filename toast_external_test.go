package toast_test

import (
	"strconv"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	toast "github.com/kevin-rieck/go-bubble-toast"
)

func TestAppCanQueryVisibleToastByID(t *testing.T) {
	model := toast.New()
	model, _, _ = model.Push(toast.Info("connected", toast.WithID("network")))

	got, ok := model.VisibleByID("network")
	if !ok {
		t.Fatal("expected visible Toast to be found by Toast ID")
	}
	if got.ID != "network" || got.Message != "connected" {
		t.Fatalf("unexpected Toast returned: %#v", got)
	}
	if !model.IsVisible("network") {
		t.Fatal("expected Toast ID to be reported visible")
	}
}

func TestAppCanQueryQueuedToastByID(t *testing.T) {
	model := toast.New(toast.WithMaxVisible(1))
	model, _, _ = model.Push(toast.NewToast("visible", toast.WithID("visible")))
	model, _, _ = model.Push(toast.Warning("waiting", toast.WithID("queued")))

	got, ok := model.QueuedByID("queued")
	if !ok {
		t.Fatal("expected queued Toast to be found by Toast ID")
	}
	if got.ID != "queued" || got.Message != "waiting" {
		t.Fatalf("unexpected Toast returned: %#v", got)
	}
	if !model.IsQueued("queued") {
		t.Fatal("expected Toast ID to be reported queued")
	}
	if model.IsVisible("queued") {
		t.Fatal("queued Toast ID should not be reported visible")
	}
}

func TestToastIDQueriesDoNotFindMissingOrDismissedToasts(t *testing.T) {
	model := toast.New(toast.WithMaxVisible(1))
	model, _, _ = model.Push(toast.NewToast("visible", toast.WithID("visible")))
	model, _, _ = model.Push(toast.NewToast("queued", toast.WithID("queued")))

	if _, ok := model.VisibleByID("missing"); ok {
		t.Fatal("missing Toast ID should not be found in visible Toast Stack")
	}
	if _, ok := model.QueuedByID("missing"); ok {
		t.Fatal("missing Toast ID should not be found in queue")
	}

	model, _ = model.Dismiss("visible")
	model, _ = model.Dismiss("queued")

	if model.IsVisible("visible") || model.IsQueued("queued") {
		t.Fatal("dismissed Toast IDs should not be reported visible or queued")
	}
}

func TestToastIDQueriesReturnCopies(t *testing.T) {
	model := toast.New(toast.WithMaxVisible(1))
	model, _, _ = model.Push(toast.NewToast("visible", toast.WithID("visible")))
	model, _, _ = model.Push(toast.NewToast("queued", toast.WithID("queued")))

	visible, ok := model.VisibleByID("visible")
	if !ok {
		t.Fatal("expected visible Toast")
	}
	visible.Message = "mutated"
	if got, _ := model.VisibleByID("visible"); got.Message == "mutated" {
		t.Fatal("VisibleByID should return a copy")
	}

	queued, ok := model.QueuedByID("queued")
	if !ok {
		t.Fatal("expected queued Toast")
	}
	queued.Message = "mutated"
	if got, _ := model.QueuedByID("queued"); got.Message == "mutated" {
		t.Fatal("QueuedByID should return a copy")
	}
}

func TestPublicDismissNewestRemovesNewestVisibleToastAndDrainsQueue(t *testing.T) {
	model := toast.New(toast.WithPlacement(toast.TopRight), toast.WithMaxVisible(2), toast.WithDefaultDuration(0))
	model, _, _ = model.Push(toast.NewToast("old", toast.WithID("old")))
	model, _, _ = model.Push(toast.NewToast("new", toast.WithID("new")))
	model, _, _ = model.Push(toast.NewToast("queued", toast.WithID("queued")))

	var cmd tea.Cmd
	model, cmd = model.Update(toast.DismissNewest()())

	if cmd == nil {
		t.Fatal("dismissing newest visible Toast should drain queued Toast and schedule its timer")
	}
	visible := model.Visible()
	if len(visible) != 2 || visible[0].ID != "queued" || visible[1].ID != "old" {
		t.Fatalf("newest Dismissal should remove newest top Toast and drain queue, visible=%#v", visible)
	}
}

func TestPublicDismissOldestRemovesOldestVisibleBottomToastAndDrainsQueue(t *testing.T) {
	model := toast.New(toast.WithPlacement(toast.BottomRight), toast.WithMaxVisible(2))
	model, _, _ = model.Push(toast.NewToast("old", toast.WithID("old")))
	model, _, _ = model.Push(toast.NewToast("new", toast.WithID("new")))
	model, _, _ = model.Push(toast.NewToast("queued", toast.WithID("queued")))

	var cmd tea.Cmd
	model, cmd = model.Update(toast.DismissOldest()())

	if cmd == nil {
		t.Fatal("dismissing oldest visible Toast should drain queued Toast and schedule its timer")
	}
	visible := model.Visible()
	if len(visible) != 2 || visible[0].ID != "new" || visible[1].ID != "queued" {
		t.Fatalf("oldest Dismissal should remove oldest bottom Toast and drain queue, visible=%#v", visible)
	}
}

func TestPublicReplaceHelperUpdatesVisibleToastByID(t *testing.T) {
	model := toast.New()
	model, _, _ = model.Push(toast.Info("syncing", toast.WithID("sync-status")))

	var cmd tea.Cmd
	model, cmd = model.Replace("sync-status", toast.Success("synced"))

	if cmd == nil {
		t.Fatal("replacing visible Toast should restart its Dismissal timer")
	}
	visible := model.Visible()
	if len(visible) != 1 {
		t.Fatalf("replacement should preserve one Toast per stable Toast ID, visible=%#v", visible)
	}
	if visible[0].ID != "sync-status" || visible[0].Kind != toast.KindSuccess || visible[0].Message != "synced" {
		t.Fatalf("visible Toast was not replaced by Toast ID: %#v", visible[0])
	}
}

func TestPublicReplaceCommandRoutesThroughBubbleTeaUpdate(t *testing.T) {
	model := toast.New()
	model, _, _ = model.Push(toast.Info("syncing", toast.WithID("sync-status")))

	var cmd tea.Cmd
	model, cmd = model.Update(toast.Replace("sync-status", toast.Success("synced"))())

	if cmd == nil {
		t.Fatal("command-based Replace should restart visible Toast Dismissal")
	}
	visible := model.Visible()
	if len(visible) != 1 || visible[0].ID != "sync-status" || visible[0].Message != "synced" {
		t.Fatalf("command-based Replace did not update stable Toast ID: %#v", visible)
	}
}

func TestPublicReplaceHelperUpdatesQueuedToastInPlace(t *testing.T) {
	model := toast.New(toast.WithMaxVisible(1))
	model, _, _ = model.Push(toast.NewToast("visible", toast.WithID("visible")))
	model, _, _ = model.Push(toast.Info("first queued", toast.WithID("first")))
	model, _, _ = model.Push(toast.Warning("second queued", toast.WithID("second")))

	var cmd tea.Cmd
	model, cmd = model.Replace("first", toast.Success("updated first queued"))

	if cmd != nil {
		t.Fatal("replacing queued Toast should not start a timer until it becomes visible")
	}
	queued := model.Queued()
	if len(queued) != 2 || queued[0].ID != "first" || queued[1].ID != "second" {
		t.Fatalf("queued replacement should preserve queued position, queued=%#v", queued)
	}
	if queued[0].Kind != toast.KindSuccess || queued[0].Message != "updated first queued" {
		t.Fatalf("queued Toast was not replaced in place: %#v", queued[0])
	}
}

func TestKindIconsRenderBuiltInToastKindIcons(t *testing.T) {
	model := toast.New(
		toast.WithMaxVisible(4),
		toast.WithKindIcons(),
		toast.WithStyle(toast.KindInfo, lipgloss.NewStyle()),
		toast.WithStyle(toast.KindSuccess, lipgloss.NewStyle()),
		toast.WithStyle(toast.KindWarning, lipgloss.NewStyle()),
		toast.WithStyle(toast.KindError, lipgloss.NewStyle()),
	)
	model, _, _ = model.Push(toast.Info("info"))
	model, _, _ = model.Push(toast.Success("success"))
	model, _, _ = model.Push(toast.Warning("warning"))
	model, _, _ = model.Push(toast.Error("error"))

	view := model.View()
	for _, want := range []string{"ℹ info", "✓ success", "⚠ warning", "✕ error"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected rendered Toast Stack to contain %q, got %q", want, view)
		}
	}
}

func TestKindIconsCanBeOverridden(t *testing.T) {
	model := toast.New(
		toast.WithKindIcons(),
		toast.WithIcon(toast.KindSuccess, "OK"),
		toast.WithStyle(toast.KindSuccess, lipgloss.NewStyle()),
	)
	model, _, _ = model.Push(toast.Success("saved"))

	view := model.View()
	if !strings.Contains(view, "OK saved") {
		t.Fatalf("expected overridden success icon in rendered Toast, got %q", view)
	}
	if strings.Contains(view, "✓ saved") {
		t.Fatalf("default success icon should be replaced, got %q", view)
	}
}

func TestKindIconsCanBeDisabledGlobally(t *testing.T) {
	model := toast.New(
		toast.WithKindIcons(),
		toast.WithoutIcons(),
		toast.WithStyle(toast.KindWarning, lipgloss.NewStyle()),
	)
	model, _, _ = model.Push(toast.Warning("careful"))

	view := model.View()
	if strings.Contains(view, "⚠") {
		t.Fatalf("expected icons to be disabled, got %q", view)
	}
	if !strings.Contains(view, "careful") {
		t.Fatalf("expected Toast message to remain rendered, got %q", view)
	}
}

func TestKindIconsDoNotChangePreRenderedContent(t *testing.T) {
	model := toast.New(
		toast.WithKindIcons(),
		toast.WithStyle(toast.KindError, lipgloss.NewStyle()),
	)
	model, _, _ = model.Push(toast.Error("ignored", toast.WithTitle("Ignored"), toast.WithContent("custom body")))

	view := model.View()
	if !strings.Contains(view, "custom body") {
		t.Fatalf("expected custom content to render, got %q", view)
	}
	for _, notWant := range []string{"✕", "ignored", "Ignored"} {
		if strings.Contains(view, notWant) {
			t.Fatalf("pre-rendered content should not include %q, got %q", notWant, view)
		}
	}
}

func TestQueueIndicatorShowsQueuedCountAndDisappearsAsQueueDrains(t *testing.T) {
	model := toast.New(
		toast.WithMaxVisible(1),
		toast.WithMaxQueued(3),
		toast.WithStyle(toast.KindNone, lipgloss.NewStyle()),
		toast.WithWidth(20),
		toast.WithMaxHeight(0),
		toast.WithGap(0),
	)
	model, _, _ = model.Push(toast.NewToast("visible", toast.WithID("visible")))
	model, _, _ = model.Push(toast.NewToast("first queued", toast.WithID("first")))
	model, _, _ = model.Push(toast.NewToast("second queued", toast.WithID("second")))

	if view := model.View(); !strings.Contains(view, "+2 more") {
		t.Fatalf("expected queued Toast count in rendered stack, got %q", view)
	}

	model, _ = model.Dismiss("visible")
	if view := model.View(); !strings.Contains(view, "+1 more") {
		t.Fatalf("expected queued Toast count to update after drain, got %q", view)
	}

	model, _ = model.Dismiss("first")
	if view := model.View(); strings.Contains(view, "more") {
		t.Fatalf("expected queue indicator to disappear when queue drains, got %q", view)
	}
}

func TestQueueIndicatorPlacementCustomizationAndDisable(t *testing.T) {
	pushQueued := func(model toast.Model) toast.Model {
		model, _, _ = model.Push(toast.NewToast("visible", toast.WithID("visible")))
		model, _, _ = model.Push(toast.NewToast("queued", toast.WithID("queued")))
		return model
	}

	top := pushQueued(toast.New(
		toast.WithPlacement(toast.TopRight),
		toast.WithMaxVisible(1),
		toast.WithStyle(toast.KindNone, lipgloss.NewStyle()),
		toast.WithWidth(20),
		toast.WithMaxHeight(0),
		toast.WithGap(0),
	))
	if lines := strings.Split(top.View(), "\n"); lines[len(lines)-1] != "+1 more" {
		t.Fatalf("top placement should render queue indicator below Toast Stack, got %q", top.View())
	}

	bottom := pushQueued(toast.New(
		toast.WithPlacement(toast.BottomRight),
		toast.WithMaxVisible(1),
		toast.WithStyle(toast.KindNone, lipgloss.NewStyle()),
		toast.WithWidth(20),
		toast.WithMaxHeight(0),
		toast.WithGap(0),
	))
	if lines := strings.Split(bottom.View(), "\n"); lines[0] != "+1 more" {
		t.Fatalf("bottom placement should render queue indicator above Toast Stack, got %q", bottom.View())
	}

	custom := pushQueued(toast.New(
		toast.WithMaxVisible(1),
		toast.WithQueueIndicator(func(ctx toast.QueueIndicatorContext) string {
			return "waiting: " + strconv.Itoa(ctx.Count)
		}),
	))
	if view := custom.View(); !strings.Contains(view, "waiting: 1") || strings.Contains(view, "+1 more") {
		t.Fatalf("expected custom queue indicator, got %q", view)
	}

	disabled := pushQueued(toast.New(toast.WithMaxVisible(1), toast.WithoutQueueIndicator()))
	if view := disabled.View(); strings.Contains(view, "more") {
		t.Fatalf("expected disabled queue indicator to be omitted, got %q", view)
	}
}

func TestPublicAPIErgonomicsForBubbleTeaApps(t *testing.T) {
	model := toast.New()
	msg := toast.Show(toast.Success("saved", toast.WithTitle("Done"), toast.WithID("save-status")))()

	var cmd tea.Cmd
	model, cmd = model.Update(msg)

	if cmd == nil {
		t.Fatal("message-based Show should schedule visible Toast Dismissal")
	}
	if model.Len() != 1 || model.Visible()[0].ID != "save-status" {
		t.Fatalf("Toast model did not route public Show message: %#v", model.Visible())
	}

	model, _ = model.Update(toast.Dismiss("save-status")())
	if model.Len() != 0 {
		t.Fatalf("public Dismiss command did not remove Toast, len=%d", model.Len())
	}
}
