// SPDX-FileCopyrightText: 2026 Klaas Freitag <opensource@freisturz.de>
//
// SPDX-License-Identifier: GPL-3.0-or-later

// Package frame implements the ImmichFrame-compatible HTTP API backed by an
// OpenCloud space.
package frame

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/immichFrame/immichframe-opencloud/internal/config"
	"github.com/immichFrame/immichframe-opencloud/internal/opencloud"
)

// versionPath is served without authentication.
const versionPath = "/api/Config/Version"

// Server wires the catalog and OpenCloud client into the HTTP handlers.
type Server struct {
	cfg     config.ClientSettings
	secret  string
	version string
	webRoot string
	catalog *Catalog
	client  *opencloud.Client
}

// NewServer builds a Server. webRoot, when non-empty, is a directory whose
// contents are served at / (with SPA fallback) so the ImmichFrame web UI and
// this API share one origin — what the desktop/web clients expect.
func NewServer(cfg config.ClientSettings, secret, version, webRoot string, catalog *Catalog, client *opencloud.Client) *Server {
	return &Server{cfg: cfg, secret: secret, version: version, webRoot: webRoot, catalog: catalog, client: client}
}

// Handler returns the fully-wired HTTP handler.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/Config", s.handleGetConfig)
	mux.HandleFunc("GET "+versionPath, s.handleGetVersion)

	mux.HandleFunc("GET /api/Asset", s.handleGetAssets)
	mux.HandleFunc("GET /api/Asset/RandomImageAndInfo", s.handleRandomImageAndInfo)
	mux.HandleFunc("GET /api/Asset/{id}/AssetInfo", s.handleGetAssetInfo)
	mux.HandleFunc("GET /api/Asset/{id}/AssetFaces", s.handleEmptyArray)
	mux.HandleFunc("GET /api/Asset/{id}/AlbumInfo", s.handleGetAlbumInfo)
	mux.HandleFunc("GET /api/Asset/{id}/Asset", s.handleGetAsset)
	mux.HandleFunc("GET /api/Asset/{id}/Image", s.handleGetAsset) // deprecated alias

	// Stubs: no weather/calendar backend, but the web UI polls these. Return
	// clean empty payloads so the overlay stays blank instead of misrendering.
	mux.HandleFunc("GET /api/Weather", s.handleWeather)
	mux.HandleFunc("GET /api/Calendar", s.handleEmptyArray)

	// Serve the web UI (if configured) for every non-API path.
	mux.HandleFunc("GET /", s.handleStatic)

	// Authenticate the API (except the version endpoint); static assets are public.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") && r.URL.Path != versionPath {
			authMiddleware(s.secret, mux).ServeHTTP(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

// handleStatic serves files from webRoot, falling back to index.html so the
// single-page web UI can handle client-side routing.
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	// Never let an unknown API route fall through to the SPA HTML — clients
	// must get a real 404, not a 200 page they'd misparse.
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeText(w, http.StatusNotFound, "not found")
		return
	}
	if s.webRoot == "" {
		http.NotFound(w, r)
		return
	}
	clean := filepath.Clean("/" + strings.TrimPrefix(r.URL.Path, "/"))
	fp := filepath.Join(s.webRoot, clean)
	if !strings.HasPrefix(fp, s.webRoot) { // guard against path traversal
		http.NotFound(w, r)
		return
	}
	if info, err := os.Stat(fp); err == nil && !info.IsDir() {
		http.ServeFile(w, r, fp)
		return
	}
	http.ServeFile(w, r, filepath.Join(s.webRoot, "index.html"))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
