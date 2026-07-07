# Bubble Toast

Bubble Toast is a Bubble Tea/Lip Gloss component for transient, non-blocking Toasts in terminal UIs.

## Demo

![Bubble Toast feature showcase](assets/showcase.gif)

### More examples

| Startup Toast | Toast stack | Progress bar |
| --- | --- | --- |
| ![Startup Toast](assets/startup.gif) | ![Multiple Toasts](assets/burst.gif) | ![Progress bar Toasts](assets/progress.gif) |

These GIFs were created with [VHS](https://github.com/charmbracelet/vhs).

## Run the example

```sh
go run ./examples/basic
```

Press `t` to show Toasts and `q` to quit.

## Basic usage

```go
type model struct {
    toasts toast.Model
}

func (m model) Init() tea.Cmd {
    return toast.Show(toast.Info("ready", toast.WithTitle("App")))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.toasts, cmd = m.toasts.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return m.toasts.Overlay("your app view")
}
```

Small apps can also push directly:

```go
m := toast.New()
m, id, cmd := m.Push(toast.Success("saved"))
_, _ = id, cmd
```

## Updating a long-running status Toast

Use a stable Toast ID when a status may change over time. `Replace` reuses the
same Toast ID, so the Toast Stack or queue keeps one entry for that status.

```go
const syncStatus = "sync-status"

cmd := toast.Replace(syncStatus, toast.Info("syncing", toast.WithPersistent()))

// Later, when the work finishes:
cmd = toast.Replace(syncStatus, toast.Success("synced"))
```

Direct model updates use the same replacement behavior:

```go
m, cmd = m.Replace(syncStatus, toast.Error("sync failed"))
```

## Toast Kind icons

Enable Toast Kind icons when color alone should not communicate Toast intent:

```go
m := toast.New(toast.WithKindIcons())
```

Built-in icons are available for info, success, warning, and error Toasts. Apps
can override icons or disable them globally for terminals and fonts where icons
are not appropriate:

```go
m := toast.New(
    toast.WithKindIcons(),
    toast.WithIcon(toast.KindSuccess, "OK"),
)

// Or disable icons globally when terminal/font support is unsuitable.
m := toast.New(toast.WithoutIcons())
```

Icons are added to Toasts rendered from title/message fields. `WithContent`
remains pre-rendered content and is not modified.

## Compatibility rendering

For constrained terminals, enable compatibility modes for built-in Toast
presentation:

```go
m := toast.New(
    toast.WithASCIIOnly(), // ASCII borders and built-in icons
    toast.WithNoColor(),   // no ANSI color sequences from built-in styles
)
```

Host-provided icon overrides and custom renderers remain under the host app's
control.

## Renderer presets

Built-in renderer presets provide common Toast presentations without replacing
the custom renderer escape hatch:

```go
m := toast.New(toast.WithRendererPreset(toast.PresetCompact))
```

Use `PresetCompact` for dense bordered Toasts, `PresetMinimal` for unboxed Toast
Content, and `PresetIcon` when Toast Kind should be led by an icon. A custom
renderer configured with `WithRenderer` still defines the full Toast
presentation and takes precedence over presets.

## Queue behavior

When the visible Toast Stack is full, Bubble Toast queues overflow Toasts. The
default full-queue behavior preserves previous releases by dropping the oldest
queued Toast when a new Toast arrives. Apps can instead drop the newest incoming
Toast:

```go
m := toast.New(toast.WithQueueOverflowPolicy(toast.DropNewestToast))
```

Matching Toast ID updates still replace the existing visible or queued Toast and
do not consume queue capacity.

Apps that emit repeated generated-ID Toasts can opt into duplicate coalescing.
When enabled, generated-ID Toasts with the same Toast Kind and message merge into
one visible or queued Toast and render an occurrence count:

```go
m := toast.New(toast.WithDuplicateCoalescing())
```

Explicit Toast IDs remain distinct so host apps keep control over configured
Toast identity.

Bubble Toast also renders a `+N more` indicator for queued Toasts. Apps can
customize or disable that stack-level affordance:

```go
m := toast.New(
    toast.WithQueueIndicator(func(ctx toast.QueueIndicatorContext) string {
        return fmt.Sprintf("waiting: %d", ctx.Count)
    }),
)

m = toast.New(toast.WithoutQueueIndicator())
```

## Release

See [CHANGELOG.md](CHANGELOG.md) for release notes.
