# Snapshot v3 — Integracion realtime (Fase 3)

## Fecha

2026-10-28

## Cambios desde el ultimo Snapshot

- Redpanda con franz-go v1.22.1 y WebSocket con gorilla/websocket v1.5.3: primeras dependencias externas, auditadas y confinadas a adapters (ADR-008); Go del modulo sube a 1.26 y go.sum lo genera y confirma la CI.
- Esquema de eventos versionado (mailbox.envelope.stored v1) con decodificacion defensiva y tema bina-mail.mailbox.envelope-stored.v1; prueba de extremo a extremo produce y consume contra un broker kfake en proceso.
- realtime: autenticacion de dos pasos dentro del canal (sesion opaca contra auth, fingerprint contra directory), senales sin contenido (ADR-005), cierre 4401 ante fallo, ping cada 30 segundos y read deadline de 90 segundos.
- gateway: proxy WS hacia realtime, logout con revocacion, CORS por origen unico, cabeceras de seguridad, recuperacion de panicos sin fuga de detalles y limite por ventana fija con fail-open documentado.
- Cache Dragonfly con cliente RESP2 en stdlib: sesiones con TTL de 30 segundos acotado por su expiracion e invalidado al revocar; fingerprints e indice por usuario con TTL de 60 segundos (DEBT-004 para la entrada por fingerprint anterior).
- Se elimina el diseno intermedio de entrega HTTP interna de eventos; el bus es la unica via de senales.
- CI: job go gated a la existencia de go.sum y job lockfiles ampliado para generar pnpm-lock.yaml y go.sum con confirmacion automatica (DEBT-003).

## Estado Actual

- Dominio E2EE e integracion realtime completos con suites de prueba; la CI verificara el gate Go en cuanto confirme go.sum en el primer flujo.
- La interfaz de usuario aun no consume el dominio ni las senales: llega en Fase 4.

## Metricas Clave

- Go: cuatro servicios con montaje final (Dragonfly, Redpanda, limite, proxy WS) y suites httptest por servicio; eliminadas ocho unidades de codigo del diseno previo; cobertura medida por la CI cuando el gate Go se active.

## Proximos Pasos

- Fase 4: interfaz de usuario (Vue/Pinia) sobre el dominio y las senales realtime.
