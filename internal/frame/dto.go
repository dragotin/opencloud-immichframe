// SPDX-FileCopyrightText: 2026 Klaas Freitag <opensource@freisturz.de>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package frame

import "time"

// AssetType values mirror ImmichFrame's integer AssetTypeEnum
// (order IMAGE, VIDEO, AUDIO, OTHER).
const (
	AssetTypeImage = 0
	AssetTypeVideo = 1
	AssetTypeAudio = 2
	AssetTypeOther = 3
)

// AssetVisibility values mirror ImmichFrame's AssetVisibility enum.
const assetVisibilityTimeline = 0

// AssetResponseDto is the subset of ImmichFrame's AssetResponseDto that we can
// populate from an OpenCloud driveItem. Field names match the swagger contract.
type AssetResponseDto struct {
	ID               string    `json:"id"`
	Checksum         string    `json:"checksum"`
	CreatedAt        time.Time `json:"createdAt"`
	FileCreatedAt    time.Time `json:"fileCreatedAt"`
	FileModifiedAt   time.Time `json:"fileModifiedAt"`
	LocalDateTime    time.Time `json:"localDateTime"`
	OriginalFileName string    `json:"originalFileName"`
	OriginalPath     string    `json:"originalPath"`
	OwnerID          string    `json:"ownerId"`
	Type             int       `json:"type"`
	UpdatedAt        time.Time `json:"updatedAt"`
	Visibility       int       `json:"visibility"`

	IsFavorite bool `json:"isFavorite"`
	IsArchived bool `json:"isArchived"`
	IsTrashed  bool `json:"isTrashed"`
}

// ImageResponse mirrors the /api/Asset/RandomImageAndInfo payload.
type ImageResponse struct {
	RandomImageBase64    string `json:"randomImageBase64"`
	ThumbHashImageBase64 string `json:"thumbHashImageBase64"`
	PhotoDate            string `json:"photoDate"`
	ImageLocation        string `json:"imageLocation"`
}

// AlbumResponseDto is the subset of ImmichFrame's AlbumResponseDto we expose.
// OpenCloud spaces have no album concept, so this is only populated as a stub;
// the fields match the swagger contract for forward compatibility.
type AlbumResponseDto struct {
	ID          string    `json:"id"`
	AlbumName   string    `json:"albumName"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	AssetCount  int       `json:"assetCount"`
	Shared      bool      `json:"shared"`
	AlbumUsers  []any     `json:"albumUsers"`
}

// ClientSettingsDto mirrors the subset of ImmichFrame's config we surface.
type ClientSettingsDto struct {
	Interval            int     `json:"interval"`
	TransitionDuration  float64 `json:"transitionDuration"`
	ShowClock           bool    `json:"showClock"`
	ClockFormat         string  `json:"clockFormat"`
	ClockDateFormat     string  `json:"clockDateFormat"`
	ShowProgressBar     bool    `json:"showProgressBar"`
	ShowPhotoDate       bool    `json:"showPhotoDate"`
	PhotoDateFormat     string  `json:"photoDateFormat"`
	ShowImageDesc       bool    `json:"showImageDesc"`
	ShowImageLocation   bool    `json:"showImageLocation"`
	ImageLocationFormat string  `json:"imageLocationFormat"`
	ShowAlbumName       bool    `json:"showAlbumName"`
	Language            string  `json:"language"`
}
