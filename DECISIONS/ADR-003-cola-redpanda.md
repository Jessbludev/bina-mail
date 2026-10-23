# ADR-003: Redpanda como bus de eventos (alternativa ligera a Kafka)

## Contexto

La arquitectura event-driven (ADR-001) requiere un bus de eventos duradero, con semántica de entrega clara y operación ligera, sin JVM ni ZooKeeper.

## Opciones Consideradas

1. NATS JetStream.
   - Pros: muy ligero; operación mínima; modelo simple.
   - Contras: otro modelo mental; sin ecosistema Kafka.
   - Coste: bajo.
2. Redpanda.
   - Pros: API compatible con Kafka; C++; sin JVM ni ZooKeeper; entregas idempotentes y transacciones; alto throughput.
   - Contras: requisitos de recursos mayores que NATS para un MVP pequeño.
   - Coste: medio.
3. Redis Streams.
   - Pros: mínima operación si ya existe Redis/Dragonfly.
   - Contras: semántica de entrega y retención limitadas para eventos duraderos y replay.
   - Coste: bajo.

## Decisión

Redpanda, por preferencia del propietario. Temas iniciales: mailbox.envelope.stored, mailbox.envelope.delivered, auth.session.revoked, realtime.signal. Consumidores idempotentes; clave de partición por usuario (hash determinista).

## Consecuencias

- Para el MVP puede iniciarse con un broker único; la recomendación en cloud gestionado es Redpanda serverless o instancia dedicada.
- El ecosistema Kafka queda reutilizable (tooling, mirror maker).

## Condición de revisión

Si tras tres meses de métricas el volumen real no justifica los recursos, migrar a NATS JetStream mediante un Adapter de publisher (Regla 4: aislamiento de dependencias externas).
