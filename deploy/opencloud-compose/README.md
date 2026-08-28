# Integrating with `opencloud-compose`

Drop-in fragments that add opencloud-immichframe to an
[opencloud-compose](https://github.com/opencloud-eu/opencloud-compose)
deployment, so the photo frame runs alongside OpenCloud on the same network and
behind the same Traefik.

This is **not** the same as the `docker-compose.yml` in the repository root.
That one is a self-contained stack — it builds the image locally, brings its own
Traefik and reaches OpenCloud over `host.docker.internal`. These fragments bring
no proxy and no build; they attach to the stack opencloud-compose already runs.

## Layout

The directory mirrors opencloud-compose's own, so the files can be copied
straight across:

```
immichframe/immichframe.yml       the service (+ web-UI init container)
traefik/immichframe.yml           expose on its own subdomain, bundled Traefik
external-proxy/immichframe.yml    expose on localhost for an external proxy
```

Pick **one** exposure fragment, matching how the rest of your stack is exposed.

## Prerequisite: a published image

opencloud-compose never builds — every service is a released image. So this
needs `opencloud-immichframe` in a registry. Until that is set up, build and tag
it locally and Compose will use it:

```sh
# in the opencloud-immichframe checkout
docker build -t ghcr.io/dragotin/opencloud-immichframe:dev .
# then in your .env: IMMICHFRAME_DOCKER_TAG=dev
```

Override `IMMICHFRAME_DOCKER_IMAGE` if you publish somewhere else.

## Install

```sh
cp -r deploy/opencloud-compose/* /path/to/opencloud-compose/
```

Then in the opencloud-compose `.env`, set `COMPOSE_FILE` — bundled Traefik:

```
COMPOSE_FILE=docker-compose.yml:immichframe/immichframe.yml:traefik/opencloud.yml:traefik/immichframe.yml
```

…or an external reverse proxy:

```
COMPOSE_FILE=docker-compose.yml:immichframe/immichframe.yml:external-proxy/opencloud.yml:external-proxy/immichframe.yml
```

`traefik/immichframe.yml` requires `traefik/opencloud.yml` too — the
`hsts-header` middleware it references is defined there.

And add the settings:

```sh
# Where the frame is reachable. Needs a DNS record pointing at this host.
IMMICHFRAME_DOMAIN=frame.opencloud.test
# The OpenCloud space to serve. Use the name, not the id: ids contain `$`,
# which Compose interpolates.
IMMICHFRAME_SPACE_NAME=Images
# An OpenCloud user and one of its app tokens (see below).
IMMICHFRAME_USERNAME=frame
IMMICHFRAME_APP_PASSWORD=
# Shared secret the frame clients must send as `Authorization: Bearer <secret>`.
# Do not leave this empty on a public subdomain.
IMMICHFRAME_AUTH_SECRET=

# Optional
#IMMICHFRAME_DOCKER_IMAGE=ghcr.io/dragotin/opencloud-immichframe
#IMMICHFRAME_DOCKER_TAG=latest
#IMMICHFRAME_WEBUI_TAG=v1.0.37.0
#IMMICHFRAME_CATALOG_REFRESH=5m
#IMMICHFRAME_INTERVAL=8
#IMMICHFRAME_PORT=8080          # external-proxy only
```

Then `docker compose up -d`.

## On the OpenCloud side

Two things, both from the web UI — **no configuration changes are needed**:

1. A space holding the photos, matching `IMMICHFRAME_SPACE_NAME`.
2. An [app token](https://docs.opencloud.eu/docs/user/admin/app-tokens/) for the
   user in `IMMICHFRAME_USERNAME`, pasted into `IMMICHFRAME_APP_PASSWORD`.

App-token auth works out of the box: the `auth-app` service starts
automatically and `PROXY_ENABLE_APP_AUTH` defaults to true. Unlike the Radicale
fragment, this one never touches the `opencloud` service.

The token grants access to everything that user can see, so give the frame its
own account and share only the photo space with it.

## Design notes

**It talks to OpenCloud internally.** `IMMICHFRAME_OPENCLOUD_URL` is
`http://opencloud:9200`, straight across the compose network. The base stack
runs the proxy with `PROXY_TLS=false`, so there is no certificate to trust and
no `OC_INSECURE` needed. The traffic never leaves the Docker network.

**It deliberately does not use `OC_URL`.** That name belongs to the `opencloud`
service in this stack, and both live in one shared `.env`. The service reads
`OC_URL;IMMICHFRAME_OPENCLOUD_URL` and the *last* name set wins, so the
`IMMICHFRAME_`-prefixed form always takes precedence.

**Own subdomain, not a path under the OpenCloud domain.** Routing it through
OpenCloud's proxy with `additional_policies` (as Radicale does for `/caldav/`)
would avoid a second certificate, but does not work here: the ImmichFrame web UI
is built for base path `/` and its assets break under a sub-path; the clients
want one origin serving both `/` and `/api`; and that route model forwards a
per-user identity, while this service is single-tenant — one app token for one
account.

**Expect restarts on a cold start.** The space is resolved once at startup and
the process exits if OpenCloud is not up yet, so on first boot it restarts until
OpenCloud finishes `opencloud init`. That is expected, not a fault.

**Only `/api/*` is authenticated.** The static UI assets are served without
auth; the photos behind them are not.

## Upstreaming

To propose these to opencloud-compose, the two-service shape should collapse to
one. No fragment upstream uses an init container; that pattern only exists here
because the official ImmichFrame image cannot run headless and its web UI has to
come from somewhere. Baking the UI into the published image removes the init
container and the shared volume entirely:

```dockerfile
COPY --from=ghcr.io/immichframe/immichframe:v1.0.37.0 /app/wwwroot /srv/web
ENV IMMICHFRAME_WEB_ROOT=/srv/web
```

`immichframe/immichframe.yml` then drops `immichframe-webui`, the `volumes:`
block, `depends_on` and `IMMICHFRAME_WEB_ROOT`, leaving a single service in
the same shape as `radicale/radicale.yml`.
