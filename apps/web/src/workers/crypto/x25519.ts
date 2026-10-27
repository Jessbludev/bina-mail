/**
 * Identidades X25519 del usuario y secreto compartido (ADR-004).
 * La clave privada nunca sale del cliente ni del Worker.
 */
import { x25519 } from "@noble/curves/ed25519";
import { KEY_SIZE } from "./params";

export interface IdentityKeys {
  readonly privateKey: Uint8Array;
  readonly publicKey: Uint8Array;
}

export function generateIdentity(): IdentityKeys {
  const privateKey = x25519.utils.randomPrivateKey();
  return { privateKey, publicKey: x25519.getPublicKey(privateKey) };
}

export function sharedSecret(privateKey: Uint8Array, publicKey: Uint8Array): Uint8Array {
  if (privateKey.length !== KEY_SIZE || publicKey.length !== KEY_SIZE) {
    throw new Error("x25519: tamano de clave invalido");
  }
  return x25519.getSharedKey(privateKey, publicKey);
}
