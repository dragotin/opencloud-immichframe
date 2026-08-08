// SPDX-FileCopyrightText: 2026 Klaas Freitag <opensource@freisturz.de>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package opencloud

import (
	"io"
	"time"

	libregraph "github.com/opencloud-eu/libre-graph-api-go"
)

// DriveItem is the flattened representation of a LibreGraph driveItem that we
// care about. Path is the space-relative path (no leading slash) usable to
// build a WebDAV URL, e.g. "photos/summer/a.jpg".
type DriveItem struct {
	ID           string
	Name         string
	Path         string
	Size         int64
	MimeType     string
	ETag         string
	LastModified time.Time
	// Album is the immediate containing folder, treated as the item's album.
	// nil when the item sits at the space root.
	Album *Album
}

// Album is a folder inside the space, exposed to ImmichFrame as an album.
type Album struct {
	ID   string
	Name string
	Path string
}

// DownloadResult carries a streamed response from the WebDAV endpoint. Body
// must be closed by the caller.
type DownloadResult struct {
	StatusCode    int
	ContentType   string
	ContentRange  string
	ContentLength int64
	Body          io.ReadCloser
}

// SpaceInfo is the resolved metadata of the served space.
type SpaceInfo struct {
	ID          string
	Name        string
	Description string
}

// driveList and driveItemList decode the LibreGraph collection envelopes
// ({"value": [...]}) returned by the drives and children endpoints into the
// generated libregraph types.
type driveList struct {
	Value []libregraph.Drive `json:"value"`
}

type driveItemList struct {
	Value []libregraph.DriveItem `json:"value"`
}
