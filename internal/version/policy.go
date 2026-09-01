package version

// Semver policy for termcord releases.
//
//   - 0.1.x — patch: security fixes, bug fixes, dependency updates
//   - 0.x.0 — minor: new features, backward-compatible config/CLI
//   - 1.0.0 — stable API contract (future)
//
// Tag releases as vMAJOR.MINOR.PATCH matching Version.
const (
	ReleaseSeries = "0.1"
)

// Supported returns supported release series for security patches.
func Supported() []string {
	return []string{ReleaseSeries}
}
