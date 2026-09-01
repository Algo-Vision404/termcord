# Security

## Token handling

- Prefer `termcord login` — stores token in the OS keyring (Windows Credential Manager, macOS Keychain, Linux Secret Service).
- `TERMCORD_TOKEN` env var overrides keyring for the current process only.
- Avoid putting tokens in `config.toml`. The example file leaves `token = ""` intentionally.

## Data storage

- Message cache is a local SQLite file (default `%LOCALAPPDATA%\termcord\messages.db`).
- Run `/clear` to clear the view; delete the `.db` file to wipe cached history.
- No telemetry, analytics, or third-party relay.

## Discord ToS

termcord connects to Discord's gateway with a **user token**. This is self-bot behavior and violates Discord's Terms of Service. Accounts may be suspended. You are responsible for how you use this software.

## Reporting issues

Do not open public issues containing tokens, message content, or user IDs.
