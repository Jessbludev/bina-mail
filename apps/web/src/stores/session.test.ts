import { afterAll, beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { clearVault, writeVault, type VaultRecord } from "../shared/vault";

vi.mock("../shared/api", () => ({
  ApiError: class extends Error {
    constructor(
      public readonly status: number,
      public readonly code: string,
    ) {
      super("api");
    }
  },
  signUp: vi.fn(),
  signIn: vi.fn(),
  exchangeSession: vi.fn(),
  revokeSession: vi.fn(),
  publishKey: vi.fn(),
  lookupKey: vi.fn(),
  storeEnvelope: vi.fn(),
  listEnvelopes: vi.fn(),
}));

vi.mock("../shared/cryptoBridge", () => ({
  cryptoRequest: vi.fn(),
}));

import { exchangeSession, publishKey, revokeSession, signIn, signUp } from "../shared/api";
import { cryptoRequest } from "../shared/cryptoBridge";
import { useSessionStore } from "./session";

const JWT = "h." + btoa(JSON.stringify({ sub: "usuario-1" })) + ".s";

const VAULT: VaultRecord = {
  saltB64: "c2FsAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
  identitySealedB64: "c2VsbG8AAAAAAAAAAAAAAAAAAAAAAAAAAAA",
  x25519PublicKeyB64: "A" + "Q".repeat(43),
  ed25519PublicKeyB64: "B" + "Q".repeat(43),
  fingerprint: "f".repeat(64),
};

const IDENTITY = {
  encryptionPrivateKeyB64: "enc-priv",
  signingPrivateKeyB64: "sig-priv",
  x25519PublicKeyB64: "enc-pub",
  ed25519PublicKeyB64: "sig-pub",
  fingerprint: VAULT.fingerprint,
};

function mockIdentity(create: boolean) {
  return vi.mocked(cryptoRequest).mockResolvedValue(
    create
      ? { ...IDENTITY, saltB64: VAULT.saltB64, identitySealedB64: VAULT.identitySealedB64 }
      : IDENTITY,
  );
}

beforeEach(() => {
  setActivePinia(createPinia());
  window.localStorage.clear();
  vi.clearAllMocks();
  vi.mocked(signIn).mockResolvedValue(JWT);
  vi.mocked(signUp).mockResolvedValue(undefined);
  vi.mocked(exchangeSession).mockResolvedValue({ token: "tok-1", expires_at: "2099-01-01T00:00:00Z" });
  vi.mocked(revokeSession).mockResolvedValue(undefined);
  vi.mocked(publishKey).mockResolvedValue({ ok: true });
});

afterAll(() => {
  clearVault();
});

describe("sesion del cliente", () => {
  it("inicia sesion con credenciales validas y carga el vault del dispositivo", async () => {
    writeVault(VAULT);
    const session = useSessionStore();
    await session.iniciarSesion("usuario@ejemplo.com", "clave-segura", false);
    expect(session.phase).toBe("signed-in");
    expect(session.sessionToken).toBe("tok-1");
    expect(session.supabaseUserId).toBe("usuario-1");
    expect(session.vault?.fingerprint).toBe(VAULT.fingerprint);
    expect(signUp).not.toHaveBeenCalled();
  });

  it("registra la cuenta antes de iniciar sesion", async () => {
    const session = useSessionStore();
    await session.iniciarSesion("nuevo@ejemplo.com", "clave-segura", true);
    expect(signUp).toHaveBeenCalledWith("nuevo@ejemplo.com", "clave-segura");
    expect(signIn).toHaveBeenCalled();
  });

  it("crea la identidad, publica las claves y guarda el vault sellado", async () => {
    const session = useSessionStore();
    await session.iniciarSesion("usuario@ejemplo.com", "clave-segura", false);
    mockIdentity(true);
    await session.crearIdentidad("frase-segura-1");
    expect(publishKey).toHaveBeenCalledWith(
      "tok-1",
      expect.objectContaining({ user_id: "usuario-1", fingerprint: VAULT.fingerprint }),
    );
    expect(session.phase).toBe("unlocked");
    expect(session.encryptionPrivateKeyB64).toBe("enc-priv");
    expect(window.localStorage.getItem("bina.v1.identity")).not.toBeNull();
  });

  it("desbloquea con la frase correcta y reproduce el fingerprint", async () => {
    writeVault(VAULT);
    const session = useSessionStore();
    await session.iniciarSesion("usuario@ejemplo.com", "clave-segura", false);
    mockIdentity(false);
    await session.desbloquear("frase-segura-1");
    expect(session.phase).toBe("unlocked");
    expect(session.fingerprint).toBe(VAULT.fingerprint);
  });

  it("rechaza el desbloqueo si el fingerprint no coincide con el vault", async () => {
    writeVault(VAULT);
    const session = useSessionStore();
    await session.iniciarSesion("usuario@ejemplo.com", "clave-segura", false);
    vi.mocked(cryptoRequest).mockResolvedValue({ ...IDENTITY, fingerprint: "otro" });
    await expect(session.desbloquear("frase-segura-1")).rejects.toThrowError("fingerprint_no_coincide");
    expect(session.encryptionPrivateKeyB64).toBe("");
  });

  it("lock limpia las claves privadas de RAM y conserva la sesion", async () => {
    writeVault(VAULT);
    const session = useSessionStore();
    await session.iniciarSesion("usuario@ejemplo.com", "clave-segura", false);
    mockIdentity(false);
    await session.desbloquear("frase-segura-1");
    session.lock();
    expect(session.phase).toBe("signed-in");
    expect(session.encryptionPrivateKeyB64).toBe("");
    expect(session.signingPrivateKeyB64).toBe("");
  });

  it("cerrar sesion revoca el token y conserva el vault del dispositivo", async () => {
    writeVault(VAULT);
    const session = useSessionStore();
    await session.iniciarSesion("usuario@ejemplo.com", "clave-segura", false);
    await session.cerrarSesion();
    expect(revokeSession).toHaveBeenCalledWith("tok-1");
    expect(session.phase).toBe("signed-out");
    expect(session.sessionToken).toBe("");
    expect(session.vault?.fingerprint).toBe(VAULT.fingerprint);
  });
});
