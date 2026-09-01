# termcord 0.1.0

**Discord in your terminal.** Direct gateway, local cache, keyboard-first.

> Using a Discord user token violates [Discord's Terms of Service](https://discord.com/terms). Use at your own risk — termcord stores messages locally and never uses a third-party relay.

## What's in 0.1.0

- **Spotlight (Ctrl+G)** — fuzzy jump to any channel, DM, or server
- **Focus mode (Ctrl+F)** — hide sidebar and footer for reading
- **Mention hop (Ctrl+M)** — cycle channels that @mention you
- **Guild sidebar** — unread and @mention badges
- **Replies, reactions, embeds** — read and send like the desktop app
- **Local SQLite cache** — `/search`, `/history`, optional AES-256 encryption
- **Doctor** — one command to verify token, REST, and gateway
- **Headless CLI** — `termcord-cli` for scripts and automation
- **Themes** — `default`, `dracula`
- **Plugins** — TOML command hooks in your plugins directory

Calm by default: `reduce_motion = true`, Cordy mascot off. Enable animations or Cordy in `config.toml` if you want them.

## Install

Requires Go 1.22+.

```bash
git clone <your-repo>/termcord.git
cd termcord
make build
```

Windows:

```powershell
.\scripts\install.ps1
```

Binaries: `bin/termcord.exe` (TUI) and `bin/termcord-cli.exe` (headless).

## Quick start

```powershell
termcord init
termcord login
termcord doctor
termcord
```

1. **`termcord init`** — create config at `%APPDATA%\termcord\config.toml`
2. **`termcord login`** — save your token to the OS keyring (hidden prompt)
3. **`termcord doctor`** — verify REST + gateway
4. **`termcord`** — open the chat UI

Session-only token (not saved):

```cmd
set TERMCORD_TOKEN=your_token_here
termcord doctor
```

## Using the TUI

The sidebar lists channels — navigate with **Ctrl+P / Ctrl+N**, **Ctrl+G** (spotlight), or **`/join #name`**.

| Key | Action |
|-----|--------|
| Enter | Send / slash command |
| Ctrl+G | Spotlight — fuzzy jump anywhere |
| Ctrl+F | Focus mode |
| Ctrl+M | Hop to next @mention |
| Ctrl+B | Toggle sidebar |
| Ctrl+P / Ctrl+N | Previous / next channel |
| Ctrl+R | Quick react (1–5) |
| Ctrl+L | Jump to bottom |
| /help | Full command list |

## Headless CLI

```bash
termcord-cli whoami
termcord-cli channels
termcord-cli history -channel CHANNEL_ID -limit 30
termcord-cli send -channel CHANNEL_ID "hello from termcord"
```

## Config

See [config.example.toml](config.example.toml). Path: `%APPDATA%\termcord\config.toml` (Windows) or `~/.config/termcord/config.toml`.

```toml
[ui]
theme = "default"
reduce_motion = true   # recommended
show_mascot = false    # Cordy lane — opt in
encrypt = true         # under [cache]
```

Product direction: [docs/VISION.md](docs/VISION.md). UI design: [docs/DESIGN.md](docs/DESIGN.md).

## vs relay-based terminal clients

| | Relay client | termcord |
|---|--------------|----------|
| Connection | Third-party server | Direct to Discord |
| Servers | Bot invite per guild | All your servers |
| DMs / threads | Often limited | Full user client |
| Message identity | Bot/webhook | You |
| Offline search | Varies | Encrypted local SQLite |

## License

MIT — see [LICENSE](LICENSE).
