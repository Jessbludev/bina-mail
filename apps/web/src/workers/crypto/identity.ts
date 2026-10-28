/**
 * Identidad sellada del usuario (ADR-004/ADR-009): par X25519 de cifrado y
 * par Ed25519 de firma (el directorio exige ambas claves publicas). El
 * bundle privado se cifra con la KEK derivada de la frase de contrasena;
 * solo el bloque sellado sale del Worker hacia localStorage del
 * dispositivo. La frase nunca se guarda ni sale del cliente.
 */
import { ed25519, x25519 } from "@noble/curves/ed25519";
import { randomBytes } from "@noble/ciphers/utils";
import { deriveKEK } from "./argon2id";
import { fromB64, toB64 } from "./b64";
import { fingerprint } from "./blake3";
import { openBox, sealBox } from "./aead";
import { generateIdentity } from "./x25519";
import { KEY_SIZE, XCHACHA_NONCE_SIZE } from "./params";

export interface SealedIdentity {
  readonly saltB64: string;
  readonly identitySealedB64: string;
}

/** Datos publicos del vault local del dispositivo. */
export interface IdentityRecord {
  readonly saltB64: string;
  readonly identitySealedB64: string;
  readonly x25519PublicKeyB64: string;
  readonly ed25519PublicKeyB64: string;
  readonly fingerprint: string;
}

/** Claves privadas solo para la RAM de la sesion; nunca localStorage ni red. */
export interface UnlockedIdentity {
  readonly encryptionPrivateKeyB64: string;
  readonly signingPrivateKeyB64: string;
  readonly x25519PublicKeyB64: string;
  readonly ed25519PublicKeyB64: string;
  readonly fingerprint: string;
}

export type IdentityBundle = IdentityRecord & UnlockedIdentity;

const SALT_SIZE = 32;
const MIN_PASSPHRASE = 8;

export function createIdentity(passphrase: string): IdentityBundle {
  if (passphrase.length < MIN_PASSPHRASE) {
    throw new Error("frase_corta");
  }
  const salt = randomBytes(SALT_SIZE);
  const kek = deriveKEK(passphrase, salt);
  const encryption = generateIdentity();
  const signingPrivate = ed25519.utils.randomPrivateKey();
  const signingPublic = ed25519.getPublicKey(signingPrivate);
  const bundle = concat(encryption.privateKey, signingPrivate);
  const sealed = sealBox(kek, bundle);
  return {
    saltB64: toB64(salt),
    identitySealedB64: toB64(concat(sealed.nonce, sealed.ciphertext)),
    x25519PublicKeyB64: toB64(encryption.publicKey),
    ed25519PublicKeyB64: toB64(signingPublic),
    fingerprint: fingerprint(encryption.publicKey),
    encryptionPrivateKeyB64: toB64(encryption.privateKey),
    signingPrivateKeyB64: toB64(signingPrivate),
  };
}

export function unlockIdentity(passphrase: string, sealed: SealedIdentity): UnlockedIdentity {
  const salt = fromB64(sealed.saltB64);
  const kek = deriveKEK(passphrase, salt);
  const raw = fromB64(sealed.identitySealedB64);
  if (raw.length <= XCHACHA_NONCE_SIZE) {
    throw new Error("paquete_invalido");
  }
  let bundle: Uint8Array;
  try {
    bundle = openBox(kek, {
      nonce: raw.subarray(0, XCHACHA_NONCE_SIZE),
      ciphertext: raw.subarray(XCHACHA_NONCE_SIZE),
    });
  } catch {
    // La autenticacion AEAD falla con cualquier frase incorrecta.
    throw new Error("frase_invalida");
  }
  if (bundle.length !== KEY_SIZE * 2) {
    throw new Error("paquete_invalido");
  }
  const encryptionPrivate = bundle.subarray(0, KEY_SIZE);
  const signingPrivate = bundle.subarray(KEY_SIZE);
  const encryptionPublic = x25519.getPublicKey(encryptionPrivate);
  const signingPublic = ed25519.getPublicKey(signingPrivate);
  return {
    encryptionPrivateKeyB64: toB64(encryptionPrivate),
    signingPrivateKeyB64: toB64(signingPrivate),
    x25519PublicKeyB64: toB64(encryptionPublic),
    ed25519PublicKeyB64: toB64(signingPublic),
    fingerprint: fingerprint(encryptionPublic),
  };
}

function concat(a: Uint8Array, b: Uint8Array): Uint8Array {
  const out = new Uint8Array(a.length + b.length);
  out.set(a);
  out.set(b, a.length);
  return out;
}
