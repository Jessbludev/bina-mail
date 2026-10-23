# Bina Mail

Mini plataforma de correo seguro y privado: mensajería interna con cifrado de extremo a extremo (E2EE) entre usuarios de la plataforma. MVP sin SMTP externo (ADR-006).

## Criterios no negociables

- Arquitectura de conocimiento cero (zero-knowledge).
- Política de zero log (ADR-005).
- Confianza cero (zero-trust).
- Cifrado de extremo a extremo (ADR-004).
- Comportamiento determinista y testing claro.
- CI/CD con reglas estrictas.
- Documentación clara y precisa.

## Stack

| Capa | Tecnología |
|---|---|
| Web | TypeScript puro, Vue 3 (Composition API), Pinia, Tailwind con clases seguras y tipadas en TypeScript, Web Workers para criptografía |
| Borde | Go con Chi: HTTP/WS público, validación de input, BFF |
| Interno | gRPC entre servicios; contratos protobuf en packages/contracts |
| Datos | Supabase (PostgreSQL) con RLS |
| Cache | Dragonfly (ADR-002) |
| Eventos | Redpanda (ADR-003) |
| Criptografía | Argon2id, BLAKE3, X25519, XChaCha20-Poly1305 |

## Estructura

- apps/web — SPA (Vue 3, Pinia, Web Workers de criptografía)
- services — Backend Go por dominio: gateway, auth, directory, mailbox, realtime
- packages — Contratos protobuf, tipos compartidos, DTOs
- DECISIONS — ADRs
- docs — Arquitectura, criptografía, modelo de amenazas, flujo de trabajo
- MAP.md — Mapa de módulos (leer primero)
- SNAPSHOT_v0.md — Estado del hito actual
- DEBT.md — Deuda explícita y auditable
- LOG/ — Registro mensual de cambios

## Operación del repositorio

1. Lea MAP.md antes de navegar el código.
2. Todo cambio que afecte a más de un módulo requiere ADR previo en DECISIONS/.
3. Los PR usan la plantilla con riesgos, pruebas realizadas e impacto en UX.
4. La CI es bloqueante: nada se fusiona con gates rojos.
5. Cada hito cierra con snapshot y línea de log; la deuda se registra en DEBT.md.
