// SPDX-FileCopyrightText: 2026 Klaas Freitag <opensource@freisturz.de>
//
// SPDX-License-Identifier: GPL-3.0-or-later

// Package defaults provides the programmatic configuration defaults. Keeping
// defaults here (rather than in `default=` struct tags) matches the OpenCloud
// convention and lets the docs generator render them from the DefaultConfig
// literal.
package defaults

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/immichFrame/opencloud-immichframe/pkg/config"
)

// FullDefaultConfig returns the fully defaulted config.
func FullDefaultConfig() *config.Config {
	cfg := DefaultConfig()
	EnsureDefaults(cfg)
	Sanitize(cfg)
	return cfg
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *config.Config {
	return &config.Config{
		Service: config.Service{
			Name: "immichframe",
		},
		HTTP: config.HTTP{
			Addr: ":8080",
		},
		ImmichFrame: config.ImmichFrame{
			CatalogRefresh: 5 * time.Minute,
		},
		Client: config.ClientSettings{
			Interval:            8,
			TransitionDuration:  1,
			ShowClock:           true,
			ClockFormat:         "HH:mm",
			ClockDateFormat:     "eee, MMM d",
			ShowProgressBar:     true,
			ShowPhotoDate:       true,
			PhotoDateFormat:     "2006-01-02",
			ShowImageLocation:   false,
			ImageLocationFormat: "City,State,Country",
			ShowImageDesc:       false,
			ShowAlbumName:       true,
			Language:            "en",
		},
	}
}

// EnsureDefaults fills in any dependent defaults. Runs BEFORE env-var binding,
// so only use it for fields that don't depend on other user-configurable
// values. Currently a no-op.
func EnsureDefaults(_ *config.Config) {}

// Sanitize normalises the config (trims trailing slashes on URLs, resolves the
// web root to an absolute path). WebRoot must come out absolute and cleaned:
// the static handler compares joined request paths against it with a prefix
// check, which a relative path or an unresolved ".." segment would defeat.
func Sanitize(cfg *config.Config) {
	cfg.OpenCloud.URL = strings.TrimRight(cfg.OpenCloud.URL, "/")

	if cfg.ImmichFrame.WebRoot != "" {
		// filepath.Abs cleans as it resolves; keep the original on failure so
		// the operator sees the path they configured in the error.
		if abs, err := filepath.Abs(cfg.ImmichFrame.WebRoot); err == nil {
			cfg.ImmichFrame.WebRoot = abs
		}
		cfg.ImmichFrame.WebRoot = strings.TrimRight(cfg.ImmichFrame.WebRoot, "/")
	}
}
