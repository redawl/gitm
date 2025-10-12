package util

import (
	"log/slog"
	"os"
	"runtime/pprof"
)

// ProfileGitm starts cpu profiling, saving the profile to /tmp/gitm.pprof
func ProfileGitm() func() {
	slog.Error("PROFILING IS ENABLED! THIS SHOULD NEVER BE ENABLED IN PRODUCTION")
	f, err := os.Create("/tmp/gitm.pprof")
	if err != nil {
		slog.Error("Error opening profile file for writing", "error", err)
		return func() {}
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		slog.Error("Error when starting profile", "error", err)
		return func() {}
	}
	return pprof.StopCPUProfile
}
