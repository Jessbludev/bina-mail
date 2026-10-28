# DEBT — Deuda explicita y auditable

Formato: ID | Descripcion | Modulo | Dueno | Fecha de revision | Estado

| ID | Descripcion | Modulo | Dueno | Fecha de revision | Estado |
|---|---|---|---|---|---|
| DEBT-001 | Cerrada por automatizacion: el job lockfiles de la CI genera y confirma pnpm-lock.yaml tras cada cambio de dependencias | apps/web | web | Cerrada por CI | Cerrada |
| DEBT-002 | gRPC diferido por ADR-007: transporte interno HTTP/JSON con protobuf canonico; revision de Fase 3 sin defectos, se mantiene hasta el primer consumidor gRPC | packages/contracts | backend | Inicio de Fase 5 | Abierta |
| DEBT-003 | Cerrada por automatizacion: el job lockfiles de la CI genera y confirma go.sum tras cada cambio de dependencias Go | services | backend | Cerrada por CI | Cerrada |
| DEBT-004 | El upsert de clave invalida solo el indice por usuario en cache; la entrada anterior por fingerprint caduca por TTL de 60 segundos | services/internal/directoryapi | backend | Inicio de Fase 5 | Abierta |
| DEBT-005 | Prueba E2E de navegador real (Playwright, dos usuarios, WS y accesibilidad con lector de pantalla) queda para Fase 5 junto al despliegue; la verificacion de Fase 4 es el flujo completo entre dos usuarios con el Worker real en vitest | apps/web | web | Inicio de Fase 5 | Abierta |

Toda deuda (tecnica, de seguridad u observabilidad) se registra aqui en el momento en que se decide dejar algo sin hacer a proposito, con dueno y fecha de revision.
