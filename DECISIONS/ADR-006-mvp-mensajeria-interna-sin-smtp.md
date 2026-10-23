# ADR-006: Alcance del MVP — mensajería interna E2EE sin SMTP

## Contexto

Definir la superficie del MVP determina la existencia o no de un MTA y la interoperabilidad con correo externo.

## Opciones Consideradas

1. SMTP completo desde el día uno (Postfix o Haraka, DKIM/SPF/DMARC).
   - Pros: interoperabilidad inmediata.
   - Contras: superficie grande; E2EE no garantizable con destinatarios externos.
   - Coste: alto.
2. Mensajería interna E2EE entre usuarios de la plataforma.
   - Pros: superficie mínima; zero-log sostenible; E2EE extremo a extremo real.
   - Contras: sin interoperabilidad externa.
   - Coste: medio.
3. Interno primero, pasarela SMTP en fase 2.
   - Pros: valor rápido con camino de evolución.
   - Contras: la fase 2 obliga a un threat model propio del SMTP externo.
   - Coste: medio.

## Decisión

Opción 2 para el MVP, por decisión del propietario. No hay MTA; la entrega es entre usuarios internos. La pasarela externa queda como fase futura con ADR propio.

## Consecuencias

- No existe interoperabilidad de correo externo en el MVP.
- WebRTC queda fuera de alcance; realtime por WebSocket.
- El modelo de amenazas se simplifica (docs/threat-model.md).

## Condición de revisión

Un requisito de interoperabilidad externa abre un nuevo ADR y una revisión del threat model.
