# Especificación criptográfica

Este documento es la fuente de verdad de diseño; la implementación vive en el código con vectores de test obligatorios.

## Derivación de cuenta

- Frase de contraseña → Argon2id (memoria 64 MiB, iteraciones 3, paralelismo 4, sal aleatoria de 16 bytes por cuenta, versión de parámetros registrada) → KEK de 32 bytes.
- La KEK cifra la clave privada de identidad X25519 en un sobre propio (self-wrapped) que permite la migración de dispositivo.

## Sobre de mensaje (formato versionado, binario compacto)

- version || nonce XChaCha (24 bytes) || ciphertext || tag.
- Lista de envolturas por destinatario: identificador de clave pública, clave pública efímera, clave efímera envuelta (X25519 + HKDF-SHA256) y fingerprint BLAKE3.
- Firma opcional Ed25519 del autor para repudio acotado.

## Determinismo

- Parámetros fijos por versión; nada dependiente del reloj en el formato.
- Seeds fijas en pruebas; reloj y aleatoriedad inyectados en la lógica de negocio.
- Mismos insumos y mismas claves producen el mismo fingerprint BLAKE3 (verificable).

## Rotación de claves

- Nueva clave de identidad publicada en directory con firma Ed25519 de la clave anterior.
- Histórico de fingerprints verificable.

## Vectores de test obligatorios

- Argon2id: vectores de RFC 9106.
- BLAKE3: vectores oficiales del proyecto.
- XChaCha20-Poly1305: vectores de libsodium.
- Extremo a extremo: vectores del formato de sobre por cada versión.
