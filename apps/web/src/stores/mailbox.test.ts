import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { useMailboxStore } from "./mailbox";
import { listEnvelopes } from "../shared/api";
import { cryptoRequest } from "../shared/cryptoBridge";

vi.mock("../shared/api", () => ({ listEnvelopes: vi.fn() }));
vi.mock("../shared/cryptoBridge", () => ({ cryptoRequest: vi.fn() }));

const filaBase = {
  id: "sobre-1",
  sender_fingerprint: "",
  recipient_fingerprint: "fp-destinatario",
  nonce: "bm9uY2U=",
  ciphertext: "Y2lwaGVydGV4dG8=",
  wraps: [
    {
      recipient_key_id: "fp-destinatario",
      ephemeral_public_key: "ZXBoZW1lcmFs",
      wrapped_key: "d3JhcGVk",
    },
  ],
  blake3_fingerprint: "blake3-sobre-1",
  size_bytes: 12,
  created_at: "2026-10-28T12:00:00Z",
};

beforeEach(() => {
  setActivePinia(createPinia());
  vi.mocked(listEnvelopes).mockReset();
  vi.mocked(cryptoRequest).mockReset();
});

describe("useMailboxStore", () => {
  it("refresca la bandeja y queda listo", async () => {
    vi.mocked(listEnvelopes).mockResolvedValue([filaBase]);
    const store = useMailboxStore();
    await store.refrescar("token-1", "fp-destinatario");
    expect(store.status).toBe("listo");
    expect(store.items).toHaveLength(1);
    expect(listEnvelopes).toHaveBeenCalledWith("token-1", "fp-destinatario");
  });

  it("deja el estado en error si el listado falla", async () => {
    vi.mocked(listEnvelopes).mockRejectedValue(new Error("red"));
    const store = useMailboxStore();
    await store.refrescar("token-1", "fp-destinatario");
    expect(store.status).toBe("error");
    expect(store.items).toHaveLength(0);
    expect(store.error).toBe("red");
  });

  it("abre un sobre pasando el contrato exacto al Worker y cachea el texto", async () => {
    vi.mocked(listEnvelopes).mockResolvedValue([filaBase]);
    vi.mocked(cryptoRequest).mockResolvedValue("texto claro");
    const store = useMailboxStore();
    await store.refrescar("token-1", "fp-destinatario");
    const texto = await store.abrir("sobre-1", "privada");
    expect(texto).toBe("texto claro");
    expect(cryptoRequest).toHaveBeenCalledTimes(1);
    const entrada = vi.mocked(cryptoRequest).mock.calls[0][0] as unknown as {
      type: string;
      envelope: {
        version: number;
        created_at_unix: number;
        sender_fingerprint?: string;
      };
    };
    expect(entrada.type).toBe("open");
    expect(entrada.envelope.version).toBe(1);
    expect(entrada.envelope.created_at_unix).toBe(
      Math.floor(Date.parse("2026-10-28T12:00:00Z") / 1000),
    );
    expect(entrada.envelope.sender_fingerprint).toBeUndefined();
    const repetido = await store.abrir("sobre-1", "privada");
    expect(repetido).toBe("texto claro");
    expect(cryptoRequest).toHaveBeenCalledTimes(1);
  });

  it("preserva el remitente cuando el sobre lo declara", async () => {
    vi.mocked(listEnvelopes).mockResolvedValue([
      { ...filaBase, sender_fingerprint: "fp-remitente" },
    ]);
    vi.mocked(cryptoRequest).mockResolvedValue("texto");
    const store = useMailboxStore();
    await store.refrescar("token-1", "fp-destinatario");
    await store.abrir("sobre-1", "privada");
    const entrada = vi.mocked(cryptoRequest).mock.calls[0][0] as unknown as {
      envelope: { sender_fingerprint?: string };
    };
    expect(entrada.envelope.sender_fingerprint).toBe("fp-remitente");
  });

  it("rechaza abrir un sobre que no esta en la bandeja", async () => {
    const store = useMailboxStore();
    await expect(store.abrir("desconocido", "privada")).rejects.toThrow("sobre_inexistente");
  });
});
