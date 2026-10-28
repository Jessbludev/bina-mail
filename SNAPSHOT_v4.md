# Snapshot v4 — Interfaz de usuario (Fase 4)

## Fecha

2026-10-28

## Cambios desde el ultimo Snapshot

- ADR-009 y sistema de diseno token-first: clases solo desde styles/tokens.ts (prohibido clases sueltas en plantillas), componentes base reutilizables (boton, campo, panel, skeleton, estado vacio), container/presentational y estados completos por vista.
- Identidad sellada en el dispositivo: KEK Argon2id derivado de la frase, bundle de pares X25519 (cifrado) y Ed25519 (firma, exigido por el directorio) sellado con XChaCha20-Poly1305; el vault persiste unicamente el bloque sellado y las claves publicas; las privadas viven en RAM durante la sesion.
- Sesion de tres fases (anonimo, sesion, desbloqueo) con guard de rutas por fase; ninguna URL transporta token; la sesion opaca se revoca al cerrar sesion.
- Buzon con descifrado bajo demanda en el Worker: cache de textos claros solo en RAM, lectura con verificacion BLAKE3 del sobre y estados completos (skeleton, vacio honesto, error con reintento).
- Redaccion: resolucion de clave publica por huella, sello del sobre en el Worker con huella del remitente y almacenamiento en el buzon; errores por codigo (404, 429) con mensajes honestos.
- Canal realtime en shared: sesion opaca como primer mensaje del WebSocket, reconexion exponencial acotada (500 ms a 15 s), abandono definitivo ante cierre 4401 y refresco de bandeja al recibir la senal sin contenido.
- Busqueda local en RAM sobre los sobres ya descifrados en la sesion: termino normalizado sin diacriticos, sin persistencia y sin participacion del servidor; SLO p95 < 50 ms con 500 sobres verificado en suite.

## Estado Actual

- La interfaz consume el dominio y las senales: cinco vistas (acceso, desbloqueo, bandeja, lectura, redaccion) con estados completos, dos stores probados (sesion y buzon) y modulos shared (api, realtime, busqueda) con suites vitest; la CI compila y ejecuta las suites en cada flujo.
- El E2E de navegador con dos usuarios reales queda como DEBT-005 para Fase 5.

## Metricas Clave

- Web: cinco vistas con estados completos y a11y (skip link, landmarks, roles alert/status, foco visible, contraste AA); suites vitest de sesion, buzon, realtime y busqueda; ningun dato claro fuera de RAM ni de localStorage sellado.

## Proximos Pasos

- Fase 5 (operacion): despliegue, E2E de navegador (DEBT-005), gRPC interno (DEBT-002) e invalidacion de cache por fingerprint (DEBT-004).
