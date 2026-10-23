/**
 * Contrato tipado de mensajes entre la app y el Worker criptográfico.
 * El worker es la única frontera para material clave (criterio zero-knowledge).
 */
export type CryptoRequestId = string;

export type CryptoRequest =
  | { id: CryptoRequestId; type: "derive-kek"; passphrase: string }
  | { id: CryptoRequestId; type: "seal"; plaintext: string; recipientFingerprints: string[] }
  | { id: CryptoRequestId; type: "open"; envelopeId: string };

export type CryptoResponse =
  | { id: CryptoRequestId; ok: true; payload: unknown }
  | { id: CryptoRequestId; ok: false; error: string };
