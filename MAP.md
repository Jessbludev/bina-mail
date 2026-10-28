# MAP — Mapa de modulos de Bina Mail

Regla: si un modulo no esta aqui, no existe operativamente. Marque lo pendiente con "?" y registrelo como trabajo nuevo. No lo invente ni lo asuma.

| Modulo | Responsabilidad unica | Dueno | SLO | Invariante | Dependencias |
|---|---|---|---|---|---|
| apps/web | SPA: UI, estado Pinia, orquestacion de features | web | TTI p95 < 2 s; LCP p75 < 2,5 s | Ninguna clave privada sale del dispositivo | gateway |
| apps/web/workers/crypto | KDF, cifrado de sobres, firmas | web | Cifrado de mensaje p95 < 100 ms | Cero acceso a DOM o red fuera del contrato de mensajes | apps/web |
| apps/web/ui | Sistema de diseño: tokens tipados y componentes base accesibles | web | Contraste AA 4,5:1 verificado en suite | Sin clases sueltas fuera de tokens.ts | apps/web |
| apps/web/views | Vistas de acceso, desbloqueo, bandeja, redacción y lectura | web | TTI p95 < 2 s | Ningún plaintext sale del Worker | apps/web/ui, workers/crypto, stores |
| apps/web/realtime | Cliente WS de señales con reconexión y refresco de bandeja | web | Señal→refresco p95 < 1 s | El token nunca viaja en URL ni query | gateway, stores |
| apps/web/search | Índice de búsqueda local en memoria sobre sobres descifrados | web | Búsqueda p95 < 50 ms con 500 sobres | El índice vive solo en RAM, nunca persiste en claro | workers/crypto |
| services/gateway | Borde público HTTP/WS (BFF), validación de input, cabeceras de seguridad, CORS por origen único | backend | p99 < 200 ms; disponibilidad 99,9 % | Ningún contenido en claro cruza este borde | auth, directory, mailbox, realtime |
| services/auth | Identidad, sesiones, MFA, rotación de tokens | backend | Disponibilidad 99,9 % | El servidor jamás almacena material clave recuperable | Supabase, Dragonfly |
| services/directory | Registro de claves públicas, fingerprints, verificación | backend | p99 < 100 ms | Almacena solo claves públicas y metadatos mínimos | Postgres, Dragonfly |
| services/mailbox | Persistencia y consulta de sobres cifrados | backend | Escritura p99 < 150 ms | El servidor nunca ve plaintext de mensajes | Postgres, Redpanda |
| services/realtime | Fan-out WS de señales de eventos | backend | Señal p95 < 500 ms | Notificaciones sin contenido ni metadatos de mensajes | Redpanda, Dragonfly |
| packages/contracts | Contratos protobuf, DTOs, tipos compartidos | plataforma | — | Es la única frontera de contrato entre servicios | — |

Formato de línea de log (LOG/YYYY-MM.md):

- [YYYY-MM-DD] <sistema> | <qué> | <por qué> | <medición antes→después> | <resultado>
