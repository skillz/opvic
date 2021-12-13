package utils

import (
	"fmt"
	"runtime"
)

var (
	Version   = "development"
	Revision  string
	Branch    string
	BuildUser string
	BuildDate string
	GoVersion = runtime.Version()
)

// Info returns version, branch and revision information.
func VersionInfo() string {
	return fmt.Sprintf("version=%s, branch=%s, revision=%s", Version, Branch, Revision)
}

// BuildContext returns goVersion, buildUser and buildDate information.
func BuildContext() string {
	return fmt.Sprintf("go=%s, user=%s, date=%s", GoVersion, BuildUser, BuildDate)
}
