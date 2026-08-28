---
sidebar_position: 1
slug: /
title: Overview
---

# OpenCloud ImmichFrame

[ImmichFrame](https://github.com/immichFrame/ImmichFrame) is an addition to Immich that displays
images on **digital photo frames**. While it is developed to work with [Immich](https://immich.app) (which
is an awesome tool for photo management that everybody should use) with this addition to OpenCloud,
the ImmichFrame clients can also display images that are stored in an **OpenCloud space**. No Immich
is required.

immichframe-opencloud serves the [ImmichFrame](https://github.com/immichFrame/ImmichFrame)
HTTP API from an [OpenCloud](https://opencloud.eu) space instead of an Immich
server. It reads photos from the space over the [LibreGraph](https://docs.opencloud.eu/de/docs/dev/server/apis/http/graph/) (listing) and [WebDAV](https://docs.opencloud.eu/de/docs/dev/server/apis/http/webdav/)
(content) APIs, so the ImmichFrame web UI and clients work unchanged.

# Immichframe-opencloud Service Reference

The following gives a detailled listing about environment variables and configuration options.

- **[Environment variables](configuration/environment-variables)** — every env var parsed by the service at startup.
- **[Example configuration](configuration/example-config)** — YAML config with all defaults.
- **[Deprecations](configuration/deprecations)** — renamed or removed env vars.

For installation and operational docs see the
[GitHub repository](https://github.com/dragotin/opencloud-immichframe#readme).

# Details

## Albums

OpenCloud does not have an album concept. **Folders inside the space are treated
as albums**. Each image's album is its immediate parent folder**. Images at the space 
**root** fall back to the space itself (name + description) as their album. 
Hidden dot-folders (e.g. OpenCloud's internal `.space`) are skipped entirely — 
they appear neither as albums nor in the slideshow. The web UI shows the album name 
when `IMMICHFRAME_SHOW_ALBUM_NAME` is enabled (the default).

## Asset IDs

OpenCloud driveItem ids are opaque (they contain `$` and `!`) and are not UUIDs.
Each is mapped to a **stable UUIDv5** (derived from a fixed namespace) so the
`{id}` path segment round-trips cleanly and clients see stable ids across
restarts. 

Album ids are derived the same way from the folder id.
