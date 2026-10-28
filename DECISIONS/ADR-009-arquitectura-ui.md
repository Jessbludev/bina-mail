# ADR-009: Arquitectura de UI de la Fase 4 (vistas, router y sistema de diseno)

- Estado: Aceptada
- Fecha: 2026-10-28

## Contexto

La Fase 4 construye la interfaz completa sobre el dominio ya verificado: acceso y desbloqueo, bandeja con senales realtime, redaccion con cifrado en el Worker, lectura con descifrado y busqueda local sobre indices cifrados del cliente. Hoy apps/web tiene la cripto completa en el Worker, el cliente API tipado y un store de sesion minimalista; no existen vistas, router ni sistema de diseno real. El cambio afecta a mas de un modulo (Regla 0: ADR previo obligatorio).

## Opciones

1. vue-router oficial con vistas en rutas lazy, contenedores Pinia por feature y componentes presentacionales con clases tipadas.
   - Pros: la URL es la unica fuente de verdad de la vista (un dueno por estado), boton atras y deep-links operativos, foco gestionable por ruta, code splitting por ruta (presupuesto de bundle, Regla W4), familia oficial Vue auditada por la CI.
   - Contras: dependencia nueva en runtime (vue-router ^4) y de desarrollo (@vue/test-utils, jsdom para pruebas de montaje).
   - Coste: bajo; la CI sincroniza pnpm-lock en cada push.
2. Navegacion manual por store Pinia, sin router.
   - Pros: cero dependencias nuevas.
   - Contras: la URL no refleja la vista, boton atras roto (fallo de usabilidad y accesibilidad), foco y deep-links gestionados a mano (mas codigo que el router).
   - Coste: medio en mantenimiento, alto en UX.
3. Vistas condicionales en un unico App.vue.
   - Pros: minimo codigo inicial.
   - Contras: no escala a cinco vistas con estados completos; viola la frontera del mapa.
   - Coste: bajo hoy, alto manana.

## Decision

Opcion 1. Detalles fijados:

- Arquitectura container/presentational: las vistas orquestan stores y Worker; los componentes base (Button, Field, Panel, Skeleton, StatusBar) son presentacionales y solo reciben clases tipadas de styles/tokens.ts. Sin clases sueltas en plantillas ni estilos ad-hoc.
- Estados completos por vista (loading con skeleton, empty, error, offline y success): una vista sin estados definidos no se considera terminada.
- Rutas: /acceso, /desbloqueo, /bandeja, /bandeja/:id, /redaccion. Guard por estado de sesion y de desbloqueo; ninguna ruta lleva token ni datos en la URL.
- Busqueda local: indice en memoria (Map por sobre) reconstruido en cada sesion a partir de los sobres descifrados; nunca persiste en claro (criterio zero-knowledge). El termino buscado no sale del dispositivo.
- Realtime: cliente WS en shared/realtime.ts con reconexion exponencial y refresco de bandeja ante senal; el token solo viaja en el primer mensaje del canal.
- Accesibilidad (Regla W1): WCAG 2.1 AA como suelo (contraste 4,5:1, teclado completo, foco visible, label asociado, prefers-reduced-motion, escalado 200%); skip link y landmarks semanticos en el shell.

## Consecuencias

- pnpm-lock se sincroniza automaticamente: el job lockfiles corre en cada push a main y regenera lockfiles desincronizados (pnpm install --lockfile-only y go mod tidy), cerrando la ventana manual.
- La prueba de flujo completo entre dos usuarios se ejecuta en vitest con el Worker real (Node); el E2E con navegador real (Playwright) queda como DEBT-005 para la Fase 5 con el despliegue.
- Revision en Fase 5 junto al hardening: mediciones Lighthouse y CWV sobre despliegue real, no en desarrollo.
