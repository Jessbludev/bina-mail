/**
 * Derivacion de la KEK desde la frase de contrasena (ADR-004).
 * Argon2id con t=3, m=64 MiB, p=4; determinista para la misma sal.
 */
import { argon2id } from "@noble/hashes/argon2";
import { ARGON2ID_PARAMS, KEY_SIZE } from "./params";

export function deriveKEK(passphrase: string, salt: Uint8Array): Uint8Array {
  if (salt.length < 16) {
    throw new Error("argon2id: la sal debe tener al menos 16 bytes");
  }
  return argon2id(passphrase, salt, { ...ARGON2ID_PARAMS, dkLen: KEY_SIZE });
}
