# MAP — Mapa de módulos de Bina Mail

Regla: si un módulo no está aquí, no existe operativamente. Marque lo pendiente con "?" y regístrelo como trabajo nuevo. No lo invente ni lo asuma.

| Módulo | Responsabilidad única | Dueño | SLO | Invariante | Dependencias |
|---|---|---|---|---|---|
| apps/web | SPA: UI, estado Pinia, orquestación de features | web | TTI p95 < 2 s; LCP p75 < 2,5 s | Ninguna clave privada sale del dispositivo | gateway |
| apps/web/workers/crypto | KDF, cifrado de sobres, firmas | web | Cifrado de mensaje p95 < 100 ms | Cero acceso a DOM o red fuera del contrato de mensajes | apps/web |
| services/gateway | Borde público HTTP/WS (Chi), validación de input, BFF | backend | p99 < 200 ms; disponibilidad 99,9 % | Ningún contenido en claro cruza este borde | auth, directory, mailbox, realtime |
| services/auth | Identidad, sesiones, MFA, rotación de tokens | backend | Disponibilidad 99,9 % | El servidor jamás almacena material clave recuperable | Supabase, Dragonfly |
| services/directory | Registro de claves públicas, fingerprints, verificación | backend | p99 < 100 ms | Almacena solo claves públicas y metadatos mínimos | Postgres, Dragonfly |
| services/mailbox | Persistencia y consulta de sobres cifrados | backend | Escritura p99 < 150 ms | El servidor nunca ve plaintext de mensajes | Postgres, Redpanda |
| services/realtime | Fan-out WS de señales de eventos | backend | Señal p95 < 500 ms | Notificaciones sin contenido ni metadatos de mensajes | Redpanda, Dragonfly |
| packages/contracts | Contratos protobuf, DTOs, tipos compartidos | plataforma | — | Es la única frontera de contrato entre servicios | — |

Formato de línea de log (LOG/YYYY-MM.md):

- [YYYY-MM-DD] <sistema> | <qué> | <por qué> | <medición antes→después> | <resultado>
