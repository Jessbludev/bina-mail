# apps/web — SPA Bina Mail

TypeScript puro, Vue 3 (Composition API), Pinia y Tailwind con clases seguras y tipadas definidas en TypeScript. Toda operación criptográfica se ejecuta en Web Workers.

## Estructura prevista

- src/features/{auth,mailbox,compose,settings} — cada feature con components, stores (Pinia) y types.
- src/shared — componentes compartidos y tokens de diseño.
- src/workers/crypto — aislamiento criptográfico (KDF, sobres, firmas).
- src/styles/tokens.ts — clases Tailwind tipadas (variantes tipadas); prohibidas las clases sueltas en plantillas.

## Reglas

- WCAG 2.1 AA.
- Bundle inicial < 1 MB; code splitting y lazy loading.
- Ninguna clave privada sale del worker ni del dispositivo.
- Estado global solo en Pinia, por feature.
