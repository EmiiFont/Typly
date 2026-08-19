# Typly Product Architecture

## Decisions

- Web UI: server-rendered Go HTML with Datastar interactions.
- Mobile UI: native Flutter application, not a web wrapper.
- Rendering: GIF and preview can move client-side later; MP4 is server-side
  because it requires ffmpeg.
- Deployment target: Railway-compatible container, but deployment is manual
  and intentionally not part of this repository change.
- Templates: standard-library `html/template`.
- Authentication, billing, and accounts: deferred until the editor workflow
  is validated.

## Current boundaries

`pkg/typly` is the reusable rendering engine. `RenderSpec` is the JSON
contract shared by the CLI, HTTP API, and Flutter client. It contains only
user-facing rendering choices and no filesystem paths or font bytes.

`cmd/typlyd` provides:

- `GET /` — initial Datastar editor shell
- `GET /api/spec/default` — default `RenderSpec`
- `POST /api/spec/validate` — Datastar signal validation example
- `POST /api/render/gif` — JSON spec to GIF bytes
- `POST /api/render/mp4` — JSON spec to MP4 bytes; requires ffmpeg

The current MP4 endpoint is intentionally synchronous for the first vertical
slice. The next server step is a bounded job queue with job IDs, cancellation,
TTL cleanup, and Datastar SSE progress updates.

## Native app

`mobile/` is a Flutter native client. It edits the same `RenderSpec` fields
and sends JSON to the HTTP API. Set the server URL at build time:

```sh
flutter run --dart-define=TYPLY_SERVER_URL=http://localhost:8080
```

The Android emulator default is `http://10.0.2.2:8080`.

## Next vertical slices

1. Add mobile file saving/share-sheet integration for returned GIF/MP4 bytes.
2. Add a real preview surface driven by `RenderSpec`.
3. Add bounded asynchronous MP4 jobs and Datastar progress SSE.
4. Add Railway Docker configuration without deploying it.
5. Add persistent projects and authentication only after the anonymous flow
   is usable.
