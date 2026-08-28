---
sidebar_position: 3
slug: /run
title: Install
---

# Run OpenCloud Immichframe

## Go Build

```sh
export OC_URL=https://cloud.example.com
export IMMICHFRAME_OPENCLOUD_SPACE_NAME="Photo Frame"
export IMMICHFRAME_OPENCLOUD_USERNAME=frame
export IMMICHFRAME_OPENCLOUD_APP_PASSWORD=xxxxxxxx 
export IMMICHFRAME_AUTH_SECRET=my-shared-secret

go run ./cmd/opencloud-immichframe
```

## Command-line flags

| Flag | Notes |
| --- | --- |
| `-insecure` | Accept self-signed / invalid TLS certificates from the OpenCloud server. Overrides `OC_INSECURE`. Development only. |

```sh
go run ./cmd/opencloud-immichframe -insecure
```

## Docker

```sh
docker build -t opencloud-immichframe .
docker run --rm -p 8080:8080 \
  -e OC_URL=https://cloud.example.com \
  -e OPENCLOUD_IMMICHFRAME="Photo Frame" \
  -e OPENCLOUD_IMMICHFRAME=frame \
  -e OPENCLOUD_IMMICHFRAME_APP_PASSWORD=xxxxxxxx \
  -e IMMICHFRAME_AUTH_SECRET=my-shared-secret \
  opencloud-immichframe
```

## Docker Compose (official web UI + OpenCloud API via Traefik)

`docker-compose.yml` runs three services so the **official** ImmichFrame web UI
displays photos from your OpenCloud space:

```
client ─▶ traefik :8080
           ├─ PathPrefix(`/api`) ─▶ opencloud-immichframe   (OpenCloud photos)
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
- Set `IMMICHFRAME_OPENCLOUD_SPACE_NAME` (e.g. `Images`) rather than `IMMICHFRAME_OPENCLOUD_SPACE_ID` —
  space ids contain `$`, which Compose interpolates (double it as `$$` if you
  must use the id).
- If OpenCloud runs on the Docker host, keep
  `OC_URL=https://host.docker.internal:9200` (the compose file maps
  `host.docker.internal`). For a self-signed dev cert, `OC_INSECURE=true`.
- Rootless Docker: point Traefik at your runtime socket (see the comment in the
  compose file).
- The official image needs `ImmichServerUrl`/`ApiKey` set just to boot; the
  placeholders in `.env.example` are enough since its `/api` is bypassed.
