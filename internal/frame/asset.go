// SPDX-FileCopyrightText: 2026 Klaas Freitag <opensource@freisturz.de>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package frame

import (
	"encoding/base64"
	"io"
	"net/http"
)

// handleGetAssets returns the full list of image assets.
func (s *Server) handleGetAssets(w http.ResponseWriter, r *http.Request) {
	assets, err := s.catalog.All(r.Context())
	if err != nil {
		writeText(w, http.StatusBadGateway, "failed to list assets: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, assets)
}

// handleGetAssetInfo returns a single asset's metadata.
func (s *Server) handleGetAssetInfo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	item, ok, err := s.catalog.ByID(r.Context(), id)
	if err != nil {
		writeText(w, http.StatusBadGateway, err.Error())
		return
	}
	if !ok {
		writeText(w, http.StatusNotFound, "asset not found")
		return
	}
	writeJSON(w, http.StatusOK, toDTO(id, item))
}

// handleGetAlbumInfo returns the album an asset belongs to. Folders inside the
// space are treated as albums, so the album is the asset's immediate parent
// folder; assets at the space root fall back to the space itself.
func (s *Server) handleGetAlbumInfo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	item, ok, err := s.catalog.ByID(r.Context(), id)
	if err != nil {
		writeText(w, http.StatusBadGateway, err.Error())
		return
	}
	if !ok {
		writeText(w, http.StatusNotFound, "asset not found")
		return
	}

	var album AlbumResponseDto
	if item.Album != nil {
		album = AlbumResponseDto{
			ID:         uuidV5(namespace, item.Album.ID),
			AlbumName:  item.Album.Name,
			AssetCount: s.catalog.AlbumAssetCount(r.Context(), item.Album.ID),
			AlbumUsers: []any{},
		}
	} else {
		space := s.client.Space()
		album = AlbumResponseDto{
			ID:          uuidV5(namespace, space.ID),
			AlbumName:   space.Name,
			Description: space.Description,
			AssetCount:  s.catalog.AlbumAssetCount(r.Context(), ""),
			AlbumUsers:  []any{},
		}
	}
	writeJSON(w, http.StatusOK, []AlbumResponseDto{album})
}

// handleEmptyArray backs endpoints we have no data for (e.g. AssetFaces).
func (s *Server) handleEmptyArray(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, []any{})
}

// handleGetAsset streams the image bytes, honouring Range requests.
func (s *Server) handleGetAsset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	item, ok, err := s.catalog.ByID(r.Context(), id)
	if err != nil {
		writeText(w, http.StatusBadGateway, err.Error())
		return
	}
	if !ok {
		writeText(w, http.StatusNotFound, "asset not found")
		return
	}

	res, err := s.client.Download(r.Context(), item, r.Header.Get("Range"))
	if err != nil {
		writeText(w, http.StatusBadGateway, "download failed: "+err.Error())
		return
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	w.Header().Set("Accept-Ranges", "bytes")

	if res.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	if res.ContentType != "" {
		w.Header().Set("Content-Type", res.ContentType)
	}

	if res.StatusCode == http.StatusPartialContent && res.ContentRange != "" {
		w.Header().Set("Content-Range", res.ContentRange)
		if res.ContentLength > 0 {
			w.Header().Set("Content-Length", itoa(res.ContentLength))
		}
		w.WriteHeader(http.StatusPartialContent)
	} else if res.ContentLength > 0 {
		w.Header().Set("Content-Length", itoa(res.ContentLength))
	}

	_, _ = io.Copy(w, res.Body)
}

// handleRandomImageAndInfo returns a random image encoded as base64 plus info.
func (s *Server) handleRandomImageAndInfo(w http.ResponseWriter, r *http.Request) {
	_, item, ok, err := s.catalog.Random(r.Context())
	if err != nil {
		writeText(w, http.StatusBadGateway, err.Error())
		return
	}
	if !ok {
		writeText(w, http.StatusNotFound, "no image asset was found")
		return
	}

	res, err := s.client.Download(r.Context(), item, "")
	if err != nil {
		writeText(w, http.StatusBadGateway, "download failed: "+err.Error())
		return
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		writeText(w, http.StatusBadGateway, "read failed: "+err.Error())
		return
	}

	photoDate := ""
	if s.cfg.ShowPhotoDate && !item.LastModified.IsZero() {
		photoDate = item.LastModified.Format(s.cfg.PhotoDateFormat)
	}

	writeJSON(w, http.StatusOK, ImageResponse{
		RandomImageBase64:    base64.StdEncoding.EncodeToString(data),
		ThumbHashImageBase64: "",
		PhotoDate:            photoDate,
		ImageLocation:        "",
	})
}
