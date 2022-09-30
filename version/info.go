//go:build go1.18

// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.
package version

import (
	"runtime/debug"
	"strconv"
	"time"
)

// GetInfo returns build information based on build time tags AND go compiler vcs information.
func GetInfo() Info {
	i := Info{
		commitHashCRC: Build.commitHashCRC,
		Timestamp:     Build.Timestamp,
		CommitHash:    Build.CommitHash,
		Version:       Build.Version,
		Release:       Build.Release,
		Modified:      Build.Modified,
	}

	undefinedTime := time.Time{}
	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				if i.CommitHash == "" {
					i.CommitHash = s.Value
				}
			case "vcs.time":
				if i.Timestamp == undefinedTime {
					i.Timestamp, _ = time.Parse(time.RFC3339Nano, s.Value)
				}
			case "vcs.modified":
				modified, err := strconv.ParseBool(s.Value)
				if err == nil {
					i.Modified = i.Modified || modified
				}
			}
		}

	}
	return i
}
