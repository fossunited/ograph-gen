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

## Adding a new template

1. Export the SVG from Figma at 1200x630
2. Replace dynamic text with Go template variables: `{{ .variable_name }}`
3. For image slots, use a parameter name ending in `_image` (e.g. `speaker_image`)
4. Save to `data/<name>.svg`
5. Add `"<name>"` to the `routes` array in `config.json`

## License

AGPL-3.0 — see [LICENSE.txt](LICENSE.txt)
