# Implementation report: 152 — RUNTIME_SERVER_MANAGEMENT

## Completed in this slice

- WG/AWG runtime peer updates are differential: unchanged peers are left intact, removed peers use `remove=true`, and new or changed peers are sent by `public_key`.
- `replace_peers=true` is not used by runtime updates, so unchanged peer sessions are not intentionally torn down.
- Added regression coverage for unchanged, removed, changed, and added peers.
- Published in `IlluminedProject/sing-box-lx` on branch `lx`.

## Verification

- `go test ./transport/wireguard ./protocol/wireguard` passed.
- `go vet ./transport/wireguard ./protocol/wireguard` passed.
- `dotnet build ShadowVPN2/ShadowVPN2.csproj --no-restore` passed.
- `dotnet test ShadowVPN2.Tests/ShadowVPN2.Tests.csproj --no-restore`: 434 passed.
- `go test -race` was not runnable in the environment because CGO is disabled.

The Clash REST API, full runtime feature, and end-to-end smoke scenario remain pending.
