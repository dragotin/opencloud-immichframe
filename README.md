# immichframe-opencloud

An [ImmichFrame](https://github.com/immichFrame/ImmichFrame)-compatible HTTP API
backend written in Go, serving photos from an **OpenCloud space** instead of an
Immich server. Point an existing ImmichFrame client at this service and it works
unchanged.

It talks to OpenCloud over the public HTTP APIs only — **LibreGraph** to list the
space, **WebDAV** (`/dav/spaces/{spaceId}/…`) to stream image bytes. No reva/gRPC
or monorepo coupling.

## Implemented endpoints

The contract matches `openApi/swagger.json` from the ImmichFrame repo:

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

Out of scope in this first cut: Weather, Calendar, EXIF/GPS location, people,
tags, thumbhash, and video assets.

## Albums

OpenCloud spaces have no album concept, so **folders inside the space are treated
as albums**. Each image's album is its immediate parent folder; `AlbumInfo`
returns that folder as the album (`albumName` = folder name, `assetCount` =
images directly in it). Images at the space **root** fall back to the space
itself (name + description) as their album. Hidden dot-folders (e.g. OpenCloud's
internal `.space`) are skipped entirely — they appear neither as albums nor in
the slideshow. The web UI shows the album name when `FRAME_SHOW_ALBUM_NAME` is
enabled (the default).

## Asset IDs

OpenCloud driveItem ids are opaque (they contain `$` and `!`) and are not UUIDs.
Each is mapped to a **stable UUIDv5** (derived from a fixed namespace) so the
`{id}` path segment round-trips cleanly and clients see stable ids across
restarts. Album ids are derived the same way from the folder id.

## Configuration

All configuration is via environment variables.

### OpenCloud backend

| Variable | Required | Default | Notes |
| --- | --- | --- | --- |
| `OPENCLOUD_BASE_URL` | yes | – | e.g. `https://cloud.example.com` |
| `OPENCLOUD_SPACE_ID` | one of id/name | – | Drive (space) id to serve. |
| `OPENCLOUD_SPACE_NAME` | one of id/name | – | Resolved via `/graph/v1.0/me/drives`. |
| `OPENCLOUD_USERNAME` | with app pw | – | For Basic auth (app token). |
| `OPENCLOUD_APP_PASSWORD` | with username | – | An OpenCloud app password/token. |
| `OPENCLOUD_BEARER_TOKEN` | alt to basic | – | Static bearer token; takes precedence. |
| `OPENCLOUD_INSECURE_TLS` | no | `false` | Accept self-signed / invalid TLS certs (dev only). Also settable via the `-insecure` flag. |

### Frame API

| Variable | Default | Notes |
| --- | --- | --- |
| `LISTEN_ADDR` | `:8080` | |
| `AUTH_SECRET` | *(empty)* | If set, clients must send `Authorization: Bearer <secret>`. Empty = open. |
| `CATALOG_REFRESH` | `5m` | How often to re-scan the space. |
| `WEB_ROOT` | *(empty)* | Directory with the built ImmichFrame web UI to serve at `/`. Set this to make `:8080` a full drop-in (UI + API on one origin), which the desktop/web clients require. |

### Serving the web UI (for the desktop / webview clients)

The ImmichFrame **desktop** app (and the web client) load a single URL that must
serve the whole ImmichFrame web app at `/` *and* the API at `/api`. Point
`WEB_ROOT` at a build of [`immichFrame.Web`](https://github.com/immichFrame/ImmichFrame):

```sh
cd ImmichFrame/immichFrame.Web && npm install && npm run build   # produces build/
# then start this service with:
WEB_ROOT=/path/to/ImmichFrame/immichFrame.Web/build ...other env... \
  immichframe-opencloud
```

Now `http://localhost:8080/` serves the UI and `http://localhost:8080/api/*` the
API. Configure the desktop client by writing that URL into its settings file
(`~/.config/immichFrame/Settings.txt` on Linux — plain text, just the URL).
Static assets are served without auth; only `/api/*` honours `AUTH_SECRET`.

### Client settings (surfaced via `/api/Config`)

These drive the web-UI slideshow. All are optional — unset (or empty) falls back
to the built-in default.

| Variable | Default | Notes |
| --- | --- | --- |
| `FRAME_INTERVAL` | `8` | Seconds each image is shown. |
| `FRAME_TRANSITION_DURATION` | `1` | Crossfade seconds. |
| `FRAME_SHOW_CLOCK` | `true` | |
| `FRAME_CLOCK_FORMAT` | `HH:mm` | date-fns **time** tokens. |
| `FRAME_CLOCK_DATE_FORMAT` | `eee, MMM d` | date-fns **date** tokens (date line above the clock). |
| `FRAME_SHOW_PROGRESS_BAR` | `true` | Slideshow progress bar. |
| `FRAME_SHOW_PHOTO_DATE` | `true` | |
| `FRAME_PHOTO_DATE_FORMAT` | `2006-01-02` | **Go** layout (formatted server-side). |
| `FRAME_SHOW_IMAGE_LOCATION` | `false` | No EXIF/GPS in OpenCloud yet. |
| `FRAME_IMAGE_LOCATION_FORMAT` | `City,State,Country` | |
| `FRAME_SHOW_IMAGE_DESC` | `false` | |
| `FRAME_SHOW_ALBUM_NAME` | `true` | Album = the image's folder in the space. |
| `FRAME_LANGUAGE` | `en` | |

> Two different date syntaxes: `FRAME_CLOCK_*` use **date-fns** tokens
> (`eee, MMM d`) because the web UI formats the clock; `FRAME_PHOTO_DATE_FORMAT`
> uses a **Go** layout (`2006-01-02`) because the backend formats the photo date.

People/faces and tags are intentionally omitted — OpenCloud spaces have no such
concept, so `showPeopleDesc`/`showTagsDesc` are never emitted.

In the Docker Compose setup these are read from `.env` and forwarded to the
container by the `environment:` block in `docker-compose.yml`; see
`.env.example` for the full commented list.

## Running

```sh
export OPENCLOUD_BASE_URL=https://cloud.example.com
export OPENCLOUD_SPACE_NAME="Photo Frame"
export OPENCLOUD_USERNAME=frame
export OPENCLOUD_APP_PASSWORD=xxxxxxxx
export AUTH_SECRET=my-shared-secret

go run ./cmd/immichframe-opencloud
```

### Command-line flags

| Flag | Notes |
| --- | --- |
| `-insecure` | Accept self-signed / invalid TLS certificates from the OpenCloud server. Overrides `OPENCLOUD_INSECURE_TLS`. Development only. |

```sh
go run ./cmd/immichframe-opencloud -insecure
```

### Docker

```sh
docker build -t immichframe-opencloud .
docker run --rm -p 8080:8080 \
  -e OPENCLOUD_BASE_URL=https://cloud.example.com \
  -e OPENCLOUD_SPACE_NAME="Photo Frame" \
  -e OPENCLOUD_USERNAME=frame \
  -e OPENCLOUD_APP_PASSWORD=xxxxxxxx \
  -e AUTH_SECRET=my-shared-secret \
  immichframe-opencloud
```

## Docker Compose (official web UI + OpenCloud API via Traefik)

`docker-compose.yml` runs three services so the **official** ImmichFrame web UI
displays photos from your OpenCloud space:

```
client ─▶ traefik :8080
           ├─ PathPrefix(`/api`) ─▶ immichframe-opencloud   (OpenCloud photos)
           └─ PathPrefix(`/`)    ─▶ immichframe (official)   (ImmichFrame web UI)
```

Traefik routes every `/api/*` request to the Go backend, so the official
container is used only for its web UI — its own Immich backend is never touched.

```sh
cp .env.example .env      # then edit it
docker compose up -d --build
# open http://localhost:8080   (or point the desktop client's Settings.txt at it)
```

Notes:
- Set `OPENCLOUD_SPACE_NAME` (e.g. `Images`) rather than `OPENCLOUD_SPACE_ID` —
  space ids contain `$`, which Compose interpolates (double it as `$$` if you
  must use the id).
- If OpenCloud runs on the Docker host, keep
  `OPENCLOUD_BASE_URL=https://host.docker.internal:9200` (the compose file maps
  `host.docker.internal`). For a self-signed dev cert, `OPENCLOUD_INSECURE_TLS=true`.
- Rootless Docker: point Traefik at your runtime socket (see the comment in the
  compose file).
- The official image needs `ImmichServerUrl`/`ApiKey` set just to boot; the
  placeholders in `.env.example` are enough since its `/api` is bypassed.

## Smoke test

If the API runs **open** (`AUTH_SECRET` empty, as in the Docker default), drop
the `-H "Authorization…"` header from every command below. Otherwise set
`SECRET` to your `AUTH_SECRET`.

```sh
SECRET=my-shared-secret
auth=(-H "Authorization: Bearer $SECRET")   # leave empty () if running open

# List assets and grab the first id.
curl -s "${auth[@]}" localhost:8080/api/Asset | jq '.[0]'
ID=$(curl -s "${auth[@]}" localhost:8080/api/Asset | jq -r '.[0].id')

# Full image, then a range request (expect 206 + Content-Range).
curl "${auth[@]}" "localhost:8080/api/Asset/$ID/Asset" -o out.jpg
curl -D- "${auth[@]}" -H 'Range: bytes=0-99' \
  "localhost:8080/api/Asset/$ID/Asset" -o /dev/null

# Album = the image's folder (empty description means it's a root-level image).
curl -s "${auth[@]}" "localhost:8080/api/Asset/$ID/AlbumInfo" | jq '.[0] | {albumName, assetCount}'

# Random image payload and version.
curl -s "${auth[@]}" localhost:8080/api/Asset/RandomImageAndInfo | jq 'keys'
curl -s localhost:8080/api/Config/Version

# Web UI served on the same origin (expect 200 text/html).
curl -s -o /dev/null -w '%{http_code} %{content_type}\n' localhost:8080/
```

## Development

```sh
go build ./...
go vet ./...
go test ./...
```

This project is developed by Klaas Freitag with the assistance of
[Claude](https://www.anthropic.com/claude). Thanks Domme for help
and motivation :-)

