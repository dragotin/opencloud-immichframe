---
sidebar_position: 4
slug: /api
title: API
---

# OpenCloud ImmichFrame API

The implemented endpoints are:


| Endpoint | Behaviour |
| --- | --- |
| `GET /api/Config` | Client settings (interval, clock, date format, …). |
| `GET /api/Config/Version` | Service version. **Unauthenticated.** |
| `GET /api/Asset` | List every image in the space. |
| `GET /api/Asset/{id}/AssetInfo` | Metadata for one asset. |
| `GET /api/Asset/{id}/AlbumInfo` | The album the asset belongs to — its parent folder (see [Albums](#albums)). |
| `GET /api/Asset/{id}/AssetFaces` | `[]` (faces not modelled). |
| `GET /api/Asset/{id}/Asset` | Stream image bytes; supports `Range` (206 / 416). |
| `GET /api/Asset/{id}/Image` | Deprecated alias of `…/Asset`. |
| `GET /api/Asset/RandomImageAndInfo` | Random image as base64 + info. |

The contract matches [openApi/swagger.json](https://github.com/immichFrame/ImmichFrame/blob/main/openApi/swagger.json) from the ImmichFrame repo.

