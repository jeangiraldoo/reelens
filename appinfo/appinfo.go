package appinfo

import "runtime/debug"

// Name is the application's display and resource name.
const Name = "reelens"

var version = "dev"

// Version returns the application version: the value stamped at build time,
// or the module version embedded by `go install reelens@<version>` when no
// ldflags were set.
func Version() string {
	if version != "dev" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return version
	}

	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}

	return version
}
