/**
 * Sobres E2EE de Bina Mail (ADR-004). Los nombres de campo replican
 * exactamente el contrato proto canonico (packages/contracts, ADR-007).
 */
import { hkdf } from "@noble/hashes/hkdf";
import { sha256 } from "@noble/hashes/sha256";
import { randomBytes } from "@noble/ciphers/utils";
import { openBox, sealBox } from "./aead";
import { toB64, fromB64 } from "./b64";
import { fingerprint } from "./blake3";
import { generateIdentity, sharedSecret } from "./x25519";
import { KEY_SIZE, PARAMS_VERSION, XCHACHA_NONCE_SIZE } from "./params";

export interface EnvelopeWrap {
  readonly recipient_key_id: string;
  readonly ephemeral_public_key: string;
  readonly wrapped_key: string;
}

export interface EnvelopeJson {
  readonly version: number;
  readonly sender_fingerprint?: string;
  readonly recipient_fingerprint: string;
  readonly nonce: string;
  readonly ciphertext: string;
  readonly wraps: readonly EnvelopeWrap[];
  readonly blake3_fingerprint: string;
  readonly created_at_unix: number;
}

export interface Recipient {
  readonly keyId: string;
  readonly publicKey: Uint8Array;
}

export interface BuildEnvelopeOptions {
  readonly senderFingerprint?: string;
  readonly now?: () => number;
}

const WRAP_INFO = `bina-mail/envelope-wrap/v${PARAMS_VERSION}`;

export function buildEnvelope(
  plaintext: string,
  recipient: Recipient,
  options?: BuildEnvelopeOptions,
): EnvelopeJson {
  const messageKey = randomBytes(KEY_SIZE);
  const box = sealBox(messageKey, new TextEncoder().encode(plaintext));
  const ephemeral = generateIdentity();
  const secret = sharedSecret(ephemeral.privateKey, recipient.publicKey);
  const wrapKey = hkdf(sha256, secret, ephemeral.publicKey, WRAP_INFO, KEY_SIZE);
  const wrapped = sealBox(wrapKey, messageKey);
  const nowFn = options?.now ?? Date.now;
  return {
    version: PARAMS_VERSION,
    sender_fingerprint: options?.senderFingerprint,
    recipient_fingerprint: recipient.keyId,
    nonce: toB64(box.nonce),
    ciphertext: toB64(box.ciphertext),
    wraps: [
      {
        recipient_key_id: recipient.keyId,
        ephemeral_public_key: toB64(ephemeral.publicKey),
        wrapped_key: toB64(concat(wrapped.nonce, wrapped.ciphertext)),
      },
    ],
    blake3_fingerprint: fingerprint(concat(box.nonce, box.ciphertext)),
    created_at_unix: Math.floor(nowFn() / 1000),
  };
}

export function openEnvelope(
  envelope: EnvelopeJson,
  privateKey: Uint8Array,
): Uint8Array {
  const nonce = fromB64(envelope.nonce);
  const ciphertext = fromB64(envelope.ciphertext);
  const expected = fingerprint(concat(nonce, ciphertext));
  if (expected !== envelope.blake3_fingerprint) {
    throw new Error("envelope: fingerprint no coincide");
  }
  for (const wrap of envelope.wraps) {
    const ephemeralPublic = fromB64(wrap.ephemeral_public_key);
    try {
      const secret = sharedSecret(privateKey, ephemeralPublic);
      const wrapKey = hkdf(sha256, secret, ephemeralPublic, WRAP_INFO, KEY_SIZE);
      const raw = fromB64(wrap.wrapped_key);
      const messageKey = openBox(wrapKey, {
        nonce: raw.subarray(0, XCHACHA_NONCE_SIZE),
        ciphertext: raw.subarray(XCHACHA_NONCE_SIZE),
      });
      return openBox(messageKey, { nonce, ciphertext });
    } catch {
      // Envoltura que no corresponde a esta clave: se intenta la siguiente.
    }
  }
  throw new Error("envelope: ninguna envoltura pudo abrirse");
}

function concat(a: Uint8Array, b: Uint8Array): Uint8Array {
  const out = new Uint8Array(a.length + b.length);
  out.set(a);
  out.set(b, a.length);
  return out;
}
