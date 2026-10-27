# ADR-007 — Transporte interno HTTP/JSON con protobuf canonico

## Estado

Aceptada (2026-10-27)

## Contexto

Los servicios internos (gateway, auth, directory, mailbox, realtime) necesitan un transporte. gRPC exige grpc-go y generacion de stubs con buf generate, lo que anade dependencias y pasos de pipeline frente a un dominio todavia pequeno. El criterio del proyecto es compilacion rapida y determinista con stdlib siempre que no haya coste funcional.

## Opciones

1. gRPC ahora: acoplamiento temprano a buf generate y go.sum gestionado por CI.
2. HTTP/JSON con patrones ServeMux de Go 1.22 y protobuf como contrato canonico.
3. Monolito con acceso directo a la base de datos: rompe las fronteras de ADR-001.

## Decision

Opcion 2. Los servicios se comunican por HTTP/JSON interno; packages/contracts sigue siendo la unica fuente de verdad y los nombres de campo JSON replican el proto campo a campo. El modulo Go permanece sin dependencias externas.

## Consecuencias

- Sin generacion de codigo Go hasta el primer consumidor gRPC (DEBT-002).
- Serializacion menos compacta; irrelevante en el volumen del MVP.
- Revision en Fase 3: activar gRPC si la serializacion o el acoplamiento generan defectos medibles.
