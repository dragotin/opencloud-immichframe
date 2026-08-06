// SPDX-FileCopyrightText: 2026 Klaas Freitag <opensource@freisturz.de>
//
// SPDX-License-Identifier: GPL-3.0-or-later

// Package config loads the service configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds everything the service needs to run.
type Config struct {
	// OpenCloud backend
	OpenCloudBaseURL string
	SpaceID          string
	SpaceName        string // used to resolve SpaceID when it is empty
	Username         string
	AppPassword      string
	BearerToken      string // optional, takes precedence over basic auth
	InsecureTLS      bool

	// Frame API
	ListenAddr     string
	AuthSecret     string
	CatalogRefresh time.Duration
	WebRoot        string // optional: directory with the built ImmichFrame web UI to serve at /

	// Client settings surfaced through /api/Config
	Client ClientSettings
}

// ClientSettings mirrors the subset of ImmichFrame's ClientSettingsDto that we
// can populate. Values are passed straight through to the frame client.
type ClientSettings struct {
	Interval            int
	TransitionDuration  float64
	ShowClock           bool
	ClockFormat         string
	ClockDateFormat     string
	ShowProgressBar     bool
	ShowPhotoDate       bool
	PhotoDateFormat     string
	ShowImageLocation   bool
	ImageLocationFormat string
	ShowImageDesc       bool
	ShowAlbumName       bool
	Language            string
}

// Load reads the configuration from the environment, applying defaults and
// validating required fields.
func Load() (*Config, error) {
	c := &Config{
		OpenCloudBaseURL: strings.TrimRight(getenv("OPENCLOUD_BASE_URL", ""), "/"),
		SpaceID:          getenv("OPENCLOUD_SPACE_ID", ""),
		SpaceName:        getenv("OPENCLOUD_SPACE_NAME", ""),
		Username:         getenv("OPENCLOUD_USERNAME", ""),
		AppPassword:      getenv("OPENCLOUD_APP_PASSWORD", ""),
		BearerToken:      getenv("OPENCLOUD_BEARER_TOKEN", ""),
		InsecureTLS:      getbool("OPENCLOUD_INSECURE_TLS", false),

		ListenAddr:     getenv("LISTEN_ADDR", ":8080"),
		AuthSecret:     getenv("AUTH_SECRET", ""),
		CatalogRefresh: getduration("CATALOG_REFRESH", 5*time.Minute),
		WebRoot:        strings.TrimRight(getenv("WEB_ROOT", ""), "/"),

		Client: ClientSettings{
			Interval:            getint("FRAME_INTERVAL", 8),
			TransitionDuration:  getfloat("FRAME_TRANSITION_DURATION", 1),
			ShowClock:           getbool("FRAME_SHOW_CLOCK", true),
			ClockFormat:         getenv("FRAME_CLOCK_FORMAT", "HH:mm"),
			ClockDateFormat:     getenv("FRAME_CLOCK_DATE_FORMAT", "eee, MMM d"),
			ShowProgressBar:     getbool("FRAME_SHOW_PROGRESS_BAR", true),
			ShowPhotoDate:       getbool("FRAME_SHOW_PHOTO_DATE", true),
			PhotoDateFormat:     getenv("FRAME_PHOTO_DATE_FORMAT", "2006-01-02"),
			ShowImageLocation:   getbool("FRAME_SHOW_IMAGE_LOCATION", false),
			ImageLocationFormat: getenv("FRAME_IMAGE_LOCATION_FORMAT", "City,State,Country"),
			ShowImageDesc:       getbool("FRAME_SHOW_IMAGE_DESC", false),
			ShowAlbumName:       getbool("FRAME_SHOW_ALBUM_NAME", true),
			Language:            getenv("FRAME_LANGUAGE", "en"),
		},
	}

	if c.OpenCloudBaseURL == "" {
		return nil, fmt.Errorf("OPENCLOUD_BASE_URL is required")
	}
	if c.SpaceID == "" && c.SpaceName == "" {
		return nil, fmt.Errorf("either OPENCLOUD_SPACE_ID or OPENCLOUD_SPACE_NAME is required")
	}
	if c.BearerToken == "" && (c.Username == "" || c.AppPassword == "") {
		return nil, fmt.Errorf("provide OPENCLOUD_BEARER_TOKEN, or OPENCLOUD_USERNAME + OPENCLOUD_APP_PASSWORD")
	}
	return c, nil
}

// getenv returns the env value, or def when the variable is unset or empty.
// Treating empty as unset lets docker-compose pass `${VAR:-}` for optional
// settings without clobbering the built-in defaults.
func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getint(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return n
		}
	}
	return def
}

func getfloat(key string, def float64) float64 {
	if v, ok := os.LookupEnv(key); ok {
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return f
		}
	}
	return def
}

func getbool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(strings.TrimSpace(v)); err == nil {
			return b
		}
	}
	return def
}

func getduration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(strings.TrimSpace(v)); err == nil {
			return d
		}
	}
	return def
}
