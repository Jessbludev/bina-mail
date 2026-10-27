/**
 * Contrato tipado de mensajes entre la app y el Worker criptografico.
 * El Worker es la unica frontera para material clave (criterio zero-knowledge).
 */
import type { EnvelopeJson } from "./envelope";

export type CryptoRequestId = string;

export type CryptoRequest =
  | { id: CryptoRequestId; type: "derive-kek"; passphrase: string; saltB64: string }
  | { id: CryptoRequestId; type: "generate-identity" }
  | {
      id: CryptoRequestId;
      type: "seal";
      plaintext: string;
      recipientPublicKeyB64: string;
      recipientFingerprint: string;
      senderFingerprint?: string;
    }
  | { id: CryptoRequestId; type: "open"; envelope: EnvelopeJson; privateKeyB64: string };

export type CryptoResponse =
  | { id: CryptoRequestId; ok: true; payload: unknown }
  | { id: CryptoRequestId; ok: false; error: string };
