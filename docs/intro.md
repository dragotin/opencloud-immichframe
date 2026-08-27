---
sidebar_position: 1
slug: /
title: Overview
---

# opencloud-immichframe service reference

opencloud-immichframe serves the [ImmichFrame](https://github.com/immichFrame/ImmichFrame)
HTTP API from an [OpenCloud](https://opencloud.eu) space instead of an Immich
server. It reads photos from the space over the LibreGraph (listing) and WebDAV
(content) APIs, so the ImmichFrame web UI and clients work unchanged.

- **[Environment variables](configuration/environment-variables)** — every env var parsed by the service at startup.
- **[Example configuration](configuration/example-config)** — YAML config with all defaults.
- **[Deprecations](configuration/deprecations)** — renamed or removed env vars.

For installation and operational docs see the
[GitHub repository](https://github.com/dragotin/immichframe-opencloud#readme).
