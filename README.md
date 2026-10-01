# ograph-gen

Open Graph image generator for [fossunited.org](https://fossunited.org). Takes query parameters, renders SVG templates, and returns PNG images.

## Dependencies

- **Go** 1.22+
- **librsvg** (`rsvg-convert`) for SVG-to-PNG conversion
- **Fonts**: Inter, FFF Forward

### Install dependencies

**Debian/Ubuntu:**
```sh
sudo apt install librsvg2-bin fonts-inter
```

**Fedora:**
```sh
sudo dnf install librsvg2-tools google-noto-sans-fonts
```

**macOS:**
```sh
brew install librsvg
```

**NixOS** (declarative):
```nix
environment.systemPackages = [ pkgs.librsvg ];
```

FFF Forward is a free pixel font — download from [fff.cmiscm.com](https://fff.cmiscm.com/fff_forward/) and install to your system fonts directory.

## Setup

```sh
git clone https://github.com/fossunited/ograph-gen.git
cd ograph-gen
go build -o ograph-gen .
./ograph-gen
```

Server starts on `:3333` by default (configurable in `config.json`).

### Using `just`

A [`justfile`](justfile) wraps the common commands:

```sh
just build       # go build (local Go toolchain + rsvg-convert on PATH)
just run         # build and run locally
just build-nix   # build via Nix magic (default.nix)
just run-nix     # build via Nix and run straight from the store
just test <route> key=value ...   # hit /gen/<route>, save the PNG to /tmp, open it
just clean       # remove local build artifacts
```

`just test` reuses an already-running server on `:3333` if there is one, otherwise it builds and starts one just for that call. Example:

```sh
just test submission talk_title=Building+FOSS speaker_name=Jane speaker_image=/files/photo.webp event_chapter=INDIAFOSS event_name=IndiaFOSS+2026
```

### Building with Nix

`default.nix` builds the binary with `buildGoModule` and wraps it so `rsvg-convert` (from nixpkgs' `librsvg`) is always on `PATH`, with no system package install needed:

```sh
nix-build
./result/bin/ograph-gen
```

## Configuration

`config.json`:
```json
{
  "host": ":3333",
  "datadir": "./data",
  "routes": ["events", "submission", "cfp", "proposals", "schedule", "rsvp", "profile", "events_new"]
}
```

- `host` — listen address (default `:3333`)
- `datadir` — directory containing SVG templates
- `routes` — each entry maps to `datadir/<name>.svg` and creates a `/gen/<name>` endpoint

## Routes

All routes are under `/gen/` and accept query parameters that fill template variables in the SVG.

| Route | Parameters |
|-------|-----------|
| `/gen/events` | `event_name`, `event_type`, `event_date`, `event_chapter`, `event_location` |
| `/gen/events_new` | `event_name`, `event_type`, `event_date`, `event_chapter`, `event_location` |
| `/gen/submission` | `talk_title`, `session_type`, `speaker_name`, `speaker_designation`, `speaker_image`, `event_chapter`, `event_name` |
| `/gen/cfp` | `event_chapter`, `event_name` |
| `/gen/proposals` | `event_chapter`, `event_name` |
| `/gen/schedule` | `event_chapter`, `event_name` |
| `/gen/rsvp` | `event_chapter`, `event_name` |
| `/gen/profile` | `profile_name`, `designation`, `username`, `profile_image` |

Parameters ending in `_image` are fetched from `fossunited.org`, converted to PNG, and embedded as base64 data URIs. If the image is missing or invalid, a default avatar is used.

### Example

```
http://localhost:3333/gen/submission?talk_title=Building+FOSS&speaker_name=Jane&speaker_image=/files/photo.webp&event_chapter=INDIAFOSS&event_name=IndiaFOSS+2026
```

## Running as a systemd service

A unit file is provided at `ograph-gen.service`. It expects the binary, `config.json`, and `data/` to live in `/opt/ograph-gen` and to run as a dedicated `ograph-gen` user.

```sh
sudo mkdir -p /opt/ograph-gen
sudo cp ograph-gen config.json ograph-gen.service /opt/ograph-gen/
sudo cp -r data /opt/ograph-gen/
sudo chown -R ograph-gen:ograph-gen /opt/ograph-gen

sudo cp /opt/ograph-gen/ograph-gen.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now ograph-gen
```

Check status and logs:

```sh
systemctl status ograph-gen
journalctl -u ograph-gen -f
```

If you install to a different path, or run as a different user, edit `WorkingDirectory`, `ExecStart`, `User`, and `Group` in `ograph-gen.service` accordingly.

## Adding a new template

1. Export the SVG from Figma at 1200x630
2. Replace dynamic text with Go template variables: `{{ .variable_name }}`
3. For image slots, use a parameter name ending in `_image` (e.g. `speaker_image`)
4. Save to `data/<name>.svg`
5. Add `"<name>"` to the `routes` array in `config.json`

## License

AGPL-3.0 — see [LICENSE.txt](LICENSE.txt)
