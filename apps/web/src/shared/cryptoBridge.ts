/**
 * Puente con el Worker criptografico: promesas correladas por id.
 * El Worker se crea de forma perezosa para que las pruebas de stores y
 * vistas puedan mockear este modulo sin instanciarlo (jsdom no implementa
 * Worker). Ninguna funcion registra contenido (politica zero-log).
 */
import type { CryptoInput, CryptoRequest, CryptoResponse } from "../workers/crypto/messages";

interface Pending {
  resolve: (payload: unknown) => void;
  reject: (reason: Error) => void;
}

let worker: Worker | null = null;
const pending = new Map<string, Pending>();
let sequence = 0;

function ensureWorker(): Worker {
  if (worker === null) {
    worker = new Worker(new URL("../workers/crypto/worker.ts", import.meta.url), {
      type: "module",
    });
    worker.onmessage = (event: MessageEvent<CryptoResponse>) => {
      const response = event.data;
      const entry = pending.get(response.id);
      if (!entry) {
        return;
      }
      pending.delete(response.id);
      if (response.ok) {
        entry.resolve(response.payload);
      } else {
        entry.reject(new Error(response.error));
      }
    };
  }
  return worker;
}

export function cryptoRequest<T>(request: CryptoInput): Promise<T> {
  const id = "req-" + String((sequence += 1));
  return new Promise<T>((resolve, reject) => {
    pending.set(id, {
      resolve: (payload) => resolve(payload as T),
      reject,
    });
    ensureWorker().postMessage({ ...request, id } as CryptoRequest);
  });
}
