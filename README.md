# opencloud-immichframe

An [ImmichFrame](https://github.com/immichFrame/ImmichFrame)-compatible HTTP API
backend written in Go, serving photos from an **OpenCloud space** instead of an
Immich server. Point an existing ImmichFrame client at this service and it works
unchanged.

It talks to OpenCloud over the public HTTP APIs only — **LibreGraph** to list the
space, **WebDAV** (`/dav/spaces/{spaceId}/…`) to stream image bytes. No reva/gRPC
or monorepo coupling.

Please find details in [the project documentation](https://dragotin.github.io/opencloud-immichframe).


## Development

```sh
go build ./...
go vet ./...
go test ./...
```

This project is developed by Klaas Freitag with the assistance of
[Claude](https://www.anthropic.com/claude). Thanks Domme for help
and motivation :-) Contributions are welcome!

