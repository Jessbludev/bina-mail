# services/ — Backend Go

Servicios por dominio (ver MAP.md): gateway (Chi, BFF, borde público), auth, directory, mailbox y realtime.

## Reglas

- Comunicación interna solo por gRPC con contratos protobuf (packages/contracts).
- Eventos solo por Redpanda; consumidores idempotentes.
- Ningún endpoint sin autenticación explícita; "internal" no es una defensa.
- Input validado al borde; secretos por entorno, nunca en el repo.
- Logs conforme ADR-005 (zero-log).
- Cada servicio despliega como contenedor independiente con healthcheck.
