# SPEC: 153 — INBOUND_MULTI_CERTIFICATE_SNI

| Field | Value |
|-------|-------|
| Type | F — multiple certificates for a TLS inbound with explicit fallback by list order |
| Status | O (open) |
| Branch | `lx` |
| Build tag | — |

## 1. Problem

An inbound accepts only one server certificate/key pair. A certificate covering several unrelated hostnames exposes all those names regardless of the requested SNI. ShadowVPN2 needs to provide separate single-name certificates to one inbound and let TLS choose the matching certificate.

## 2. Contract

`tls.certificates` is an ordered list of certificate entries. Each entry contains exactly one certificate source and one matching key source, either inline (`certificate`, `key`) or file-based (`certificate_path`, `key_path`). The first list entry is the default certificate for absent SNI, unknown SNI, and client-hello algorithms for which no later certificate is compatible. For a known SNI, Go `crypto/tls` selects a compatible certificate whose DNS SAN matches the name.

 Singular inbound certificate fields remain accepted by the loader for compatibility but are omitted from the generated schema; ShadowVPN2 emits only the new `certificates` list. Singular and plural settings cannot be combined. An empty list or an entry with incomplete/mixed sources is an error.

Certificate file paths are watched. A change reloads the complete ordered set atomically; malformed replacement leaves the active set unchanged. QUIC and TCP use the same `tls.Config` selection behavior. Reality remains incompatible with ordinary certificates.

## 3. Fallback semantics

 No separate fallback name is needed. The first configured certificate is the explicit fallback. ShadowVPN2 puts the node-domain certificate first. For unknown or absent SNI Go `crypto/tls` selects the first certificate compatible with the client's signature algorithms.

## 4. ShadowVPN2 contract

 `InboundTlsConfig` serializes `certificates` as ordered certificate/key objects. Contributors put the node-domain certificate first, then remaining managed certificates in stable domain order, de-duplicated by certificate identity. If no managed certificate exists, the configured self-signed protocol certificate is the sole entry. `server_name` continues to identify the protocol's expected domain and is not used for server certificate selection.

Managed certificate storage and ACME issuance are per domain. The old per-node multi-domain entity is replaced by one entity per node/domain. Existing stored multi-domain certificates are not served after this change; renewal issues one certificate per desired domain. Certificate reuse is allowed only when the source certificate covers exactly the target domain.

## 5. Tests and acceptance

- Two distinct certificates: matching SNI gets matching leaf; unknown and absent SNI get the first certificate.
- A client offering incompatible signature algorithms receives the first compatible certificate according to Go TLS behavior.
- Invalid or incomplete certificate entry fails configuration creation.
- File reload updates the selected set atomically; an invalid update does not replace the active certificates.
- Naive TCP and Hysteria2 QUIC inbounds receive the same ordered certificate set.
- ShadowVPN2 ACME requests one domain per order, persists each domain independently, and generated config places node-domain first and does not emit a multi-SAN legacy certificate.
- `sing-box check` accepts generated config; TLS loopback smoke verifies known SNI, unknown SNI, and no-SNI certificate selection.

## 6. Scope

 Singular inbound certificate fields remain accepted by the loader for compatibility but are omitted from the generated schema. No new certificate-provider type. No changes to client TLS, client certificates, ECH or Reality semantics. Choosing the node-domain fallback makes fallback deterministic and avoids returning a certificate listing all configured domains; it does not claim to reduce active-probing detectability.
