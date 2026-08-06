// SPDX-FileCopyrightText: 2026 Klaas Freitag <opensource@freisturz.de>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package opencloud

import (
	"io"
	"time"
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

// graphChildrenResponse is the envelope returned by the LibreGraph children
// endpoints.
type graphChildrenResponse struct {
	Value []graphDriveItem `json:"value"`
}

type graphDriveItem struct {
	ID                   string          `json:"id"`
	Name                 string          `json:"name"`
	Size                 int64           `json:"size"`
	ETag                 string          `json:"eTag"`
	LastModifiedDateTime time.Time       `json:"lastModifiedDateTime"`
	File                 *graphFileFacet `json:"file,omitempty"`
	Folder               *graphFolder    `json:"folder,omitempty"`
}

type graphFileFacet struct {
	MimeType string `json:"mimeType"`
}

type graphFolder struct {
	ChildCount int64 `json:"childCount"`
}

// graphDrivesResponse is returned by /graph/v1.0/me/drives.
type graphDrivesResponse struct {
	Value []graphDrive `json:"value"`
}

type graphDrive struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SpaceInfo is the resolved metadata of the served space.
type SpaceInfo struct {
	ID          string
	Name        string
	Description string
}
