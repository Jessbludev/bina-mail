# packages/contracts — Contratos Bina Mail

Unica frontera de contrato entre servicios y cliente (ADR-001).

- proto/bina/mail/v1 — sobres y servicio de buzon.
- proto/bina/directory/v1 — directorio de claves publicas.
- proto/bina/realtime/v1 — senales WS sin contenido.

Comandos: buf lint, buf breaking (contra main cuando existan ramas), buf generate (buf.gen.yaml).

## Mapeo JSON (ADR-007)

El transporte interno actual es HTTP/JSON. Los campos del sobre JSON replican el mensaje proto Envelope campo a campo: recipient_fingerprint, nonce, ciphertext, wraps (recipient_key_id, ephemeral_public_key, wrapped_key), blake3_fingerprint y sender_fingerprint opcional. Cualquier cambio de campo se hace primero en proto y luego en los codificadores JSON de Go y del Worker.

La generacion de codigo Go queda registrada como DEBT-002 hasta el primer consumidor gRPC (revision en Fase 3).
