// SPDX-FileCopyrightText: 2026 Klaas Freitag <opensource@freisturz.de>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package frame

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/immichFrame/opencloud-immichframe/internal/opencloud"
)

// namespace is a fixed UUID used to derive stable UUIDv5 asset ids from opaque
// OpenCloud driveItem ids. Randomly generated once; never change it or existing
// clients would see all assets as new.
var namespace = mustUUID("6f0c9d2e-1b3a-4c5d-8e9f-0a1b2c3d4e5f")

// ownerID is a fixed synthetic owner for every served asset.
var ownerID = mustUUID("00000000-0000-4000-8000-000000000001")

type entry struct {
	uuid string
	item opencloud.DriveItem
}

// Catalog holds the current set of images from the space plus the uuid<->item
// mapping. It is safe for concurrent use and refreshes in the background.
type Catalog struct {
	client  *opencloud.Client
	refresh time.Duration

	mu      sync.RWMutex
	entries []entry
	byUUID  map[string]opencloud.DriveItem
	loaded  bool
}

// NewCatalog builds a Catalog and performs an initial (best-effort) load.
func NewCatalog(ctx context.Context, client *opencloud.Client, refresh time.Duration) *Catalog {
	c := &Catalog{client: client, refresh: refresh, byUUID: map[string]opencloud.DriveItem{}}
	_ = c.Reload(ctx) // errors surface later on request; don't block startup
	return c
}

// Run refreshes the catalog on the configured interval until ctx is cancelled.
func (c *Catalog) Run(ctx context.Context) {
	if c.refresh <= 0 {
		return
	}
	t := time.NewTicker(c.refresh)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			_ = c.Reload(ctx)
		}
	}
}

// Reload fetches the current image list from the space and swaps it in.
func (c *Catalog) Reload(ctx context.Context) error {
	items, err := c.client.ListImages(ctx)
	if err != nil {
		return err
	}
	entries := make([]entry, 0, len(items))
	byUUID := make(map[string]opencloud.DriveItem, len(items))
	for _, it := range items {
		id := uuidV5(namespace, it.ID)
		entries = append(entries, entry{uuid: id, item: it})
		byUUID[id] = it
	}

	c.mu.Lock()
	c.entries = entries
	c.byUUID = byUUID
	c.loaded = true
	c.mu.Unlock()
	return nil
}

// ensureLoaded lazily loads the catalog if the initial load failed.
func (c *Catalog) ensureLoaded(ctx context.Context) error {
	c.mu.RLock()
	loaded := c.loaded
	c.mu.RUnlock()
	if loaded {
		return nil
	}
	return c.Reload(ctx)
}

// All returns every asset as a DTO.
func (c *Catalog) All(ctx context.Context) ([]AssetResponseDto, error) {
	if err := c.ensureLoaded(ctx); err != nil {
		return nil, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]AssetResponseDto, 0, len(c.entries))
	for _, e := range c.entries {
		out = append(out, toDTO(e.uuid, e.item))
	}
	return out, nil
}

// AlbumAssetCount returns how many images belong to the album (folder) with the
// given OpenCloud folder id. An empty albumID counts root-level images.
func (c *Catalog) AlbumAssetCount(ctx context.Context, albumID string) int {
	if err := c.ensureLoaded(ctx); err != nil {
		return 0
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	n := 0
	for _, e := range c.entries {
		id := ""
		if e.item.Album != nil {
			id = e.item.Album.ID
		}
		if id == albumID {
			n++
		}
	}
	return n
}

// ByID returns the driveItem for an asset uuid.
func (c *Catalog) ByID(ctx context.Context, id string) (opencloud.DriveItem, bool, error) {
	if err := c.ensureLoaded(ctx); err != nil {
		return opencloud.DriveItem{}, false, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	it, ok := c.byUUID[id]
	return it, ok, nil
}

// Random returns a random asset (uuid + item).
func (c *Catalog) Random(ctx context.Context) (string, opencloud.DriveItem, bool, error) {
	if err := c.ensureLoaded(ctx); err != nil {
		return "", opencloud.DriveItem{}, false, err
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.entries) == 0 {
		return "", opencloud.DriveItem{}, false, nil
	}
	e := c.entries[rand.Intn(len(c.entries))] //nolint:gosec // non-crypto shuffle is fine
	return e.uuid, e.item, true, nil
}

func toDTO(id string, it opencloud.DriveItem) AssetResponseDto {
	mt := it.LastModified
	checksum := it.ETag
	if checksum == "" {
		checksum = uuidV5(namespace, it.ID+":checksum")
	}
	return AssetResponseDto{
		ID:               id,
		Checksum:         checksum,
		CreatedAt:        mt,
		FileCreatedAt:    mt,
		FileModifiedAt:   mt,
		LocalDateTime:    mt,
		OriginalFileName: it.Name,
		OriginalPath:     it.Path,
		OwnerID:          ownerID,
		Type:             AssetTypeImage,
		UpdatedAt:        mt,
		Visibility:       assetVisibilityTimeline,
	}
}

// uuidV5 derives an RFC 4122 version-5 UUID from a namespace UUID and a name.
func uuidV5(ns, name string) string {
	nsBytes := parseUUID(ns)
	h := sha1.New() //nolint:gosec // UUIDv5 is defined in terms of SHA-1
	h.Write(nsBytes)
	h.Write([]byte(name))
	sum := h.Sum(nil)

	var u [16]byte
	copy(u[:], sum[:16])
	u[6] = (u[6] & 0x0f) | 0x50 // version 5
	u[8] = (u[8] & 0x3f) | 0x80 // RFC 4122 variant
	return formatUUID(u)
}

func parseUUID(s string) []byte {
	var out []byte
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '-' {
			continue
		}
		out = append(out, ch)
	}
	b := make([]byte, 16)
	for i := 0; i < 16; i++ {
		_, _ = fmt.Sscanf(string(out[i*2:i*2+2]), "%02x", &b[i])
	}
	return b
}

func formatUUID(u [16]byte) string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(u[0:4]),
		binary.BigEndian.Uint16(u[4:6]),
		binary.BigEndian.Uint16(u[6:8]),
		binary.BigEndian.Uint16(u[8:10]),
		u[10:16],
	)
}

func mustUUID(s string) string {
	// validate by round-tripping; panics on malformed literals above
	b := parseUUID(s)
	var u [16]byte
	copy(u[:], b)
	return formatUUID(u)
}
