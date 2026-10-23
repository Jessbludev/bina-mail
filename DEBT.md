# DEBT — Deuda explícita y auditable

Formato: ID | Descripción | Módulo | Dueño | Fecha de revisión | Estado

| ID | Descripción | Módulo | Dueño | Fecha de revisión | Estado |
|---|---|---|---|---|---|
| DEBT-001 | apps/web sin pnpm-lock.yaml: el gate web de la CI se activa cuando exista el lockfile (condición explícita en ci.yml). Generarlo con pnpm install en entorno con red y subirlo en su PR | apps/web | web | 2026-11-05 | Abierta |
| DEBT-002 | Código Go generado desde packages/contracts (buf generate) pendiente hasta el primer consumidor gRPC en Fase 2 | packages/contracts | backend | Inicio de Fase 2 | Abierta |

Toda deuda (técnica, de seguridad u observabilidad) se registra aquí en el momento en que se decide dejar algo sin hacer a propósito, con dueño y fecha de revisión.
