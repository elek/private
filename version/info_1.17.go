//go:build !go1.18

// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.
package version

// GetInfo returns build information based on build time tags AND go compiler vcs information.
func GetInfo() Info {
	return Build
}
