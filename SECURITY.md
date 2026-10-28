# Politica de seguridad

## Versiones con soporte

| Version | Soporte |
| ------- | ------- |
| Rama main | Si |
| Releases etiquetados | Aun no existen |

El proyecto esta en MVP: solo la rama main recibe correcciones de seguridad.

## Alcance

Dentro del alcance:

- Codigo del repositorio (apps/web, services, packages).
- Flujos de CI y acciones de GitHub.
- Configuracion de seguridad del gateway y de los servicios internos.

Fuera del alcance:

- Avisos de dependencias sin exploit contra este repositorio: reportelos a los proyectos originales. Dependabot (npm, go, actions), govulncheck y CodeQL ya los rastrean aqui.
- Errores funcionales, rendimiento o preguntas de uso: usen issues normales.

## Como reportar una vulnerabilidad

1. Usen "Report a vulnerability" en la pestana Security del repositorio (informe privado de GitHub).
2. Incluyan descripcion, pasos de reproduccion, impacto estimado y commit afectado.
3. No abran issues publicos con detalles de seguridad.

## Que esperar

- Acuse de recibo en 5 dias habiles.
- Decision (aceptada o rechazada) en 14 dias habiles, con razon documentada.
- Si se acepta: correccion en main y aviso en el reporte; credito al reportante si lo desea.
- Divulgacion coordinada: no se publican detalles hasta que exista correccion.

## Compromiso del proyecto

- Zero-log: el backend no registra contenido de mensajes, remitentes ni destinatarios. Cualquier hallazgo que rompa esa politica se trata como vulnerabilidad.
- Las claves privadas de los usuarios nunca salen de su dispositivo; el servidor solo almacena ciphertext.

## Summary (English)

Only the main branch is supported; there are no tagged releases yet. Report vulnerabilities through GitHub Private Vulnerability Reporting (Security tab, "Report a vulnerability"), never in public issues. Expect acknowledgement within 5 business days and a decision within 14. Coordinated disclosure only. Out of scope: dependency advisories without an exploit against this repository (Dependabot, govulncheck and CodeQL already track them) and functional bugs.
