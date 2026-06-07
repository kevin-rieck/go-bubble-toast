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

## Release

See [CHANGELOG.md](CHANGELOG.md) for release notes.
