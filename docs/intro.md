---
sidebar_position: 1
slug: /
title: Overview
---

# OpenCloud ImmichFrame

[ImmichFrame](https://github.com/immichFrame/ImmichFrame) is an addition to Immich that displays
images on digital photo frames. While it is developed specifically for [Immich](immich.app) (which
is an awesome tool for photo management that everybody should use) with this addition to OpenCloud,
the ImmichFrame clients can also display images that are stored in an OpenCloud space. No Immich
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
