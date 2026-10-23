# ADR-004: Criptografía E2EE y bibliotecas

## Contexto

Zero-knowledge: el servidor no puede leer mensajes. Se necesita KDF de frase de contraseña, intercambio de claves, cifrado autenticado e integridad, con comportamiento determinista y verificable.

## Opciones Consideradas

1. go-libsodium (binding CGO de libsodium).
   - Pros: primitivas auditadas; API simple.
   - Contras: CGO rompe la compilación cruzada determinista y complica la cadena de suministro.
   - Coste: medio.
2. Stdlib Go + golang.org/x/crypto + BLAKE3 en Go puro.
   - Pros: sin CGO; Argon2id en x/crypto; XChaCha20-Poly1305 en x/crypto/chacha20poly1305; binarias reproducibles.
   - Contras: exige fijar parámetros propios y vectores de test propios; disciplina de uso correcto.
   - Coste: medio.
3. Módulos FIPS 140-3.
   - Pros: cumplimiento normativo estatal.
   - Contras: primitivas orientadas a compliance; no alineadas con E2EE moderna; overhead innecesario.
   - Coste: alto.

## Decisión

Opción 2 en el servidor. En el cliente (Web Workers), bibliotecas auditadas sin dependencias: @noble/curves y @noble/hashes. Diseño:

- Frase de contraseña → Argon2id (memoria 64 MiB, iteraciones 3, paralelismo 4; parámetros versionados) → KEK de 32 bytes.
- Identidad: keypair X25519; firma Ed25519 para autoría y rotación.
- Sobre por mensaje: clave efímera simétrica; cifrado XChaCha20-Poly1305; la clave efímera se envuelve por destinatario con X25519 + HKDF-SHA256.
- BLAKE3: integridad, deduplicación y fingerprints de claves públicas.
- El servidor almacena solo sobres cifrados y claves públicas.

## Consecuencias

- Vectores de test obligatorios por primitiva y por versión de parámetros (determinismo verificable).
- La rotación de claves usa un sobre propio del usuario (self-wrapped) para migración de dispositivo.

## Condición de revisión

Revisión criptográfica externa antes de producción; reevaluar parámetros de Argon2id anualmente según guías OWASP.
