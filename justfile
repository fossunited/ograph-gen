set shell := ["bash", "-euo", "pipefail", "-c"]

default:
    @just --list

# Build with the local Go toolchain (requires go + rsvg-convert on PATH; see README).
# Targets linux/amd64 (deploy target server), which also covers this dev machine.
build:
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ograph-gen .

# Build with Nix (pins Go and vendors deps, wraps rsvg-convert in automatically).
build-nix:
    nix-build --no-out-link
    cp -f $(nix-build --no-out-link)/bin/ograph-gen ./ograph-gen

# Build locally and run.
run: build
    ./ograph-gen

# Build with Nix and run straight out of the store (no local Go/librsvg needed).
run-nix:
    $(nix-build --no-out-link)/bin/ograph-gen

# Remove local build artifacts.
clean:
    rm -f ograph-gen
    rm -rf result

# (private) Build with whichever toolchain is available: prefer local go, fall back to nix.
_build-auto:
    #!/usr/bin/env bash
    set -euo pipefail
    if command -v go >/dev/null 2>&1; then
        just build
    elif command -v nix >/dev/null 2>&1; then
        just build-nix
    else
        echo "error: need 'go' or 'nix' installed to build ograph-gen" >&2
        exit 1
    fi

# Make sure ./ograph-gen exists, building only if it's missing (fast: skips the
# build entirely on repeat calls). Run `just clean` first to force a rebuild.
ensure:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -x ./ograph-gen ]; then
        echo "ograph-gen already built, reusing it"
    else
        just _build-auto
    fi

# (private) True if something already answers on :3333.
_up:
    @curl -sf -o /dev/null http://localhost:3333/ping

# Hit /gen/<route> (e.g. `just test cfp event_chapter=INDIAFOSS event_name=Test`),
# save the PNG to /tmp, and open it. Reuses an already-running server on :3333 if
# there is one; otherwise reuses (or builds, see `ensure`) ./ograph-gen and starts
# a server just for this call.
test route *args:
    #!/usr/bin/env bash
    set -euo pipefail

    started=0
    if ! just _up 2>/dev/null; then
        just ensure
        ./ograph-gen >/tmp/ograph-gen-test.log 2>&1 &
        pid=$!
        started=1
        for _ in $(seq 1 50); do
            just _up 2>/dev/null && break
            sleep 0.1
        done
    fi

    query="$(echo "{{args}}" | tr ' ' '&')"
    out="/tmp/ograph-gen-{{route}}-$(date +%s).png"

    curl -sf "http://localhost:3333/gen/{{route}}?${query}" -o "$out"
    echo "Saved $out"

    if [ "$started" = "1" ]; then
        kill "$pid" 2>/dev/null || true
    fi

    xdg-open "$out" >/dev/null 2>&1 || open "$out" >/dev/null 2>&1 || true
