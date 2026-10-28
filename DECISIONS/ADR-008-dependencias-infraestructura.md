# ADR-008: Dependencias de infraestructura (franz-go y gorilla/websocket)

- Estado: Aceptada
- Fecha: 2026-10-28

## Contexto

El modulo Go mantenia cero dependencias externas desde Fase 1. La Fase 3 exige dos clientes de infraestructura maduros: Redpanda (API de Kafka) para los eventos de buzon y WebSocket (RFC 6455) para las senales en tiempo real. Implementar ambos protocolos a mano en stdlib excederia el alcance del MVP y concentraria el riesgo en codigo critico sin la madurez de una biblioteca auditada por la comunidad.

## Opciones

1. Mantener cero dependencias e implementar los protocolos desde cero. Coste desproporcionado y riesgo alto en codigo critico de red.
2. Reabrir ADR-007 con gRPC y proxies de mensajeria. Rechazado en la revision de Fase 3: el transporte HTTP/JSON opera sin defectos y gRPC permanece diferido (DEBT-002).
3. Incorporar dependencias puntuales, auditadas y confinadas a adapters con superficie minima.

## Decision

- franz-go v1.22.1 como cliente de Redpanda, encapsulado en services/internal/kafka; el resto del modulo no lo importa.
- gorilla/websocket v1.5.3 para el WS de realtime y el proxy del gateway, encapsulado en services/internal/realtimeapi y services/internal/gateway.
- Go del modulo sube a 1.26, requisito declarado por franz-go.
- kfake v0 (broker en proceso) queda como dependencia de pruebas del adapter kafka, con la unica prueba de extremo a extremo produce y consume.
- go.sum no se genera localmente: el job lockfiles de la CI ejecuta go mod tidy y confirma go.sum al repositorio (DEBT-003).
- govulncheck corre sobre el grafo completo de dependencias en cada push activo.

## Consecuencias

- Solo los adapters mencionados importan paquetes externos; el dominio conserva su superficie stdlib.
- Existe una ventana de un flujo entre un cambio de go.mod y la confirmacion de go.sum por la CI; el job go queda gated a la existencia de go.sum durante esa ventana.
- Revision en Fase 5 (despliegue): mTLS entre servicios y sustitucion inmediata de cualquier dependencia con CVE critica sin parche.
