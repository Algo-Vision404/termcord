# Termcord vision

Termcord 0.1.0 is the **first release** — a personal terminal client where your account, your cache, and your machine stay in sync. Not a relay wrapper.

## Design principles

1. **Direct** — one binary talks to Discord. No operator in the middle.
2. **Local-first** — history, search, and optional encryption live on your disk.
3. **Keyboard-native** — spotlight jump, focus mode, and slash commands beat mouse-driven UIs.
4. **Honest** — user-token clients violate Discord ToS; we say so and never pretend otherwise.
5. **Extensible** — plugins and `termcord-cli` for automation without forking the app.

## What shipped in 0.1.0

- Direct user gateway with reconnect and `termcord doctor`
- Clean TUI: merged status bar, stream-style chat, single shortcut footer
- Spotlight, focus mode, mention hop, tab completion
- Guild sidebar, threads, replies, reactions, embeds
- Local SQLite cache with `/search` and optional encryption
- Kitty/iTerm2 image preview, themes, plugins, headless CLI

## After 0.1.0

- Richer markdown rendering
- AI-native helpers (`/summarize`, local LLM hooks) built on the cache
- Release channels and signed binaries (Scoop, Homebrew, GitHub Releases)
