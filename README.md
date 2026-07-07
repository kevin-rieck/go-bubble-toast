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

## Release

See [CHANGELOG.md](CHANGELOG.md) for release notes.
