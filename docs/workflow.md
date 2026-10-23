# Flujo de trabajo y calidad

## Ciclo por tarea

1. Tomar la tarea del backlog (issue con plantilla).
2. Análisis de impacto sobre MAP.md; si afecta a más de un módulo o a un componente crítico, redactar ADR antes de codificar.
3. Test-first: pruebas que fallen antes de implementar.
4. Implementación contra contratos de packages/contracts.
5. Auto-revisión: OWASP, WCAG 2.1 AA, bundle inicial < 1 MB, sin secretos, cumplimiento zero-log (ADR-005).
6. Actualizar el SNAPSHOT al cerrar el hito.
7. Pull request con plantilla: riesgos, pruebas realizadas, impacto en UX.
8. CI estricta: todo verde es obligatorio; nada se fusiona con gates rojos.

## Ramas y commits

- Trunk-based: rama main protegida; ramas cortas feat/, fix/, chore/.
- Commits convencionales: feat, fix, refactor, docs, test, chore.

## Pirámide de pruebas

- 70 % unitarias (vitest en web, go test en backend; vectores criptográficos incluidos).
- 20 % integración (gRPC con dockertest; RLS de Postgres).
- 10 % E2E (Playwright sobre entorno de preview).
- Cobertura mínima: 80 % en dominio; 100 % en módulos criptográficos.

## Gates de CI (bloqueantes)

- Tipado estricto (vue-tsc, go vet).
- Lint (eslint, golangci-lint con gosec).
- Pruebas con detección de carrera de datos (go test -race) y cobertura.
- Análisis de secretos (gitleaks).
- Auditoría de patrones de log prohibidos (zero-log).
- actionlint y markdownlint.
- Escaneo de vulnerabilidades de dependencias (govulncheck).

## Determinismo

- Reloj y aleatoriedad inyectados; seeds fijas en pruebas.
- Sin lógica dependiente de zona horaria no declarada.

## Definition of Done

Pruebas verdes, CI verde, PR revisada, documentación y MAP.md actualizados, línea de log escrita, deuda registrada en DEBT.md si aplica.

## Registros

- Línea de log (LOG/YYYY-MM.md): fecha, sistema, qué, por qué, medición antes→después, resultado.
- Snapshot por hito: estado, cambios, métricas, próximos pasos.
- Deuda: registro explícito con dueño y fecha de revisión.
