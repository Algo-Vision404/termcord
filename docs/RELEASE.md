# Release process

termcord follows [Semantic Versioning](https://semver.org/) starting at **0.1.0**.

## Version scheme

| Bump   | Example   | When |
|--------|-----------|------|
| Patch  | 0.1.0 → 0.1.1 | Security fixes, bug fixes, dependency CVE patches |
| Minor  | 0.1.x → 0.2.0 | New features, backward-compatible CLI/config changes |
| Major  | 0.x → 1.0.0   | Stable API contract (future) |

While in **0.x**, minor releases may include small breaking changes; patch releases stay compatible.

## Before a release

1. Update `internal/version/version.go` (`Version` constant).
2. Sync version strings in `README.md`, `config.example.toml`, packaging manifests (Scoop, Homebrew, `.goreleaser.yml`).
3. Run quality gates:
   ```bash
   make test
   make security
   go build ./cmd/termcord
   ```
4. Review `SECURITY.md` supported versions table if the release series changes.

## Tagging

```bash
git tag -a v0.1.1 -m "v0.1.1 — security and bug fixes"
git push origin v0.1.1
```

Tags must match `Version` with a `v` prefix. GoReleaser uses `.goreleaser.yml`.

## Security releases

- Target the latest **0.1.x** patch.
- Document fixes in release notes without exploit details.
- Bump patch version even for dependency-only fixes when they address known vulnerabilities.

## Supported series

See `internal/version/policy.go` — currently **0.1.x** receives security patches.
