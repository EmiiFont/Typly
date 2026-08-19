# Typly

Typly generates typing-animation media — Unicode text typed out character by
character with a blinking cursor, in the style of
[typed.js](https://github.com/mattboldt/typed.js/) — as an animated GIF,
an MP4 video, or a sequence of PNG frames.

Each sentence is typed out, the cursor blinks, the sentence is "deleted"
by playing the typing frames in reverse, and the next sentence begins.
The GIF loops forever.

## Example

![Typly color emoji example](docs/typly-example.gif)

## Install

```sh
go install github.com/EmiiFont/typly@latest
```

Or build from source:

```sh
git clone https://github.com/EmiiFont/typly
cd typly
go build
```

GIF and PNG output work everywhere with no dependencies. MP4 output
requires [ffmpeg](https://ffmpeg.org/) on your `PATH`.

The default build embeds Noto Sans and Noto Sans Symbols, plus the Twemoji
PNG set for color emoji such as `🌍`, `🚀`, and `🎉`. Use `-emoji mono` to
disable color assets and use font glyphs instead. A custom `-font` can be
provided for additional scripts or a different visual style.

TrueType collections (`.ttc`) are accepted as custom fonts too, which is
useful for CJK fonts such as Noto Sans CJK:

```sh
typly -text "世界" -font /path/to/NotoSansCJK-Regular.ttc
```

## Usage

```sh
typly -text "Hello there!;Use your own sentences!;Have a great day!"
```

Sentences are separated by `;`. By default this writes `typly.gif` to
the current directory.

More options:

```sh
typly -text "Dark mode;Looks nice" \
  -out ./out \
  -gif -mp4 -frames \
  -size 1280x720 \
  -fps 12 \
  -font-size 72 \
  -fg '#00FF87' -bg '#1E1E2E' \
  -cursor '_' \
  -blinks 3 \
  -font ./MyFont.ttf
```

| Flag | Default | Description |
|---|---|---|
| `-text` | (required) | Sentences to animate, separated by `;` |
| `-out` | `.` | Output directory |
| `-gif` | `true` | Write `typly.gif` |
| `-mp4` | `false` | Write `typly.mp4` (requires ffmpeg) |
| `-frames` | `false` | Write PNG frames to `<out>/frames` |
| `-size` | `1280x720` | Canvas size as `WIDTHxHEIGHT` |
| `-fps` | `12` | Typing frame rate |
| `-font` | embedded | Path to a TTF/OTF font (default: [Go Regular](https://go.dev/blog/go-fonts)) |
| `-font-size` | `72` | Font size in points |
| `-fg` | `#000000` | Text color (hex) |
| `-bg` | `#FFFFFF` | Background color (hex) |
| `-cursor` | `\|` | Cursor glyph |
| `-blinks` | `3` | Cursor blink cycles at the end of each sentence |
| `-emoji` | `color` | Emoji mode: `color` or `mono` |

## Development

```sh
go test ./...
```

The test suite includes an MP4 integration test that is skipped
automatically when ffmpeg is not installed.

## Product App

The reusable engine lives in `pkg/typly`. A first Go/Datastar server is
available with:

```sh
go run ./cmd/typlyd
```

It serves the editor at `http://localhost:8080` and exposes JSON GIF/MP4
render endpoints using the shared `RenderSpec` contract. The native Flutter
client is in `mobile/`:

```sh
cd mobile
flutter run --dart-define=TYPLY_SERVER_URL=http://localhost:8080
```

See [`docs/product-architecture.md`](docs/product-architecture.md) for the
web, mobile, and future Railway deployment boundaries. No deployment is
performed by this repository.

## History

Typly was originally a Windows-only C# / .NET Framework application
built on GDI+, AForge.NET, and NGif. It was rewritten in Go for
cross-platform support with a single static binary.
