# Termcord design system (v0.1.0)

Termcord separates **what things say** from **how they look** and **where they live on screen**.

## TUI chrome (0.1.0)

- **App bar** — merged status + channel (4 lines, horizontal rules only)
- **Message stream** — no per-line borders; blank line between messages
- **Compose** — single input row with rule above
- **Footer** — one dim hint line (hidden in focus mode)
- **Sidebar** — underlined section titles, no nested boxes

## Layers

```
internal/ds/     tokens, boxes, themes, CLI chrome  ← single source of truth
internal/art/    ASCII scenes, mascot, calm vs animated content
internal/tui/    Bubble Tea layout and interaction
internal/ux/     friendly copy and error strings
```

### `internal/ds` — design system

Shared primitives for TUI and CLI:

| File | Role |
|------|------|
| `tokens.go` | Brand name, tagline, standard widths |
| `box.go` | `Top`, `Row`, `Bottom`, `Build`, `BuildDynamic`, sidebar sections |
| `theme.go` | `Theme` palette + `LoadTheme(name)` |
| `style.go` | `ApplyBorder`, `ApplyFooter`, `ApplyChrome3`, `ApplyCordyLane` |
| `chrome.go` | App bar, compose row, footer, stream layout |
| `wordmark.go` | Boot ASCII wordmark |
| `ascii.go` | `BootSplash`, `CLILogo` |
| `help.go` | `/help` block framing |
| `cli.go` | Spinner + `RunWithSpinner` for headless commands |

**Rule:** any new box, border, or theme color goes here first.

### `internal/art` — content scenes

ASCII illustrations and motion. Uses `ds` for boxes but stays free of layout decisions:

- Connecting / loading / empty / welcome / farewell screens
- Cordy mascot sprites and lane content
- Scene dispatch (`RenderConnecting`, calm vs animated)

Art should not duplicate box drawing — import `ds` or `art/box.go` wrappers.

### `internal/tui` — application shell

Bubble Tea model: channels, messages, input, spotlight. Imports `ds.Theme` (via type alias) and applies styling through `ds.Apply*`.

### `internal/ux` — copy

Human-readable errors and hints. No lipgloss, no boxes.

## Width tokens

| Token | Value | Use |
|-------|-------|-----|
| `PanelInner(w)` | terminal − 2 | Header, channel strip, compose, footer |
| `ChatInner(w, sidebar, open)` | terminal − sidebar − 2 | Message panel |
| `StandardInner` | 58 | Fallback when size unknown |
| `ClampInner(w)` | 40–78 | CLI boxes and fixed-width art |

## Themes

Configured in `config.toml` as `ui.theme`:

- `default` — warm pink accent
- `dracula` — purple / green palette

Unknown names fall back to `default` with a status note.

## Adding UI

1. **New framed panel** → `ds.Build(inner, title, rows)` + `ds.ApplyBorder(theme, raw)`
2. **New accent row in a 3-line strip** → `ds.ApplyChrome3(theme, raw, theme.Status)`
3. **New splash screen** → plain text in `art/`, boxes via `ds`
4. **New CLI command output** → `ds.BuildDynamic` or `ds.CLILogo`
5. **New user-facing error** → `ux` package only

## Calm defaults

`reduce_motion = true` and `show_mascot = false` by default. Animated paths remain in `art` but design tokens and chrome stay stable either way.
