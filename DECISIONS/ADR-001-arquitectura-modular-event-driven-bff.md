# ADR-001: Arquitectura modular event-driven con BFF

## Contexto

Bina Mail requiere modularidad, eventos internos y separación de responsabilidades entre el borde público y la lógica de dominio, con criterios zero-trust y despliegue en cloud gestionado.

## Opciones Consideradas

1. Monolito modular (un solo despliegue).
   - Pros: operación mínima; sin red interna; la forma más simple de empezar.
   - Contras: escala uniforme; un fallo afecta a todo; fronteras solo disciplinarias.
   - Coste: bajo al inicio, alto en crecimiento.
2. Microservicios completos (por entidad).
   - Pros: escala fina por servicio.
   - Contras: sobrecarga operativa prematura para un MVP; distribución antes de necesitarla.
   - Coste: alto.
3. Módulos desplegables con BFF y bus de eventos: gateway Chi en el borde + servicios por dominio vía gRPC + eventos en Redpanda.
   - Pros: fronteras reales; escala por responsabilidad; event-driven; el BFF aísla al cliente del dominio.
   - Contras: algo más de operación que un monolito.
   - Coste: medio.

## Decisión

Opción 3, con el mínimo de servicios del mapa (MAP.md): gateway (BFF), auth, directory, mailbox y realtime. gRPC entre servicios; eventos por Redpanda; el cliente web solo habla con el gateway.

## Consecuencias

- Cada módulo tiene una razón para cambiar; las fronteras se declaran en MAP.md.
- Los contratos viven en packages/contracts (protobuf); cambios de contrato en PR dedicado con buf breaking en CI.
- WebRTC queda fuera del MVP; la señal realtime es WebSocket (ADR-006).

## Condición de revisión

Si el coste operativo o el p99 del borde exceden el SLO de forma sostenida, consolidar servicios en un monolito modular mediante Strangler Fig.
