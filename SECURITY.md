# Security

## Supported versions

Security fixes are released for the current **0.1.x** series only. See [docs/RELEASE.md](docs/RELEASE.md) for versioning.

| Version | Supported |
|---------|-----------|
| 0.1.x   | Yes       |
| < 0.1.0 | No        |

## Token handling

**Priority order:** `TERMCORD_TOKEN` env var → `config.toml` `[general].token` → OS keyring.

- **Prefer** `termcord login` (interactive) — stores token in the OS keyring (Windows Credential Manager, macOS Keychain, Linux Secret Service).
- **Avoid** passing tokens on the command line — they may appear in process listings (`ps`, Task Manager).
- **Avoid** plaintext tokens in `config.toml`. Use `token = ""` and run `termcord login`.
- `TERMCORD_TOKEN` is convenient for dev but is inherited by child processes unless stripped.

Run `termcord doctor` to see token source and security warnings.

## Local data

- Message cache: local SQLite file (default `%LOCALAPPDATA%\termcord\messages.db` on Windows).
- With `cache.encrypt = true` (default), message bodies are encrypted at rest (AES-256-GCM). The key lives in the OS keyring.
- Cache directory and database file use owner-only permissions on Unix (`0700` / `0600`).
- `termcord logout --purge-cache` removes the token, deletes the cache file, and clears the encryption key.
- No telemetry, analytics, or third-party relay.

## Plugins

Plugins are **disabled by default** (`plugins.enabled = false`). When enabled, each plugin runs a shell command with:

- `TERMCORD_CHANNEL_ID`, `TERMCORD_CHANNEL`, `TERMCORD_USER`, `TERMCORD_MESSAGE`, `TERMCORD_ARGS`

Secrets are **not** passed to plugins (`TERMCORD_TOKEN`, `TERMCORD_CONFIG`, etc. are stripped). Only install plugin `.toml` files you trust — they run arbitrary commands as your user.

## Discord ToS

termcord connects to Discord's gateway with a **user token**. This is self-bot behavior and violates Discord's Terms of Service. Accounts may be suspended. You are responsible for how you use this software.

## Reporting vulnerabilities

**Do not** open public GitHub issues for security vulnerabilities.

1. Do not include tokens, message content, or user IDs in reports.
2. Email the maintainer privately (or use GitHub Security Advisories if enabled).
3. Include termcord version (`termcord version`) and platform.
4. We aim to acknowledge within 72 hours and ship a patch in the next **0.1.x** release when applicable.

## Dependency updates

Run `make security` (or CI) to scan with [govulncheck](https://go.dev/security/vuln/). Patch releases may bump dependencies for known CVEs.
