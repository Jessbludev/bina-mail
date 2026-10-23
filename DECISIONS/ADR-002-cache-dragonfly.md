# ADR-002: Dragonfly como capa de cache

## Contexto

Se necesita cache de sesiones, fingerprints de claves y rate limiting, con baja latencia, protocolo compatible con Redis y operación en cloud gestionado, bajo la política zero-log.

## Opciones Consideradas

1. Redis OSS.
   - Pros: estándar de facto; ecosistema máximo.
   - Contras: historial de cambios de licencia; gobernanza centrada en el proveedor; modelo single-thread dominante.
   - Coste: bajo.
2. Valkey.
   - Pros: fork estable respaldado por Linux Foundation; licencia BSD; drop-in de Redis.
   - Contras: modelo de ejecución heredado de Redis (menor uso de núcleos).
   - Coste: bajo.
3. Dragonfly.
   - Pros: motor multi-thread (aprovecha núcleos); menor huella de memoria; API compatible con Redis; datasheets con throughput superior en hardware multicore.
   - Contras: proyecto más joven; ecosistema menor.
   - Coste: medio.

## Decisión

Dragonfly, por preferencia del propietario. Despliegue como contenedor en el host de la plataforma o como servicio gestionado. Usos: sesiones con TTL corto, fingerprints de claves públicas y contadores de rate limiting. Nunca contenido de mensajes ni material clave.

## Consecuencias

- La compatibilidad con el protocolo Redis habilita el cliente go-redis sin adaptadores.
- Se debe seguir el ciclo de releases de Dragonfly (riesgo de vértigo de versión).

## Condición de revisión

Inestabilidad en producción o coste superior al de Valkey equivalente.
