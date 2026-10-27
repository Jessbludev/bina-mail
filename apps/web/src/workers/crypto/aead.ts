/**
 * Cifrado autenticado XChaCha20-Poly1305 (ADR-004).
 * El ciphertext incluye la etiqueta AEAD de 16 bytes; la apertura
 * falla si el ciphertext o el nonce fueron manipulados.
 */
import { xchacha20poly1305 } from "@noble/ciphers/chacha";
import { randomBytes } from "@noble/ciphers/utils";
import { KEY_SIZE, XCHACHA_NONCE_SIZE } from "./params";

export interface SealedBox {
  readonly nonce: Uint8Array;
  readonly ciphertext: Uint8Array;
}

export function sealBox(key: Uint8Array, plaintext: Uint8Array): SealedBox {
  requireKey(key);
  const nonce = randomBytes(XCHACHA_NONCE_SIZE);
  const ciphertext = xchacha20poly1305(key, nonce).encrypt(plaintext);
  return { nonce, ciphertext };
}

export function openBox(key: Uint8Array, box: SealedBox): Uint8Array {
  requireKey(key);
  if (box.nonce.length !== XCHACHA_NONCE_SIZE) {
    throw new Error("aead: nonce invalido");
  }
  return xchacha20poly1305(key, box.nonce).decrypt(box.ciphertext);
}

function requireKey(key: Uint8Array): void {
  if (key.length !== KEY_SIZE) {
    throw new Error("aead: tamano de clave invalido");
  }
}
