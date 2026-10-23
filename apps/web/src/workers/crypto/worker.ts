import type { CryptoRequest, CryptoResponse } from "./messages";

/**
 * Worker criptográfico de Bina Mail (Fase 1: esqueleto).
 * La implementación real (Argon2id, XChaCha20-Poly1305, BLAKE3) llega en Fase 2
 * con vectores de test obligatorios (ADR-004).
 */
const ctx = self as unknown as {
  onmessage: ((event: MessageEvent<CryptoRequest>) => void) | null;
  postMessage(message: CryptoResponse): void;
};

ctx.onmessage = (event) => {
  const request = event.data;
  const response: CryptoResponse = {
    id: request.id,
    ok: false,
    error: "not-implemented: Fase 2 (ADR-004)",
  };
  ctx.postMessage(response);
};
