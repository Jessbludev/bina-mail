# Contribuir a Bina Mail

Gracias por mejorar una plataforma que promete privacidad real: cada cambio debe respetar los criterios del repositorio (zero-knowledge, zero-log, determinismo y CI estricta).

## Requisitos

- Go 1.26
- Node.js 22 y pnpm 9
- buf (contratos protobuf), opcional si no toca packages/contracts

## Puesta a punto

Desde la raiz del repositorio:

```bash
cd apps/web && pnpm install
cd ../services && go build ./...
```

## Gates locales (los mismos que ejecuta la CI)

Web, en apps/web:

```bash
pnpm typecheck
pnpm lint
pnpm test
```

Go, en services:

```bash
golangci-lint run
go test ./... -race
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

Contratos, en packages/contracts:

```bash
buf lint
```

El lockfile de pnpm y el go.sum se sincronizan automaticamente en cada push a main; no hace falta generarlos a mano en un PR.

## Reglas del repositorio

1. Lea MAP.md antes de navegar el codigo; el mapa es la unica vista operacional de los modulos.
2. Todo cambio que afecte a mas de un modulo requiere un ADR previo en DECISIONS/, siguiendo el formato de los ADR existentes.
3. Test primero: ningun cambio de dominio sin pruebas que lo cubran.
4. Zero-log: prohibido registrar contenido de mensajes, remitentes, destinatarios o frases de identidad.
5. Criptografia: los paquetes @noble permanecen fijados a version exacta; no los actualice sin un ADR que lo justifique.
6. Interfaz: clases solo desde styles/tokens.ts y estados completos por vista (cargando, vacio, error, exito).
7. Commits con estilo conventional: feat, fix, docs, chore, test, refactor.
8. La CI es bloqueante: gitleaks, actionlint, markdownlint, buf lint, golangci-lint, pruebas con -race, typecheck y tests deben quedar en verde.
9. Cada hito cierra con un SNAPSHOT y una linea en LOG/; toda deuda se registra en DEBT.md antes de fusionar.

## Pull requests

- Usen la plantilla de PR del repositorio: riesgos, pruebas realizadas e impacto en UX.
- Expliquen el por que antes del como; enlacen el ADR cuando aplique.
- La revision de un maintainer es obligatoria antes de fusionar.

## Seguridad

No abran issues publicos con vulnerabilidades; sigan SECURITY.md. Las incidencias de conducta se gestionan segun CODE_OF_CONDUCT.md.

## Summary (English)

Set up with Go 1.26, Node 22 and pnpm 9. Run the same gates as CI locally: web typecheck/lint/tests in apps/web, golangci-lint and go test -race in services, buf lint in packages/contracts. Read MAP.md first; any cross-module change needs an ADR in DECISIONS/ before coding. Tests come first, logging of message content or identities is forbidden, @noble packages stay pinned, and UI classes come only from styles/tokens.ts. Lockfiles are synced automatically on push to main. PRs use the repository template and require maintainer review. Report security issues privately per SECURITY.md.
