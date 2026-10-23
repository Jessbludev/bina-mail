# Snapshot v1 — Cimentación

## Fecha

2026-10-23

## Cambios desde el último Snapshot

- Monorepo: módulo Go en services/ (comandos gateway, auth, directory, mailbox y realtime con healthcheck) y aplicación web en apps/web (Vue 3, Pinia, Tailwind tipado en TypeScript, Worker criptográfico con contrato tipado).
- Contratos protobuf en packages/contracts (mail, directory, realtime) con buf lint como gate de CI.
- Proyecto Supabase bina-mail creado (us-east-1, ref pnxtkuzwhkfkfsbsagdf). Migración inicial aplicada: key_directory y envelopes con RLS activo y denegación por defecto para roles anónimo y autenticado.
- CI: gates de seguridad (gitleaks, actionlint, auditoría zero-log), documentación, Go y contratos activos. Gate web condicionado a la existencia de pnpm-lock.yaml (DEBT-001).

## Estado Actual

- Compilable y verificable: go test ./... y vitest en verde sobre los esqueletos.
- Sin lógica de dominio: llega en Fase 2.

## Métricas Clave

- Líneas de código: esqueleto (aprox. 500). Cobertura: solo health; no representativa del dominio aún.

## Próximos Pasos

- Fase 2: criptografía E2EE con vectores de test, autenticación, directorio de claves y buzón de sobres.
