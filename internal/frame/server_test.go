// SPDX-FileCopyrightText: 2026 Klaas Freitag <opensource@freisturz.de>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package frame

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"github.com/immichFrame/opencloud-immichframe/pkg/config"
	"github.com/immichFrame/opencloud-immichframe/internal/opencloud"
)

const space = "space-1"

func openCloudMock(t *testing.T) *httptest.Server {
	t.Helper()
	img := bytes.Repeat([]byte("A"), 100)

	mux := http.NewServeMux()
	// Root has one image (a.jpg) and one folder (trips).
	mux.HandleFunc("GET /graph/v1.0/drives/"+space+"/items/"+space+"/children", func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{"value": []map[string]any{
			{"id": "id-a", "name": "a.jpg", "size": 100, "eTag": `"etag-a"`,
				"lastModifiedDateTime": time.Unix(1700000000, 0).UTC().Format(time.RFC3339),
				"file":                 map[string]any{"mimeType": "image/jpeg"}},
			{"id": "id-trips", "name": "trips", "folder": map[string]any{"childCount": 1}},
		}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	// The trips folder holds one image (b.jpg).
	mux.HandleFunc("GET /graph/v1.0/drives/"+space+"/items/id-trips/children", func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{"value": []map[string]any{
			{"id": "id-b", "name": "b.jpg", "size": 100, "eTag": `"etag-b"`,
				"lastModifiedDateTime": time.Unix(1700000000, 0).UTC().Format(time.RFC3339),
				"file":                 map[string]any{"mimeType": "image/jpeg"}},
		}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("GET /graph/v1.0/drives/"+space, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": space, "name": "My Space", "description": "holiday pics",
		})
	})
	// Serve any file in the space.
	mux.HandleFunc("GET /dav/spaces/"+space+"/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		http.ServeContent(w, r, "img.jpg", time.Unix(1700000000, 0), bytes.NewReader(img))
	})
	return httptest.NewServer(mux)
}

func newTestServer(t *testing.T, secret string) http.Handler {
	t.Helper()
	oc := openCloudMock(t)
	t.Cleanup(oc.Close)

	client, err := opencloud.New(context.Background(), opencloud.Options{
		BaseURL: oc.URL, SpaceID: space, BearerToken: "x",
	})
	if err != nil {
		t.Fatalf("opencloud.New: %v", err)
	}
	cat := NewCatalog(context.Background(), client, 0)
	return NewServer(config.ClientSettings{PhotoDateFormat: "2006-01-02", ShowPhotoDate: true}, secret, "test-1.0", "", cat, client).Handler()
}

func do(t *testing.T, h http.Handler, method, path, auth string, hdr map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestAuth(t *testing.T) {
	h := newTestServer(t, "s3cr3t")

	if rec := do(t, h, "GET", "/api/Asset", "", nil); rec.Code != http.StatusUnauthorized {
		t.Errorf("no auth: got %d, want 401", rec.Code)
	}
	if rec := do(t, h, "GET", "/api/Asset", "Bearer wrong", nil); rec.Code != http.StatusUnauthorized {
		t.Errorf("wrong secret: got %d, want 401", rec.Code)
	}
	if rec := do(t, h, "GET", "/api/Asset", "Bearer s3cr3t", nil); rec.Code != http.StatusOK {
		t.Errorf("good secret: got %d, want 200", rec.Code)
	}
	// Version is public.
	if rec := do(t, h, "GET", "/api/Config/Version", "", nil); rec.Code != http.StatusOK {
		t.Errorf("version unauth: got %d, want 200", rec.Code)
	}
}

func TestGetAssetsShape(t *testing.T) {
	h := newTestServer(t, "")
	rec := do(t, h, "GET", "/api/Asset", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var assets []AssetResponseDto
	if err := json.Unmarshal(rec.Body.Bytes(), &assets); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(assets) != 2 {
		t.Fatalf("want 2 assets (a.jpg + trips/b.jpg), got %d", len(assets))
	}
	a := assets[0]
	if a.Type != AssetTypeImage {
		t.Errorf("type = %d, want %d (IMAGE)", a.Type, AssetTypeImage)
	}
	if a.OriginalFileName != "a.jpg" {
		t.Errorf("originalFileName = %q", a.OriginalFileName)
	}
	if len(a.ID) != 36 {
		t.Errorf("id is not a uuid: %q", a.ID)
	}
}

// albumOf fetches AlbumInfo for an asset and returns the single album.
func albumOf(t *testing.T, h http.Handler, assetID string) AlbumResponseDto {
	t.Helper()
	rec := do(t, h, "GET", "/api/Asset/"+assetID+"/AlbumInfo", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("AlbumInfo status %d", rec.Code)
	}
	var albums []AlbumResponseDto
	if err := json.Unmarshal(rec.Body.Bytes(), &albums); err != nil {
		t.Fatalf("AlbumInfo unmarshal: %v", err)
	}
	if len(albums) != 1 {
		t.Fatalf("AlbumInfo len = %d, want 1", len(albums))
	}
	return albums[0]
}

// assetIDByName returns the uuid of the asset with the given original filename.
func assetIDByName(t *testing.T, h http.Handler, name string) string {
	t.Helper()
	rec := do(t, h, "GET", "/api/Asset", "", nil)
	var assets []AssetResponseDto
	_ = json.Unmarshal(rec.Body.Bytes(), &assets)
	for _, a := range assets {
		if a.OriginalFileName == name {
			return a.ID
		}
	}
	t.Fatalf("asset %q not found", name)
	return ""
}

func TestGetAssetBinaryAndRange(t *testing.T) {
	h := newTestServer(t, "")
	// discover the id
	rec := do(t, h, "GET", "/api/Asset", "", nil)
	var assets []AssetResponseDto
	_ = json.Unmarshal(rec.Body.Bytes(), &assets)
	id := assets[0].ID

	full := do(t, h, "GET", "/api/Asset/"+id+"/Asset", "", nil)
	if full.Code != http.StatusOK {
		t.Fatalf("full status %d", full.Code)
	}
	if full.Header().Get("Accept-Ranges") != "bytes" {
		t.Errorf("missing Accept-Ranges")
	}
	if full.Body.Len() != 100 {
		t.Errorf("body len %d, want 100", full.Body.Len())
	}

	part := do(t, h, "GET", "/api/Asset/"+id+"/Asset", "", map[string]string{"Range": "bytes=0-9"})
	if part.Code != http.StatusPartialContent {
		t.Fatalf("range status %d, want 206", part.Code)
	}
	if !strings.HasPrefix(part.Header().Get("Content-Range"), "bytes 0-9/") {
		t.Errorf("Content-Range = %q", part.Header().Get("Content-Range"))
	}
	if part.Body.Len() != 10 {
		t.Errorf("partial body len %d, want 10", part.Body.Len())
	}

	bad := do(t, h, "GET", "/api/Asset/"+id+"/Asset", "", map[string]string{"Range": "bytes=99999-"})
	if bad.Code != http.StatusRequestedRangeNotSatisfiable {
		t.Errorf("bad range status %d, want 416", bad.Code)
	}
}

func TestEmptyArraysAndInfo(t *testing.T) {
	h := newTestServer(t, "")
	rec := do(t, h, "GET", "/api/Asset", "", nil)
	var assets []AssetResponseDto
	_ = json.Unmarshal(rec.Body.Bytes(), &assets)
	id := assets[0].ID

	faces := do(t, h, "GET", "/api/Asset/"+id+"/AssetFaces", "", nil)
	if faces.Code != http.StatusOK || strings.TrimSpace(faces.Body.String()) != "[]" {
		t.Errorf("AssetFaces = %d %q, want 200 []", faces.Code, faces.Body.String())
	}

	// A root-level asset (a.jpg) falls back to the space as its album.
	rootAlbum := albumOf(t, h, id)
	if rootAlbum.ID != uuidV5(namespace, space) {
		t.Errorf("root album id = %q, want space uuid", rootAlbum.ID)
	}
	if rootAlbum.AlbumName != "My Space" || rootAlbum.Description != "holiday pics" {
		t.Errorf("root album name/desc = %q/%q, want space name+description", rootAlbum.AlbumName, rootAlbum.Description)
	}
	if rootAlbum.AssetCount != 1 {
		t.Errorf("root album assetCount = %d, want 1 (only a.jpg at root)", rootAlbum.AssetCount)
	}

	// An asset inside a folder reports that folder as its album.
	folderAlbum := albumOf(t, h, assetIDByName(t, h, "b.jpg"))
	if folderAlbum.ID != uuidV5(namespace, "id-trips") {
		t.Errorf("folder album id = %q, want trips folder uuid", folderAlbum.ID)
	}
	if folderAlbum.AlbumName != "trips" {
		t.Errorf("folder album name = %q, want \"trips\"", folderAlbum.AlbumName)
	}
	if folderAlbum.AssetCount != 1 {
		t.Errorf("folder album assetCount = %d, want 1", folderAlbum.AssetCount)
	}

	info := do(t, h, "GET", "/api/Asset/"+id+"/AssetInfo", "", nil)
	if info.Code != http.StatusOK {
		t.Fatalf("AssetInfo status %d", info.Code)
	}
	var dto AssetResponseDto
	if err := json.Unmarshal(info.Body.Bytes(), &dto); err != nil || dto.ID != id {
		t.Errorf("AssetInfo mismatch: %v id=%q", err, dto.ID)
	}

	miss := do(t, h, "GET", "/api/Asset/00000000-0000-4000-8000-000000000000/AssetInfo", "", nil)
	if miss.Code != http.StatusNotFound {
		t.Errorf("unknown id status %d, want 404", miss.Code)
	}
}

func TestRandomImageAndInfo(t *testing.T) {
	h := newTestServer(t, "")
	rec := do(t, h, "GET", "/api/Asset/RandomImageAndInfo", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var resp ImageResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	data, err := base64.StdEncoding.DecodeString(resp.RandomImageBase64)
	if err != nil || len(data) != 100 {
		t.Errorf("bad base64 image: err=%v len=%d", err, len(data))
	}
	if resp.PhotoDate == "" {
		t.Errorf("expected photoDate")
	}
}

func TestWeatherCalendarStubsAndUnknownAPI(t *testing.T) {
	h := newTestServer(t, "")

	wx := do(t, h, "GET", "/api/Weather", "", nil)
	if wx.Code != http.StatusOK || strings.TrimSpace(wx.Body.String()) != "null" {
		t.Errorf("Weather = %d %q, want 200 null", wx.Code, wx.Body.String())
	}

	cal := do(t, h, "GET", "/api/Calendar", "", nil)
	if cal.Code != http.StatusOK || strings.TrimSpace(cal.Body.String()) != "[]" {
		t.Errorf("Calendar = %d %q, want 200 []", cal.Code, cal.Body.String())
	}

	// Unknown API routes must 404, never fall through to SPA HTML.
	unknown := do(t, h, "GET", "/api/DoesNotExist", "", nil)
	if unknown.Code != http.StatusNotFound {
		t.Errorf("unknown api = %d, want 404", unknown.Code)
	}
	if strings.Contains(unknown.Body.String(), "<!doctype html") {
		t.Errorf("unknown api returned HTML: %q", unknown.Body.String())
	}
}

func TestConfig(t *testing.T) {
	h := newTestServer(t, "")
	rec := do(t, h, "GET", "/api/Config", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var cfg ClientSettingsDto
	if err := json.Unmarshal(rec.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	ver := do(t, h, "GET", "/api/Config/Version", "", nil)
	if strings.TrimSpace(ver.Body.String()) != `"test-1.0"` {
		t.Errorf("version body = %q", ver.Body.String())
	}
}

func TestUUIDStable(t *testing.T) {
	a := uuidV5(namespace, "id-a")
	b := uuidV5(namespace, "id-a")
	c := uuidV5(namespace, "id-b")
	if a != b {
		t.Errorf("uuid not stable: %q != %q", a, b)
	}
	if a == c {
		t.Errorf("uuid collision for distinct ids")
	}
	if len(a) != 36 || a[14] != '5' {
		t.Errorf("not a v5 uuid: %q", a)
	}
}
