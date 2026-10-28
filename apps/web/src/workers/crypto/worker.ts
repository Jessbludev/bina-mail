import { deriveKEK } from "./argon2id";
import { fingerprint } from "./blake3";
import { fromB64, toB64 } from "./b64";
import { buildEnvelope, openEnvelope } from "./envelope";
import { createIdentity, unlockIdentity } from "./identity";
import { generateIdentity } from "./x25519";
import type { CryptoRequest, CryptoResponse } from "./messages";

/**
 * Worker criptografico de Bina Mail (ADR-004).
 * Toda primitiva criptografica vive aqui; la app solo ve base64 y JSON.
 */
const ctx = self as unknown as {
  onmessage: ((event: MessageEvent<CryptoRequest>) => void) | null;
  postMessage(message: CryptoResponse): void;
};

ctx.onmessage = (event: MessageEvent<CryptoRequest>) => {
  const request = event.data;
  try {
    ctx.postMessage({ id: request.id, ok: true, payload: handle(request) });
  } catch (error) {
    ctx.postMessage({
      id: request.id,
      ok: false,
      error: error instanceof Error ? error.message : "error_desconocido",
    });
  }
};

function handle(request: CryptoRequest): unknown {
  switch (request.type) {
    case "derive-kek":
      return { kekB64: toB64(deriveKEK(request.passphrase, fromB64(request.saltB64))) };
    case "generate-identity": {
      const identity = generateIdentity();
      return {
        privateKeyB64: toB64(identity.privateKey),
        publicKeyB64: toB64(identity.publicKey),
        fingerprint: fingerprint(identity.publicKey),
      };
    }
    case "create-identity":
      return createIdentity(request.passphrase);
    case "unlock-identity":
      return unlockIdentity(request.passphrase, {
        saltB64: request.saltB64,
        identitySealedB64: request.identitySealedB64,
      });
    case "seal":
      return buildEnvelope(
        request.plaintext,
        {
          keyId: request.recipientFingerprint,
          publicKey: fromB64(request.recipientPublicKeyB64),
        },
        { senderFingerprint: request.senderFingerprint },
      );
    case "open":
      return {
        plaintext: new TextDecoder().decode(
          openEnvelope(request.envelope, fromB64(request.privateKeyB64)),
        ),
      };
  }
  throw new Error("peticion_desconocida");
}
