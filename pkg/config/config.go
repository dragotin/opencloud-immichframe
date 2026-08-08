// SPDX-FileCopyrightText: 2026 Klaas Freitag <opensource@freisturz.de>
//
// SPDX-License-Identifier: GPL-3.0-or-later

// Package config defines the service configuration. Every user-configurable
// field carries `yaml`, `env` and `desc` struct tags following the OpenCloud
// service convention: `yaml` is the config-file key, `env` the (semicolon
// separated) environment variable names, and `desc` the human documentation
// consumed by the docs generator. Defaults live in the defaults package, not
// in tags, so the generator can render them from a single source.
package config

import "time"

// Config combines all configuration parts.
type Config struct {
	// Service identifies the service; used to locate its yaml config file.
	Service Service `yaml:"-"`

	OpenCloud   OpenCloud      `yaml:"opencloud"`
	HTTP        HTTP           `yaml:"http"`
	ImmichFrame ImmichFrame    `yaml:"immichframe"`
	Client      ClientSettings `yaml:"client"`
}

// Service defines the service name.
type Service struct {
	Name string
}

// OpenCloud holds the connection and auth settings for the OpenCloud backend
// (LibreGraph API + WebDAV). Provide either a bearer token, or a username plus
// app password.
type OpenCloud struct {
	URL         string `yaml:"url" env:"OC_URL;IMMICHFRAME_OPENCLOUD_URL" desc:"Base URL of the OpenCloud instance (used for Graph API and WebDAV calls)."`
	SpaceID     string `yaml:"space_id" env:"IMMICHFRAME_OPENCLOUD_SPACE_ID" desc:"OpenCloud space id to serve photos from. Provide this or IMMICHFRAME_OPENCLOUD_SPACE_NAME."`
	SpaceName   string `yaml:"space_name" env:"IMMICHFRAME_OPENCLOUD_SPACE_NAME" desc:"OpenCloud space name; used to resolve the space id when IMMICHFRAME_OPENCLOUD_SPACE_ID is empty."`
	Username    string `yaml:"username" env:"IMMICHFRAME_OPENCLOUD_USERNAME" desc:"Username for app-password authentication."`
	AppPassword string `yaml:"app_password" env:"IMMICHFRAME_OPENCLOUD_APP_PASSWORD" desc:"App password for the given username."`
	BearerToken string `yaml:"bearer_token" env:"IMMICHFRAME_OPENCLOUD_BEARER_TOKEN" desc:"Bearer token for authentication; takes precedence over username/app-password."`
	Insecure    bool   `yaml:"insecure" env:"OC_INSECURE;IMMICHFRAME_OPENCLOUD_INSECURE" desc:"Skip TLS verification when talking to OpenCloud (self-signed certs in dev)."`
}

// HTTP defines the frame HTTP service.
type HTTP struct {
	Addr string `yaml:"addr" env:"IMMICHFRAME_HTTP_ADDR" desc:"Bind address of the frame HTTP service."`
}

// ImmichFrame defines the immichframe server behaviour.
type ImmichFrame struct {
	AuthSecret     string        `yaml:"auth_secret" env:"IMMICHFRAME_AUTH_SECRET" desc:"Shared secret clients must send as 'Authorization: Bearer <secret>'. Empty runs the API open."`
	CatalogRefresh time.Duration `yaml:"catalog_refresh" env:"IMMICHFRAME_CATALOG_REFRESH" desc:"How often the photo catalog is refreshed (Go duration, e.g. 5m)."`
	WebRoot        string        `yaml:"web_root" env:"IMMICHFRAME_WEB_ROOT" desc:"Directory with the built ImmichFrame web UI to serve at /. Empty serves the API only."`
}

// ClientSettings mirrors the subset of ImmichFrame's ClientSettingsDto that we
// can populate. Values are surfaced through /api/Config and passed straight
// through to the frame client.
type ClientSettings struct {
	Interval            int     `yaml:"interval" env:"IMMICHFRAME_INTERVAL" desc:"Seconds each image is shown."`
	TransitionDuration  float64 `yaml:"transition_duration" env:"IMMICHFRAME_TRANSITION_DURATION" desc:"Crossfade duration in seconds."`
	ShowClock           bool    `yaml:"show_clock" env:"IMMICHFRAME_SHOW_CLOCK" desc:"Show the clock overlay."`
	ClockFormat         string  `yaml:"clock_format" env:"IMMICHFRAME_CLOCK_FORMAT" desc:"Clock time format (date-fns tokens)."`
	ClockDateFormat     string  `yaml:"clock_date_format" env:"IMMICHFRAME_CLOCK_DATE_FORMAT" desc:"Clock date format (date-fns tokens)."`
	ShowProgressBar     bool    `yaml:"show_progress_bar" env:"IMMICHFRAME_SHOW_PROGRESS_BAR" desc:"Show the slideshow progress bar."`
	ShowPhotoDate       bool    `yaml:"show_photo_date" env:"IMMICHFRAME_SHOW_PHOTO_DATE" desc:"Show the photo date overlay."`
	PhotoDateFormat     string  `yaml:"photo_date_format" env:"IMMICHFRAME_PHOTO_DATE_FORMAT" desc:"Photo date format (Go layout; formatted server-side)."`
	ShowImageLocation   bool    `yaml:"show_image_location" env:"IMMICHFRAME_SHOW_IMAGE_LOCATION" desc:"Show the image location overlay (no EXIF/GPS in OpenCloud yet)."`
	ImageLocationFormat string  `yaml:"image_location_format" env:"IMMICHFRAME_IMAGE_LOCATION_FORMAT" desc:"Image location format string."`
	ShowImageDesc       bool    `yaml:"show_image_desc" env:"IMMICHFRAME_SHOW_IMAGE_DESC" desc:"Show the image description overlay."`
	ShowAlbumName       bool    `yaml:"show_album_name" env:"IMMICHFRAME_SHOW_ALBUM_NAME" desc:"Show the album name overlay (album = the image's folder)."`
	Language            string  `yaml:"language" env:"IMMICHFRAME_LANGUAGE" desc:"UI language code."`
}
