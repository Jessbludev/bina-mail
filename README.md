# Bina Mail

Mini plataforma de correo seguro y privado: mensajeria interna con cifrado de extremo a extremo (E2EE) entre usuarios de la plataforma. MVP sin SMTP externo (ADR-006).

## Criterios no negociables

- Arquitectura de conocimiento cero (zero-knowledge).
- Politica de zero log (ADR-005).
- Confianza cero (zero-trust).
- Cifrado de extremo a extremo (ADR-004).
- Comportamiento determinista y testing claro.
- CI/CD con reglas estrictas.
- Documentacion clara y precisa.

## Stack

| Capa | Tecnologia |
|---|---|
| Web | TypeScript puro, Vue 3 (Composition API), Pinia, Tailwind con clases seguras y tipadas en TypeScript, Web Workers para criptografia |
| Borde | Go con net/http (stdlib): HTTP/WS publico, validacion de input, BFF |
| Interno | HTTP/JSON entre servicios (ADR-007); contratos protobuf en packages/contracts (gRPC diferido, DEBT-002) |
| Datos | Supabase (PostgreSQL) con RLS |
| Cache | Dragonfly (ADR-002) |
| Eventos | Redpanda (ADR-003) |
| Criptografia | Argon2id, BLAKE3, X25519, XChaCha20-Poly1305, Ed25519 |

## Estructura

- apps/web — SPA (Vue 3, Pinia, Web Workers de criptografia)
- services — Backend Go por dominio: gateway, auth, directory, mailbox, realtime
- packages — Contratos protobuf, tipos compartidos, DTOs
- DECISIONS — ADRs
- docs — Arquitectura, criptografia, modelo de amenazas, flujo de trabajo
- MAP.md — Mapa de modulos (leer primero)
- SNAPSHOT_v4.md — Estado del hito actual (cada hito publica uno nuevo)
- DEBT.md — Deuda explicita y auditable
- LOG/ — Registro mensual de cambios
- SECURITY.md — Politica de seguridad
- CONTRIBUTING.md — Como contribuir
- CODE_OF_CONDUCT.md — Codigo de conducta

## Seguridad

La CI ejecuta gitleaks, actionlint, CodeQL, govulncheck y Dependabot (npm, go, actions) en cada cambio. Los reportes de vulnerabilidades siguen el procedimiento privado de SECURITY.md; nunca se abren issues publicos con detalles de seguridad.

## Operacion del repositorio

1. Lea MAP.md antes de navegar el codigo.
2. Todo cambio que afecte a mas de un modulo requiere ADR previo en DECISIONS/.
3. Los PR usan la plantilla con riesgos, pruebas realizadas e impacto en UX; el detalle operativo esta en CONTRIBUTING.md.
4. La CI es bloqueante: nada se fusiona con gates rojos.
5. Cada hito cierra con snapshot y linea de log; la deuda se registra en DEBT.md.

## Summary (English)

Bina Mail is a private, secure mini email platform: internal E2EE messaging between platform users, no external SMTP in the MVP. Non-negotiables: zero-knowledge, zero-log, zero-trust, E2EE, determinism, strict CI/CD. Web: TypeScript, Vue 3, Pinia, typed Tailwind tokens, crypto in Web Workers. Backend: Go with net/http at the edge (BFF), HTTP/JSON between services (protobuf contracts; gRPC deferred), Supabase with RLS, Dragonfly cache, Redpanda events. CI runs gitleaks, actionlint, CodeQL, govulncheck and Dependabot. Read MAP.md first; cross-module changes need an ADR; security reports go through SECURITY.md.
