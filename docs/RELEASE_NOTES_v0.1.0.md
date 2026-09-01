# termcord v0.1.0

First public release — a terminal-native Discord client with direct gateway connection, local SQLite cache, and a keyboard-first Bubble Tea TUI.

## Highlights

### Core client
- Direct Discord gateway connection (no third-party relay)
- Guild-grouped sidebar with unread and @mention badges
- Spotlight (`Ctrl+G`) — fuzzy jump to channels, DMs, and servers
- Focus mode (`Ctrl+F`) — hide chrome for reading
- Mention hop (`Ctrl+M`) — cycle channels that @mention you
- Replies, reactions, embeds, threads
- Local SQLite cache with `/search`, `/history`, optional AES-256-GCM encryption
- Headless `termcord-cli` for scripts and automation
- Themes: `default`, `dracula`
- Plugin hooks via TOML command definitions

### CLI & onboarding
- `termcord init`, `login`, `logout`, `doctor`, `version`
- OS keyring token storage (Windows Credential Manager, macOS Keychain, Linux Secret Service)
- CLI boot splash with Big figlet wordmark
- Calm defaults: `reduce_motion = true`, Cordy mascot off

### Design system (`internal/ds/`)
- Shared tokens, theme, chrome, layout, and CLI presentation layer
- Stream layout: merged app bar, borderless message stream, flat embeds
- Underlined sidebar sections (no per-panel boxes)

### Gateway reliability
- Correct identify payload (`capabilities: 16381`, full client properties)
- Connection retry on open (3 attempts) with clearer close-code errors
- Channel name resolution (snowflake IDs → readable names in UI/DMs)

## Security (0.1.0)

- **Plugins off by default** — opt in with `plugins.enabled = true`
- **Plugin env stripping** — `TERMCORD_TOKEN`, `TERMCORD_CONFIG`, etc. never passed to plugin subprocesses
- **Cache permissions** — owner-only dir/file modes on Unix (`0700` / `0600`)
- **Login warnings** — CLI token args warn about process-list exposure; interactive/`--file` recommended
- **`termcord logout --purge-cache`** — remove token, delete cache, clear encryption key
- **Startup/doctor warnings** — surface plaintext config tokens, env tokens, disabled encryption
- **CI** — GitHub Actions with `go test` and `govulncheck`
- **`SECURITY.md`** — supported versions (0.1.x), disclosure policy, plugin threat model
- **`.gitignore`** — excludes `token.txt`, `*.token`, local `config.toml`, database files

## Versioning policy

| Bump | Example | When |
|------|---------|------|
| Patch | 0.1.0 → 0.1.1 | Security fixes, bug fixes, dependency CVE patches |
| Minor | 0.1.x → 0.2.0 | New features, backward-compatible changes |
| Major | 0.x → 1.0.0 | Stable API contract (future) |

See [docs/RELEASE.md](docs/RELEASE.md) and [SECURITY.md](SECURITY.md).

## Install

```bash
git clone https://github.com/termcord/termcord.git
cd termcord
make build
```

Windows: `.\scripts\install.ps1`

## Quick start

```bash
termcord init
termcord login
termcord doctor
termcord
```

## Warning

Using a Discord **user token** violates Discord's Terms of Service. Use at your own risk.
