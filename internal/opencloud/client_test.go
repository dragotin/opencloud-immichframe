// SPDX-FileCopyrightText: 2026 Klaas Freitag <opensource@freisturz.de>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package opencloud

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testSpace = "space-1"

// mockServer serves a tiny space: /a.jpg, /notes.txt, /sub/b.png.
func mockServer(t *testing.T) *httptest.Server {
	t.Helper()
	content := map[string][]byte{
		"a.jpg":     bytes.Repeat([]byte("A"), 100),
		"sub/b.png": bytes.Repeat([]byte("B"), 50),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /graph/v1.0/drives/"+testSpace+"/items/"+testSpace+"/children", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"value": []map[string]any{
			{"id": "id-a", "name": "a.jpg", "size": 100, "eTag": `"etag-a"`,
				"lastModifiedDateTime": time.Unix(1700000000, 0).UTC().Format(time.RFC3339),
				"file":                 map[string]any{"mimeType": "image/jpeg"}},
			{"id": "id-notes", "name": "notes.txt", "size": 5, "file": map[string]any{"mimeType": "text/plain"}},
			{"id": "id-sub", "name": "sub", "folder": map[string]any{"childCount": 1}},
			{"id": "id-dot", "name": ".space", "folder": map[string]any{"childCount": 1}},
		}})
	})
	mux.HandleFunc("GET /graph/v1.0/drives/"+testSpace+"/items/id-sub/children", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"value": []map[string]any{
			{"id": "id-b", "name": "b.png", "size": 50, "eTag": `"etag-b"`, "file": map[string]any{"mimeType": "image/png"}},
		}})
	})
	// A hidden dot-folder whose image must be excluded from the walk.
	mux.HandleFunc("GET /graph/v1.0/drives/"+testSpace+"/items/id-dot/children", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"value": []map[string]any{
			{"id": "id-hidden", "name": "cover.jpg", "size": 10, "file": map[string]any{"mimeType": "image/jpeg"}},
		}})
	})
	mux.HandleFunc("GET /dav/spaces/"+testSpace+"/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/dav/spaces/"+testSpace+"/")
		data, ok := content[name]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		http.ServeContent(w, r, name, time.Unix(1700000000, 0), bytes.NewReader(data))
	})

	return httptest.NewServer(mux)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func newTestClient(t *testing.T, base string) *Client {
	t.Helper()
	c, err := New(context.Background(), Options{
		BaseURL:     base,
		SpaceID:     testSpace,
		BearerToken: "test",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestListImages(t *testing.T) {
	srv := mockServer(t)
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	items, err := c.ListImages(context.Background())
	if err != nil {
		t.Fatalf("ListImages: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("want 2 images (txt + .space filtered), got %d: %+v", len(items), items)
	}

	byPath := map[string]DriveItem{}
	for _, it := range items {
		byPath[it.Path] = it
		if strings.HasPrefix(it.Path, ".space/") {
			t.Errorf("dot-folder image not skipped: %q", it.Path)
		}
	}
	if a, ok := byPath["a.jpg"]; !ok {
		t.Errorf("missing a.jpg: %+v", items)
	} else if a.Album != nil {
		t.Errorf("root image should have no album, got %+v", a.Album)
	}
	if b, ok := byPath["sub/b.png"]; !ok {
		t.Errorf("missing sub/b.png: %+v", items)
	} else {
		if b.ETag != "etag-b" {
			t.Errorf("etag not unquoted: %q", b.ETag)
		}
		if b.Album == nil || b.Album.Name != "sub" || b.Album.ID != "id-sub" {
			t.Errorf("expected album 'sub' (id-sub), got %+v", b.Album)
		}
	}
}

func TestDownloadFull(t *testing.T) {
	srv := mockServer(t)
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	res, err := c.Download(context.Background(), DriveItem{Path: "a.jpg", MimeType: "image/jpeg"}, "")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", res.StatusCode)
	}
	data, _ := io.ReadAll(res.Body)
	if len(data) != 100 {
		t.Fatalf("len = %d, want 100", len(data))
	}
}

func TestDownloadRange(t *testing.T) {
	srv := mockServer(t)
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	res, err := c.Download(context.Background(), DriveItem{Path: "a.jpg"}, "bytes=0-9")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", res.StatusCode)
	}
	if res.ContentRange == "" {
		t.Fatalf("missing Content-Range")
	}
	data, _ := io.ReadAll(res.Body)
	if len(data) != 10 {
		t.Fatalf("len = %d, want 10", len(data))
	}
}

func TestDownloadRangeNotSatisfiable(t *testing.T) {
	srv := mockServer(t)
	defer srv.Close()
	c := newTestClient(t, srv.URL)

	res, err := c.Download(context.Background(), DriveItem{Path: "a.jpg"}, "bytes=99999-")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if res.Body != nil {
		res.Body.Close()
	}
	if res.StatusCode != http.StatusRequestedRangeNotSatisfiable {
		t.Fatalf("status = %d, want 416", res.StatusCode)
	}
}
