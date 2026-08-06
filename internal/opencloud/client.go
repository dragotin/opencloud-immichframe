// SPDX-FileCopyrightText: 2026 Klaas Freitag <opensource@freisturz.de>
//
// SPDX-License-Identifier: GPL-3.0-or-later

// Package opencloud provides a thin client for reading images from an OpenCloud
// space over the public LibreGraph (listing) and WebDAV (content) APIs.
package opencloud

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// imageExtensions is the fallback set used when a driveItem carries no mime
// type in the file facet.
var imageExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
	".gif": true, ".heic": true, ".heif": true, ".bmp": true,
	".tif": true, ".tiff": true, ".avif": true,
}

// Client talks to a single OpenCloud instance and space.
type Client struct {
	baseURL string
	spaceID string
	space   SpaceInfo
	auth    func(*http.Request)
	http    *http.Client
}

// Options configures a Client.
type Options struct {
	BaseURL     string
	SpaceID     string
	SpaceName   string
	Username    string
	AppPassword string
	BearerToken string
	InsecureTLS bool
	Timeout     time.Duration
}

// New builds a Client. When SpaceID is empty, SpaceName is resolved against
// /graph/v1.0/me/drives.
func New(ctx context.Context, o Options) (*Client, error) {
	if o.Timeout == 0 {
		o.Timeout = 60 * time.Second
	}
	transport := &http.Transport{}
	if o.InsecureTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // opt-in for self-signed dev servers
	}

	var auth func(*http.Request)
	switch {
	case o.BearerToken != "":
		auth = func(r *http.Request) { r.Header.Set("Authorization", "Bearer "+o.BearerToken) }
	default:
		auth = func(r *http.Request) { r.SetBasicAuth(o.Username, o.AppPassword) }
	}

	c := &Client{
		baseURL: strings.TrimRight(o.BaseURL, "/"),
		spaceID: o.SpaceID,
		auth:    auth,
		http:    &http.Client{Timeout: o.Timeout, Transport: transport},
	}

	if c.spaceID == "" {
		d, err := c.resolveSpaceByName(ctx, o.SpaceName)
		if err != nil {
			return nil, err
		}
		c.spaceID = d.ID
		c.space = SpaceInfo{ID: d.ID, Name: d.Name, Description: d.Description}
	} else {
		// Best-effort: fetch the drive so we know its name/description too.
		if d, err := c.fetchDrive(ctx, c.spaceID); err == nil {
			c.space = SpaceInfo{ID: d.ID, Name: d.Name, Description: d.Description}
		} else {
			c.space = SpaceInfo{ID: c.spaceID}
		}
	}
	return c, nil
}

// SpaceID returns the resolved space (drive) id.
func (c *Client) SpaceID() string { return c.spaceID }

// Space returns the resolved space metadata (id, name, description).
func (c *Client) Space() SpaceInfo { return c.space }

func (c *Client) resolveSpaceByName(ctx context.Context, name string) (graphDrive, error) {
	u := c.baseURL + "/graph/v1.0/me/drives"
	var resp graphDrivesResponse
	if err := c.getJSON(ctx, u, &resp); err != nil {
		return graphDrive{}, fmt.Errorf("list drives: %w", err)
	}
	for _, d := range resp.Value {
		if strings.EqualFold(d.Name, name) {
			return d, nil
		}
	}
	return graphDrive{}, fmt.Errorf("no space named %q found", name)
}

func (c *Client) fetchDrive(ctx context.Context, id string) (graphDrive, error) {
	u := fmt.Sprintf("%s/graph/v1.0/drives/%s", c.baseURL, url.PathEscape(id))
	var d graphDrive
	if err := c.getJSON(ctx, u, &d); err != nil {
		return graphDrive{}, err
	}
	return d, nil
}

// ListImages recursively walks the space and returns every image asset.
func (c *Client) ListImages(ctx context.Context) ([]DriveItem, error) {
	var out []DriveItem
	// For an OpenCloud space the root item id equals the drive (space) id, so
	// the walk starts there. OpenCloud's Graph has no /drives/{id}/root/children
	// route; children are always fetched via /items/{itemId}/children.
	if err := c.walk(ctx, c.spaceID, "", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// walk lists the children of the item itemID whose contents live under prefix
// (space-relative, e.g. "photos/"), recursing into folders. album is the folder
// currently being listed (nil at the space root); images are tagged with it.
func (c *Client) walk(ctx context.Context, itemID, prefix string, album *Album, out *[]DriveItem) error {
	u := fmt.Sprintf("%s/graph/v1.0/drives/%s/items/%s/children", c.baseURL, url.PathEscape(c.spaceID), url.PathEscape(itemID))

	var resp graphChildrenResponse
	if err := c.getJSON(ctx, u, &resp); err != nil {
		return fmt.Errorf("list children of %q: %w", prefix, err)
	}

	for _, it := range resp.Value {
		itemPath := prefix + it.Name
		switch {
		case it.Folder != nil:
			// Skip hidden/dot folders (e.g. OpenCloud's internal ".space").
			if strings.HasPrefix(it.Name, ".") {
				continue
			}
			childAlbum := &Album{ID: it.ID, Name: it.Name, Path: itemPath}
			if err := c.walk(ctx, it.ID, itemPath+"/", childAlbum, out); err != nil {
				return err
			}
		case it.File != nil || isImageName(it.Name):
			mime := ""
			if it.File != nil {
				mime = it.File.MimeType
			}
			if !isImage(mime, it.Name) {
				continue
			}
			*out = append(*out, DriveItem{
				ID:           it.ID,
				Name:         it.Name,
				Path:         itemPath,
				Size:         it.Size,
				MimeType:     mime,
				ETag:         strings.Trim(it.ETag, `"`),
				LastModified: it.LastModifiedDateTime,
				Album:        album,
			})
		}
	}
	return nil
}

// Download streams the content of item over WebDAV, forwarding rangeHeader when
// non-empty.
func (c *Client) Download(ctx context.Context, item DriveItem, rangeHeader string) (*DownloadResult, error) {
	u := c.baseURL + "/dav/spaces/" + url.PathEscape(c.spaceID) + "/" + webdavEscapePath(item.Path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	c.auth(req)
	if rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		resp.Body.Close()
		if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
			return &DownloadResult{StatusCode: resp.StatusCode}, nil
		}
		return nil, fmt.Errorf("download %q: status %d: %s", item.Path, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = item.MimeType
	}
	return &DownloadResult{
		StatusCode:    resp.StatusCode,
		ContentType:   ct,
		ContentRange:  resp.Header.Get("Content-Range"),
		ContentLength: resp.ContentLength,
		Body:          resp.Body,
	}, nil
}

func (c *Client) getJSON(ctx context.Context, u string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	c.auth(req)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

func isImage(mime, name string) bool {
	if strings.HasPrefix(strings.ToLower(mime), "image/") {
		return true
	}
	return isImageName(name)
}

func isImageName(name string) bool {
	return imageExtensions[strings.ToLower(path.Ext(name))]
}

// webdavEscapePath percent-encodes each path segment while preserving slashes.
func webdavEscapePath(p string) string {
	segments := strings.Split(p, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return strings.Join(segments, "/")
}
