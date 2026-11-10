# PLAN: 153 — INBOUND_MULTI_CERTIFICATE_SNI

Contract: [SPEC.md](SPEC.md), Russian copy: [SPEC.ru.md](SPEC.ru.md).

## Stage 1. sing-box TLS

 1. [done] Add the ordered `tls.certificates` option and entry type; retain legacy singular fields only as schema-omitted loader compatibility.
 2. [done] Load all inline/path entries, validate complete pairs, and build `tls.Config.Certificates` in order.
 3. [done] Watch all certificate/key paths and reload the complete set atomically.
 4. [done] Extend `cmd_merge` path merging and generated schema.
 5. [done] Add TLS selection, validation, and reload tests.

## Stage 2. ShadowVPN2 certificate lifecycle

1. Persist one managed certificate per node/domain; update lookup, reuse and renewal.
2. Issue each desired domain independently and record failures per domain without suppressing other renewals.
3. Emit the configured protocol domain first, then remaining domain certificates, into Naive and Hysteria2 inbound configs.
4. Make shared-port Kestrel fallback use the node-domain certificate, never a multi-domain certificate.
5. Add focused storage, issuance-selection, serialization, and contributor tests.

## Stage 3. Verification

1. `go test` relevant packages, config schema/check, loopback TLS certificate-selection smoke.
2. `dotnet test` focused ShadowVPN2 tests and a generated-config smoke validating per-name certs and first-entry fallback.
