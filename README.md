# termcord

**termcord** is a terminal-native Discord client — direct gateway, local history, no relay.

> Using a Discord user token violates [Discord's Terms of Service](https://discord.com/terms). Use at your own risk.

## Features (v0.1)

- Full account access — all guilds, DMs, threads (no bot install)
- Guild-grouped sidebar with unread + @mention badges
- Message replies, reactions, embed rendering, markdown cleanup
- Local SQLite cache with `/search` and `/history`
- Headless CLI for scripts and CI
- Themes: `default`, `dracula`
- Encrypted local cache (AES-256-GCM, key in OS keyring)
- Plugin commands via TOML in plugins directory
- Gateway reconnect status notifications
- Token stored in OS keyring — never written to config by default

## Install

Requires Go 1.22+.

```bash
git clone <your-repo>/termcord.git
cd termcord
make build
```

Or on Windows:

```powershell
.\scripts\install.ps1
```

Scoop (after publishing releases):

```powershell
scoop bucket add termcord .\scoop
scoop install termcord
```

Binaries land in `bin/termcord` and `bin/termcord-cli`.

## Quick start

```powershell
# 1. Create config
.\bin\termcord.exe init

# 2. Save your Discord user token (prompts if omitted)
.\bin\termcord.exe login

# 3. Verify connection
.\bin\termcord.exe doctor

# 4. Launch
.\bin\termcord.exe
```

Alternative: `$env:TERMCORD_TOKEN = "..."` without keyring.

## Commands (TUI)

| Key | Action |
|-----|--------|
| Enter | Send / slash command |
| Ctrl+B | Toggle sidebar |
| Ctrl+P / Ctrl+N | Previous / next channel |
| Ctrl+T | Threads panel |
| Ctrl+R | Quick react (1–5) |
| Ctrl+↑ / Ctrl+↓ | Select message |
| PgUp / PgDn | Scroll chat (empty input) |
| Ctrl+L | Jump to bottom |
| Ctrl+C | Quit |

| Slash | Action |
|-------|--------|
| `/help` | Show commands in chat |
| `/join #name` | Switch channel |
| `/history` | Load older messages |
| `/reply` | Reply to selected message |
| `/threads` `/thread` `/parent` | Thread navigation |
| `/react 👍` | Toggle reaction |
| `/search query` | Search local cache |
| `/clear` `/quit` | Clear view / exit |

## Headless CLI

```bash
termcord-cli whoami
termcord-cli channels
termcord-cli history -channel CHANNEL_ID -limit 30
termcord-cli send -channel CHANNEL_ID "ship it"
termcord-cli send -channel CHANNEL_ID -reply MSG_ID "ack"
```

## Config

Path: `%APPDATA%\termcord\config.toml` (Windows) or `~/.config/termcord/config.toml`.

See [config.example.toml](config.example.toml). Notable options:

```toml
[ui]
theme = "dracula"
notify_on_mention = true
history_page_size = 50
```

## vs molly-terminal

| | molly | termcord |
|---|-------|----------|
| Connection | Third-party relay | Direct gateway |
| Servers | Bot required | All yours |
| DMs / threads | Limited | Yes |
| Identity | Webhook/bot | You |
| Privacy | Messages on relay | Local only |

## Project layout

```
cmd/termcord/          interactive client
cmd/termcord-cli/      headless commands
internal/gateway/      Discord connection
internal/cache/        SQLite message store
internal/tui/          Bubble Tea UI
internal/text/         Markdown + embed formatting
```

## License

MIT — see [LICENSE](LICENSE).
