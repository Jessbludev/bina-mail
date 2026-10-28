import { describe, expect, it } from "vitest";
import { createIdentity, unlockIdentity, type IdentityBundle } from "./identity";

const FRASE = "frase de prueba segura";

describe("identidad sellada", () => {
  it("crea un paquete sellado cuya clave privada no aparece en claro", () => {
    const identity = createIdentity(FRASE);
    expect(identity.identitySealedB64).not.toContain(identity.encryptionPrivateKeyB64);
    expect(identity.identitySealedB64).not.toContain(identity.signingPrivateKeyB64);
    expect(identity.fingerprint).toHaveLength(64);
    expect(identity.x25519PublicKeyB64).not.toHaveLength(0);
    expect(identity.ed25519PublicKeyB64).not.toHaveLength(0);
  });

  it("desbloquea con la frase correcta y reproduce el fingerprint", () => {
    const identity: IdentityBundle = createIdentity(FRASE);
    const unlocked = unlockIdentity(FRASE, {
      saltB64: identity.saltB64,
      identitySealedB64: identity.identitySealedB64,
    });
    expect(unlocked.fingerprint).toBe(identity.fingerprint);
    expect(unlocked.encryptionPrivateKeyB64).toBe(identity.encryptionPrivateKeyB64);
    expect(unlocked.signingPrivateKeyB64).toBe(identity.signingPrivateKeyB64);
  });

  it("rechaza una frase incorrecta", () => {
    const identity = createIdentity(FRASE);
    expect(() =>
      unlockIdentity("frase completamente distinta", {
        saltB64: identity.saltB64,
        identitySealedB64: identity.identitySealedB64,
      }),
    ).toThrowError("frase_invalida");
  });

  it("rechaza un paquete manipulado", () => {
    const identity = createIdentity(FRASE);
    const sello = atob(identity.identitySealedB64);
    const manipulado = btoa(sello.slice(0, -2) + "zz");
    expect(() =>
      unlockIdentity(FRASE, {
        saltB64: identity.saltB64,
        identitySealedB64: manipulado,
      }),
    ).toThrowError("frase_invalida");
  });

  it("rechaza frases demasiado cortas", () => {
    expect(() => createIdentity("corta")).toThrowError("frase_corta");
  });
});
