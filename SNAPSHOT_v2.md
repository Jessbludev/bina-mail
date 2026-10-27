# Snapshot v2 — Dominio core (Fase 2)

## Fecha

2026-10-27

## Cambios desde el ultimo Snapshot

- Backend de dominio en Go stdlib: verificacion JWT HS256, sesiones opacas de 30 minutos (solo el hash SHA-256 se persiste, con revocacion), directorio de claves con upsert y busqueda por fingerprint, buzon de sobres con limite de tamano y orden determinista, y gateway BFF con autenticacion por sesion e inyeccion de usuario (X-Bina-User).
- Transporte interno HTTP/JSON decidido en ADR-007; protobuf permanece canonico y gRPC queda diferido a revision en Fase 3 (DEBT-002).
- Migracion Supabase #2 aplicada: tabla sessions con RLS y envelopes.sender_fingerprint.
- Criptografia client-side completa en el Worker: Argon2id (t=3, m=64 MiB, p=4), XChaCha20-Poly1305, BLAKE3 (vector oficial de la entrada vacia en pruebas), X25519 efimero con HKDF-SHA256 para envolver la clave por mensaje.
- Pruebas: suites httptest por servicio contra stubs del borde PostgREST; vitest de primitivas cripto y de sobres (ida y vuelta, clave ajena, manipulacion detectada).
- CI: job lockfiles que genera y confirma pnpm-lock.yaml cuando falta (cierra DEBT-001); el gate web se activa en el flujo siguiente al commit del lockfile.

## Estado Actual

- Dominio E2EE completo con suites de prueba; la CI verifica compilacion, lint y pruebas en cada push. go.mod sigue sin dependencias externas.
- La interfaz de usuario aun no consume el dominio: llega en Fase 4.

## Metricas Clave

- 21 archivos Go nuevos con pruebas; 3 suites vitest de cripto; cobertura del dominio medida por la CI (gate Go activo).

## Proximos Pasos

- Fase 3: Redpanda, realtime WS, Dragonfly y endurecimiento del BFF.
