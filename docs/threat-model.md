# Modelo de amenazas

## Activos protegidos

1. Contenido de mensajes (valor máximo).
2. Metadatos sociales (quién se comunica con quién y cuándo).
3. Claves privadas de los usuarios.
4. Disponibilidad del servicio.

## Adversarios y mitigaciones

- Servidor o infraestructura comprometida: el zero-knowledge limita el daño a sobres cifrados y metadatos mínimos; RLS y menor privilegio.
- Atacante de red: TLS externo, mTLS interno y E2EE por encima del transporte.
- Cliente comprometido: riesgo residual total, no mitigable en el servidor; mitigación con aislamiento del Web Worker, CSP estricta y auditoría de dependencias.
- Ataque a la cadena de suministro: lockfiles, versiones fijadas, verificación de módulos, gitleaks y escaneo de dependencias en CI.

## STRIDE (resumen)

| Amenaza | Mitigación principal |
|---|---|
| Spoofing | Autenticación mutua; firmas Ed25519 en rotación de claves |
| Tampering | AEAD (XChaCha20-Poly1305) + BLAKE3 + mTLS |
| Repudiation | Firma opcional de autoría por mensaje |
| Information disclosure | E2EE, zero-log, RLS, señales sin contenido |
| Denegación de servicio | Rate limiting con Dragonfly, circuit breakers, límites de tamaño |
| Elevación de privilegios | Menor privilegio por servicio; validación de input en el borde |

## Riesgo residual aceptado

El servidor puede inferir grafo social y patrones temporales (mitigación parcial: señales mínimas y agregación temporal). Documentado como residual y revisado en Fase 5.
