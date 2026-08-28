# dev/build-docs.sh
#!/usr/bin/env bash
set -euo pipefail
REF="$(grep -oE 'dschmidt/opencloud-service-docs-action@[^ "]+' \
         .github/workflows/docs.yml | head -1 | cut -d@ -f2)"
DIR=".cache/service-docs/action"
if [ "$(git -C "$DIR" rev-parse HEAD 2>/dev/null)" != "$REF" ]; then
  rm -rf "$DIR"
  git clone https://github.com/dschmidt/opencloud-service-docs-action.git "$DIR"
  git -C "$DIR" checkout "$REF"
fi
exec bash "$DIR/build.sh"

