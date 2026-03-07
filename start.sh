#!/bin/sh
set -e

if [ ! -f ".env" ]; then
  cp .env.example .env
fi

# Frontend build caching notes:
# - Previously we only built when web/dist/index.html was missing.
# - That can go stale when web/src changes (dist still exists).
# - We rebuild when any tracked frontend input is newer than dist/index.html.
DIST_HTML="web/dist/index.html"

need_build_frontend() {
  [ ! -f "$DIST_HTML" ] && return 0

  # If any of these inputs are newer than the dist entrypoint, rebuild.
  # Keep the list conservative but correct.
  for p in \
    web/src \
    web/public \
    web/scripts \
    web/package.json \
    web/package-lock.json \
    web/npm-shrinkwrap.json \
    web/vite.config.* \
    web/tsconfig*.json \
    web/postcss.config.* \
    web/tailwind.config.*
  do
    # Skip non-existing globs/paths.
    [ -e $p ] || continue

    if [ -d $p ]; then
      if find $p -type f -newer "$DIST_HTML" -print -quit 2>/dev/null | grep -q .; then
        return 0
      fi
    else
      if [ $p -nt "$DIST_HTML" ]; then
        return 0
      fi
    fi
  done

  return 1
}

if need_build_frontend; then
  echo "Building frontend..."
  (
    cd web
    npm ci
    npm run build
  )
fi

echo "Building backend..."
go build -o apihub ./cmd/server

echo "ApiHub listening on http://localhost:9011"
./apihub
