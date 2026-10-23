# packages/contracts — Contratos Bina Mail

Única frontera de contrato entre servicios y cliente (ADR-001).

- proto/bina/mail/v1 — sobres y servicio de buzón.
- proto/bina/directory/v1 — directorio de claves públicas.
- proto/bina/realtime/v1 — señales WS sin contenido.

Comandos: buf lint, buf breaking (contra main cuando existan ramas), buf generate (buf.gen.yaml).

La generación de código Go queda registrada como DEBT-002 hasta el primer consumidor gRPC en Fase 2.
