# ADR-005: Política zero-log y observabilidad

## Contexto

El criterio de proyecto exige zero-log. La observabilidad debe existir sin registrar contenido ni metadatos de mensajes.

## Opciones Consideradas

1. Logs verbosos con PII filtrada.
   - Pros: debugging fácil.
   - Contras: riesgo alto de fuga de metadatos; incompatible con el criterio.
   - Coste: bajo a corto plazo, alto en incidentes.
2. Zero-log estricto: métricas agregadas, errores sanitizados y correlación por identificadores opacos rotables.
   - Pros: compatible con el criterio; riesgo mínimo.
   - Contras: debugging requiere reproducción local con trazas muestreadas sanitizadas.
   - Coste: medio.
3. Zero-log en servidor más logs locales opt-in exportables por el usuario.
   - Pros: el usuario conserva evidencia si la desea.
   - Contras: requiere superficie de exportación en el cliente.
   - Coste: bajo.

## Decisión

Combinación de las opciones 2 y 3. Reglas:

- Ningún campo de log contiene contenido, asunto, remitente, destinatario ni IP en claro.
- Trazas y métricas RED/USE solo con nombres de operación, códigos y duraciones.
- Correlación por trace-ID opaco con TTL corto; sin perfiles de usuario.
- La CI audita patrones de log prohibidos en cada cambio (job security).

## Consecuencias

- Cualquier hallazgo de auditoría genera deuda inmediata en DEBT.md con dueño y fecha.

## Condición de revisión

Auditoría de logs en cada release mayor.
