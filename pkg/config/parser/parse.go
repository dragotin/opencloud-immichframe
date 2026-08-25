// SPDX-FileCopyrightText: 2026 Klaas Freitag <opensource@freisturz.de>
//
// SPDX-License-Identifier: GPL-3.0-or-later

// Package parser loads configuration from the known sources (a yaml config
// file, then environment variables) and validates it.
package parser

import (
	"errors"
	"fmt"

	occfg "github.com/opencloud-eu/opencloud/pkg/config"
	"github.com/opencloud-eu/opencloud/pkg/config/envdecode"

	"github.com/immichFrame/opencloud-immichframe/pkg/config"
	"github.com/immichFrame/opencloud-immichframe/pkg/config/defaults"
)

// ParseConfig applies the config sources to cfg and validates it.
// Precedence: defaults -> yaml file -> env vars. yaml via opencloud's
// BindSourcesToStructs, env via envdecode. Importing opencloud/pkg/config
// pulls in its service configs; opencloud-eu/opencloud#3270 moves
// BindSourcesToStructs to a leaf package to drop those.
func ParseConfig(cfg *config.Config) error {
	if err := occfg.BindSourcesToStructs(cfg.Service.Name, cfg); err != nil {
		return err
	}

	defaults.EnsureDefaults(cfg)

	if err := envdecode.Decode(cfg); err != nil {
		// No env vars set at all is fine; anything else is a real error.
		if !errors.Is(err, envdecode.ErrNoTargetFieldsAreSet) {
			return err
		}
	}

	defaults.Sanitize(cfg)

	return Validate(cfg)
}

// Validate checks mandatory fields.
func Validate(cfg *config.Config) error {
	if cfg.OpenCloud.URL == "" {
		return errors.New("immichframe: OC_URL (or IMMICHFRAME_OPENCLOUD_URL) is required")
	}
	if cfg.OpenCloud.SpaceID == "" && cfg.OpenCloud.SpaceName == "" {
		return errors.New("immichframe: either IMMICHFRAME_OPENCLOUD_SPACE_ID or IMMICHFRAME_OPENCLOUD_SPACE_NAME is required")
	}
	if cfg.OpenCloud.BearerToken == "" && (cfg.OpenCloud.Username == "" || cfg.OpenCloud.AppPassword == "") {
		return fmt.Errorf("immichframe: provide IMMICHFRAME_OPENCLOUD_BEARER_TOKEN, or IMMICHFRAME_OPENCLOUD_USERNAME + IMMICHFRAME_OPENCLOUD_APP_PASSWORD")
	}
	return nil
}
