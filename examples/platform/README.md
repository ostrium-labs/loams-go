# platform

Opt-in, read-only listing of Live apps through `client.LiveAdmin()`. The Go SDK 0.1.2 does not expose auth or general admin modules, and House has no SDK module in this release.

```console
LOAMS_PLATFORM_EXAMPLE=1 go run ./examples/platform
```

Requires an authorized `LOAMS_API_KEY` or `LOAMS_TOKEN`; Live apps are listed only when the server advertises Live.
