# Snapshot v0 — Inception

## Fecha

2026-10-23

## Cambios desde el último Snapshot

- Inicial.

## Estado Actual

- Alcance decidido: mensajería interna E2EE entre usuarios de la plataforma, sin SMTP en el MVP (ADR-006).
- Despliegue objetivo: cloud gestionado (Supabase + Fly.io/Railway).
- Cache: Dragonfly (ADR-002). Colas: Redpanda (ADR-003). Criptografía: Argon2id + BLAKE3 + X25519 + XChaCha20-Poly1305 (ADR-004).
- Sin código de aplicación. Documentación, mapa, ADRs, CI y plantillas definidas.
- Biblia maestra del proyecto manejando internamente aplicando (visión, roadmap, amenazas, flujo de trabajo).

## Métricas Clave

- Líneas de código: 0. Cobertura: n/a.

## Próximos Pasos

- Fase 1 (cimentación): workspaces (pnpm + go.work), contratos protobuf, esquema inicial Supabase, esqueletos de servicios y web, CI activa.
