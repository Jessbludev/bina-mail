import { defineStore } from "pinia";
import { listEnvelopes, type MailboxEnvelope } from "../shared/api";
import { cryptoRequest } from "../shared/cryptoBridge";
import type { EnvelopeJson } from "../workers/crypto/envelope";
import { PARAMS_VERSION } from "../workers/crypto/params";

export type MailboxStatus = "vacio" | "cargando" | "listo" | "error";

/**
 * Buzon del usuario. Los sobres viajan cifrados y se descifran bajo demanda
 * en el Worker; los textos claros viven unicamente en RAM (zero-knowledge).
 */
export const useMailboxStore = defineStore("mailbox", {
  state: (): {
    items: MailboxEnvelope[];
    status: MailboxStatus;
    error: string;
    abiertos: Record<string, string>;
  } => ({
    items: [],
    status: "vacio",
    error: "",
    abiertos: {},
  }),

  getters: {
    porId: (state) => (id: string): MailboxEnvelope | null =>
      state.items.find((item) => item.id === id) ?? null,
  },

  actions: {
    async refrescar(sessionToken: string, fingerprint: string): Promise<void> {
      this.status = "cargando";
      this.error = "";
      try {
        this.items = await listEnvelopes(sessionToken, fingerprint);
        this.status = "listo";
      } catch (error) {
        this.items = [];
        this.status = "error";
        this.error = error instanceof Error ? error.message : "error_red";
      }
    },

    async abrir(id: string, privateKeyB64: string): Promise<string> {
      const cache = this.abiertos[id];
      if (cache !== undefined) {
        return cache;
      }
      const row = this.porId(id);
      if (row === null) {
        throw new Error("sobre_inexistente");
      }
      const plaintext = await cryptoRequest<string>({
        type: "open",
        envelope: aEnvelopeJson(row),
        privateKeyB64,
      });
      this.abiertos[id] = plaintext; // Cache en RAM: nunca se persiste.
      return plaintext;
    },

    limpiar(): void {
      this.items = [];
      this.abiertos = {};
      this.status = "vacio";
      this.error = "";
    },
  },
});

/** Traduce la fila del backend al contrato EnvelopeJson del Worker. */
function aEnvelopeJson(row: MailboxEnvelope): EnvelopeJson {
  return {
    version: PARAMS_VERSION,
    sender_fingerprint: row.sender_fingerprint === "" ? undefined : row.sender_fingerprint,
    recipient_fingerprint: row.recipient_fingerprint,
    nonce: row.nonce,
    ciphertext: row.ciphertext,
    wraps: row.wraps,
    blake3_fingerprint: row.blake3_fingerprint,
    created_at_unix: Math.floor(Date.parse(row.created_at) / 1000),
  };
}
