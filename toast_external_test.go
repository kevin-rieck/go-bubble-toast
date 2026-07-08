package toast_test

import (
	"strconv"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	toast "github.com/kevin-rieck/go-bubble-toast"
	"github.com/muesli/termenv"
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

func TestAppCanCheckVisibleAndQueuedToastsByID(t *testing.T) {
	model := toast.New(toast.WithMaxVisible(1))
	model, _, _ = model.Push(toast.NewToast("visible", toast.WithID("visible")))
	model, _, _ = model.Push(toast.NewToast("queued", toast.WithID("queued")))

	if !model.HasVisible("visible") || model.HasVisible("queued") {
		t.Fatalf("HasVisible should only match visible Toast IDs")
	}
	if !model.HasQueued("queued") || model.HasQueued("visible") {
		t.Fatalf("HasQueued should only match queued Toast IDs")
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

func TestAppCanQueryAnyToastByID(t *testing.T) {
	model := toast.New(toast.WithMaxVisible(1))
	model, _, _ = model.Push(toast.NewToast("visible", toast.WithID("visible")))
	model, _, _ = model.Push(toast.NewToast("queued", toast.WithID("queued")))

	visible, ok := model.Get("visible")
	if !ok || visible.ID != "visible" || visible.Message != "visible" {
		t.Fatalf("Get should find visible Toasts by Toast ID, got %#v ok=%v", visible, ok)
	}
	queued, ok := model.Get("queued")
	if !ok || queued.ID != "queued" || queued.Message != "queued" {
		t.Fatalf("Get should find queued Toasts by Toast ID, got %#v ok=%v", queued, ok)
	}
	if _, ok := model.Get("missing"); ok {
		t.Fatal("Get should not find missing Toast IDs")
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

func TestHighPriorityToastBecomesVisibleWhenStackIsFull(t *testing.T) {
	model := toast.New(toast.WithMaxVisible(1), toast.WithMaxQueued(2))
	model, _, _ = model.Push(toast.NewToast("routine", toast.WithID("routine")))
	model, _, _ = model.Push(toast.NewToast("critical", toast.WithID("critical"), toast.WithPriority(10)))

	if !model.IsVisible("critical") {
		t.Fatalf("high-priority Toast should be visible promptly, visible=%#v queued=%#v", model.Visible(), model.Queued())
	}
	if !model.IsQueued("routine") {
		t.Fatalf("lower-priority visible Toast should move to queue, visible=%#v queued=%#v", model.Visible(), model.Queued())
	}
}

func TestKindPriorityMakesMatchingToastVisibleWhenStackIsFull(t *testing.T) {
	model := toast.New(toast.WithMaxVisible(1), toast.WithMaxQueued(2), toast.WithKindPriority(toast.KindError, 10))
	model, _, _ = model.Push(toast.Info("routine", toast.WithID("routine")))
	model, _, _ = model.Push(toast.Error("critical", toast.WithID("critical")))

	if !model.IsVisible("critical") || !model.IsQueued("routine") {
		t.Fatalf("kind priority should make matching Toast visible, visible=%#v queued=%#v", model.Visible(), model.Queued())
	}
}

func TestPriorityOverflowKeepsHighestPriorityQueuedToasts(t *testing.T) {
	model := toast.New(toast.WithMaxVisible(1), toast.WithMaxQueued(2))
	model, _, _ = model.Push(toast.NewToast("visible", toast.WithID("visible"), toast.WithPriority(20)))
	model, _, _ = model.Push(toast.NewToast("low", toast.WithID("low"), toast.WithPriority(1)))
	model, _, _ = model.Push(toast.NewToast("medium", toast.WithID("medium"), toast.WithPriority(5)))
	model, _, _ = model.Push(toast.NewToast("high", toast.WithID("high"), toast.WithPriority(10)))

	if model.IsQueued("low") || !model.IsQueued("medium") || !model.IsQueued("high") {
		t.Fatalf("priority overflow should keep highest-priority queued Toasts, queued=%#v", model.Queued())
	}
}

func TestLowerPriorityOverflowDoesNotDisplaceHigherPriorityQueue(t *testing.T) {
	model := toast.New(toast.WithMaxVisible(1), toast.WithMaxQueued(2))
	model, _, _ = model.Push(toast.NewToast("visible", toast.WithID("visible"), toast.WithPriority(20)))
	model, _, _ = model.Push(toast.NewToast("medium", toast.WithID("medium"), toast.WithPriority(5)))
	model, _, _ = model.Push(toast.NewToast("high", toast.WithID("high"), toast.WithPriority(10)))
	model, _, _ = model.Push(toast.NewToast("low", toast.WithID("low"), toast.WithPriority(1)))

	if model.IsQueued("low") || !model.IsQueued("medium") || !model.IsQueued("high") {
		t.Fatalf("lower-priority overflow should not displace higher-priority queue, queued=%#v", model.Queued())
	}
}

func TestPriorityQueueDrainsHighestPriorityFirst(t *testing.T) {
	model := toast.New(toast.WithMaxVisible(1), toast.WithMaxQueued(3))
	model, _, _ = model.Push(toast.NewToast("visible", toast.WithID("visible"), toast.WithPriority(20)))
	model, _, _ = model.Push(toast.NewToast("medium", toast.WithID("medium"), toast.WithPriority(5)))
	model, _, _ = model.Push(toast.NewToast("high", toast.WithID("high"), toast.WithPriority(10)))

	model, _ = model.Dismiss("visible")
	if !model.IsVisible("high") {
		t.Fatalf("priority queue should drain highest-priority Toast first, visible=%#v queued=%#v", model.Visible(), model.Queued())
	}
}

func TestQueueOverflowPolicyCanDropNewestIncomingToast(t *testing.T) {
	model := toast.New(
		toast.WithMaxVisible(1),
		toast.WithMaxQueued(2),
		toast.WithQueueOverflowPolicy(toast.DropNewestToast),
	)
	model, _, _ = model.Push(toast.NewToast("visible", toast.WithID("visible")))
	model, _, _ = model.Push(toast.NewToast("first queued", toast.WithID("first")))
	model, _, _ = model.Push(toast.NewToast("second queued", toast.WithID("second")))

	model, _, _ = model.Push(toast.NewToast("dropped", toast.WithID("dropped")))

	queued := model.Queued()
	got := string(queued[0].ID) + "," + string(queued[1].ID)
	if got != "first,second" {
		t.Fatalf("DropNewestToast should preserve existing queued Toasts when queue is full, got %s", got)
	}
	if model.IsQueued("dropped") {
		t.Fatal("new incoming Toast should not be queued when DropNewestToast overflows")
	}

	model, _, _ = model.Push(toast.NewToast("updated second queued", toast.WithID("second")))
	if got, _ := model.QueuedByID("second"); got.Message != "updated second queued" {
		t.Fatalf("matching Toast ID update should bypass full queue capacity, got %#v", got)
	}

	model, _ = model.Dismiss("visible")
	visible := model.Visible()
	if len(visible) != 1 || visible[0].ID != "first" {
		t.Fatalf("queue should drain in FIFO order after DropNewestToast overflow, visible=%#v", visible)
	}
}

func TestRendererPresetMinimalRendersToastContentWithoutBorder(t *testing.T) {
	model := toast.New(toast.WithRendererPreset(toast.PresetMinimal), toast.WithWidth(20))
	model, _, _ = model.Push(toast.Success("saved", toast.WithTitle("Done")))

	view := model.View()
	if !strings.Contains(view, "Done") || !strings.Contains(view, "saved") {
		t.Fatalf("minimal preset should render Toast Content, got %q", view)
	}
	for _, border := range []string{"╭", "╮", "╰", "╯", "─", "│"} {
		if strings.Contains(view, border) {
			t.Fatalf("minimal preset should avoid bordered presentation, got %q", view)
		}
	}
}

func TestRendererPresetCompactKeepsBorderWithLessPadding(t *testing.T) {
	model := toast.New(toast.WithRendererPreset(toast.PresetCompact), toast.WithWidth(12), toast.WithMaxHeight(0))
	model, _, _ = model.Push(toast.Success("saved"))

	view := model.View()
	if !strings.Contains(view, "│saved") {
		t.Fatalf("compact preset should reduce padding inside bordered Toast, got %q", view)
	}
	if !strings.Contains(view, "╭") || !strings.Contains(view, "╯") {
		t.Fatalf("compact preset should keep bordered presentation, got %q", view)
	}
}

func TestRendererPresetIconRendersKindAffordance(t *testing.T) {
	model := toast.New(
		toast.WithRendererPreset(toast.PresetIcon),
		toast.WithStyle(toast.KindError, lipgloss.NewStyle()),
	)
	model, _, _ = model.Push(toast.Error("failed"))

	view := model.View()
	if !strings.Contains(view, "✕ failed") {
		t.Fatalf("icon preset should render a Toast Kind affordance, got %q", view)
	}
}

func TestRendererPresetDoesNotOverrideCustomRenderer(t *testing.T) {
	model := toast.New(
		toast.WithRendererPreset(toast.PresetIcon),
		toast.WithRenderer(func(t toast.Toast, _ toast.RenderContext) string {
			return "custom:" + t.Message
		}),
	)
	model, _, _ = model.Push(toast.Error("failed"))

	view := model.View()
	if view != "custom:failed" {
		t.Fatalf("custom renderer should take precedence over renderer preset, got %q", view)
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

func TestNoAnimationModeDisablesLifetimeProgressRendering(t *testing.T) {
	model := toast.New(
		toast.WithProgress(true),
		toast.WithNoAnimation(),
		toast.WithRendererPreset(toast.PresetMinimal),
		toast.WithWidth(30),
	)
	model, _, _ = model.Push(toast.NewToast("static lifetime"))

	view := model.View()
	if strings.Contains(view, "─") {
		t.Fatalf("no-animation mode should suppress animated progress output, got %q", view)
	}
	if !strings.Contains(view, "static lifetime") {
		t.Fatalf("no-animation mode should preserve Toast Content, got %q", view)
	}
}

func TestNoAnimationModeWinsRegardlessOfOptionOrder(t *testing.T) {
	model := toast.New(
		toast.WithNoAnimation(),
		toast.WithProgress(true),
		toast.WithRendererPreset(toast.PresetMinimal),
		toast.WithWidth(30),
	)
	model, _, _ = model.Push(toast.NewToast("static lifetime"))

	if view := model.View(); strings.Contains(view, "─") {
		t.Fatalf("no-animation mode should win regardless of option order, got %q", view)
	}
}

func TestASCIIOnlyRenderingKeepsHostIconOverrides(t *testing.T) {
	model := toast.New(
		toast.WithKindIcons(),
		toast.WithIcon(toast.KindSuccess, "OK"),
		toast.WithASCIIOnly(),
		toast.WithStyle(toast.KindSuccess, lipgloss.NewStyle()),
	)
	model, _, _ = model.Push(toast.Success("saved"))

	view := model.View()
	if !strings.Contains(view, "OK saved") {
		t.Fatalf("expected host icon override to remain under host control, got %q", view)
	}
	if strings.Contains(view, "v saved") || strings.Contains(view, "✓ saved") {
		t.Fatalf("ASCII-only mode should not replace host icon override, got %q", view)
	}
}

func TestNoColorRenderingAvoidsANSIEscapeSequences(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(previousProfile) })

	model := toast.New(toast.WithNoColor(), toast.WithWidth(18))
	model, _, _ = model.Push(toast.Error("failed"))

	view := model.View()
	if strings.Contains(view, "\x1b[") {
		t.Fatalf("expected no-color built-in rendering to avoid ANSI escape sequences, got %q", view)
	}
	if !strings.Contains(view, "failed") {
		t.Fatalf("expected Toast message to remain rendered, got %q", view)
	}
}

func TestASCIIOnlyRenderingUsesASCIIAffordances(t *testing.T) {
	model := toast.New(
		toast.WithASCIIOnly(),
		toast.WithKindIcons(),
		toast.WithMaxVisible(4),
		toast.WithWidth(18),
		toast.WithMaxHeight(0),
	)
	model, _, _ = model.Push(toast.Info("info"))
	model, _, _ = model.Push(toast.Success("saved"))
	model, _, _ = model.Push(toast.Warning("careful"))
	model, _, _ = model.Push(toast.Error("failed"))

	view := model.View()
	for _, want := range []string{"i info", "v saved", "! careful", "x failed"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected ASCII-only Toast Stack to contain %q, got %q", want, view)
		}
	}
	for _, notWant := range []string{"ℹ", "✓", "⚠", "✕", "╭", "╮", "╰", "╯", "─", "│"} {
		if strings.Contains(view, notWant) {
			t.Fatalf("ASCII-only rendering should avoid %q, got %q", notWant, view)
		}
	}
}

func TestDuplicateToastCoalescingShowsOccurrenceCount(t *testing.T) {
	model := toast.New(
		toast.WithDuplicateCoalescing(),
		toast.WithStyle(toast.KindWarning, lipgloss.NewStyle()),
	)
	model, _, _ = model.Push(toast.Warning("connection failed"))
	model, _, _ = model.Push(toast.Warning("connection failed"))
	model, _, _ = model.Push(toast.Warning("connection failed"))

	if model.Len() != 1 {
		t.Fatalf("duplicate Toasts should coalesce into one Toast, len=%d", model.Len())
	}
	view := model.View()
	if !strings.Contains(view, "connection failed") || !strings.Contains(view, "(x3)") {
		t.Fatalf("coalesced Toast should render message and occurrence count, got %q", view)
	}
}

func TestDuplicateToastCoalescingMergesQueuedToasts(t *testing.T) {
	model := toast.New(
		toast.WithDuplicateCoalescing(),
		toast.WithMaxVisible(1),
		toast.WithStyle(toast.KindNone, lipgloss.NewStyle()),
		toast.WithWidth(30),
		toast.WithMaxHeight(0),
	)
	model, _, _ = model.Push(toast.NewToast("visible"))
	model, _, _ = model.Push(toast.NewToast("retrying"))
	model, _, _ = model.Push(toast.NewToast("retrying"))

	queued := model.Queued()
	if len(queued) != 1 {
		t.Fatalf("duplicate queued Toasts should coalesce into one queued Toast, queued=%#v", queued)
	}
	model, _ = model.Dismiss(string(model.Visible()[0].ID))
	if view := model.View(); !strings.Contains(view, "retrying") || !strings.Contains(view, "(x2)") {
		t.Fatalf("coalesced queued Toast should render occurrence count when visible, got %q", view)
	}
}

func TestDuplicateToastCoalescingKeepsDistinctKindsAndDefaultBehavior(t *testing.T) {
	plain := toast.New(toast.WithStyle(toast.KindNone, lipgloss.NewStyle()))
	plain, _, _ = plain.Push(toast.NewToast("same"))
	plain, _, _ = plain.Push(toast.NewToast("same"))
	if plain.Len() != 2 {
		t.Fatalf("duplicate Toasts should remain distinct unless coalescing is enabled, len=%d", plain.Len())
	}

	coalescing := toast.New(toast.WithDuplicateCoalescing(), toast.WithMaxVisible(2))
	coalescing, _, _ = coalescing.Push(toast.Warning("same"))
	coalescing, _, _ = coalescing.Push(toast.Error("same"))
	if coalescing.Len() != 2 {
		t.Fatalf("Toasts with different Toast Kinds should remain distinct, len=%d", coalescing.Len())
	}
}

func TestDuplicateToastCoalescingKeepsExplicitToastIDsDistinct(t *testing.T) {
	model := toast.New(toast.WithDuplicateCoalescing(), toast.WithMaxVisible(2))
	model, _, _ = model.Push(toast.Warning("same", toast.WithID("first")))
	model, _, _ = model.Push(toast.Warning("same", toast.WithID("second")))

	if model.Len() != 2 || !model.IsVisible("first") || !model.IsVisible("second") {
		t.Fatalf("explicit Toast IDs should keep matching messages distinct, visible=%#v", model.Visible())
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

func TestKeyboardActionToastRendersUserVisibleHint(t *testing.T) {
	model := toast.New(toast.WithStyle(toast.KindNone, lipgloss.NewStyle()))
	model, _, _ = model.Push(toast.NewToast("file deleted", toast.WithAction("u", "undo", nil)))

	view := model.View()
	if !strings.Contains(view, "file deleted") || !strings.Contains(view, "[u] undo") {
		t.Fatalf("action Toast should render user-visible key hint, got %q", view)
	}
}

func TestKeyboardActionToastTriggersCommandAndDismisses(t *testing.T) {
	type undoMsg struct{}
	model := toast.New()
	model, _, _ = model.Push(toast.NewToast("file deleted", toast.WithID("delete"), toast.WithAction("u", "undo", func() tea.Msg {
		return undoMsg{}
	})))

	var cmd tea.Cmd
	model, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	if cmd == nil {
		t.Fatal("matching Toast action key should return action command")
	}
	if _, ok := cmd().(undoMsg); !ok {
		t.Fatalf("matching Toast action key returned unexpected message %#v", cmd())
	}
	if model.IsVisible("delete") {
		t.Fatal("invoking a Toast action should dismiss that Toast")
	}
}

func TestToastWithoutActionsIgnoresKeyMessages(t *testing.T) {
	model := toast.New()
	model, _, _ = model.Push(toast.NewToast("plain", toast.WithID("plain")))

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	if cmd != nil {
		t.Fatal("Toast without actions should not return a command for key messages")
	}
	if !updated.IsVisible("plain") {
		t.Fatal("Toast without actions should remain visible after key messages")
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
