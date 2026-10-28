import { defineStore } from "pinia";
import {
  exchangeSession,
  publishKey,
  revokeSession,
  signIn,
  signUp,
} from "../shared/api";
import { cryptoRequest } from "../shared/cryptoBridge";
import { readVault, writeVault, type VaultRecord } from "../shared/vault";

export type SessionPhase = "signed-out" | "signed-in" | "unlocked";

/** Respuesta del Worker para create-identity y unlock-identity. */
export interface IdentityResponse {
  readonly encryptionPrivateKeyB64: string;
  readonly signingPrivateKeyB64: string;
  readonly x25519PublicKeyB64: string;
  readonly ed25519PublicKeyB64: string;
  readonly fingerprint: string;
  readonly saltB64?: string;
  readonly identitySealedB64?: string;
}

/**
 * Estado de sesion del cliente. Las claves privadas viven solo en RAM;
 * el vault persistente guarda unicamente el bloque sellado y lo publico.
 */
export const useSessionStore = defineStore("session", {
  state: (): {
    phase: SessionPhase;
    sessionToken: string;
    expiresAt: string;
    supabaseUserId: string;
    vault: VaultRecord | null;
    fingerprint: string;
    x25519PublicKeyB64: string;
    ed25519PublicKeyB64: string;
    encryptionPrivateKeyB64: string;
    signingPrivateKeyB64: string;
  } => ({
    phase: "signed-out",
    sessionToken: "",
    expiresAt: "",
    supabaseUserId: "",
    vault: null,
    fingerprint: "",
    x25519PublicKeyB64: "",
    ed25519PublicKeyB64: "",
    encryptionPrivateKeyB64: "",
    signingPrivateKeyB64: "",
  }),

  getters: {
    tieneIdentidad: (state) => state.vault !== null,
  },

  actions: {
    async iniciarSesion(email: string, password: string, registro: boolean): Promise<void> {
      if (registro) {
        await signUp(email, password);
      }
      const jwt = await signIn(email, password);
      this.supabaseUserId = subjectFromJwt(jwt);
      const sesion = await exchangeSession(jwt);
      this.sessionToken = sesion.token;
      this.expiresAt = sesion.expires_at;
      this.vault = readVault();
      this.phase = "signed-in";
    },

    async crearIdentidad(passphrase: string): Promise<void> {
      if (this.phase !== "signed-in" || !this.sessionToken) {
        throw new Error("sesion_requerida");
      }
      const identity = await cryptoRequest<IdentityResponse>({
        type: "create-identity",
        passphrase,
      });
      if (!identity.saltB64 || !identity.identitySealedB64) {
        throw new Error("identidad_invalida");
      }
      const record: VaultRecord = {
        saltB64: identity.saltB64,
        identitySealedB64: identity.identitySealedB64,
        x25519PublicKeyB64: identity.x25519PublicKeyB64,
        ed25519PublicKeyB64: identity.ed25519PublicKeyB64,
        fingerprint: identity.fingerprint,
      };
      await publishKey(this.sessionToken, {
        user_id: this.supabaseUserId,
        x25519_public_key: identity.x25519PublicKeyB64,
        ed25519_signing_public_key: identity.ed25519PublicKeyB64,
        fingerprint: identity.fingerprint,
      });
      writeVault(record);
      this.vault = record;
      this.asignarIdentidad(identity);
      this.phase = "unlocked";
    },

    async desbloquear(passphrase: string): Promise<void> {
      if (this.phase === "signed-out" || !this.vault) {
        throw new Error("identidad_ausente");
      }
      const identity = await cryptoRequest<IdentityResponse>({
        type: "unlock-identity",
        passphrase,
        saltB64: this.vault.saltB64,
        identitySealedB64: this.vault.identitySealedB64,
      });
      if (identity.fingerprint !== this.vault.fingerprint) {
        throw new Error("fingerprint_no_coincide");
      }
      this.asignarIdentidad(identity);
      this.phase = "unlocked";
    },

    lock(): void {
      this.encryptionPrivateKeyB64 = "";
      this.signingPrivateKeyB64 = "";
      this.phase = this.sessionToken ? "signed-in" : "signed-out";
    },

    async cerrarSesion(): Promise<void> {
      if (this.sessionToken) {
        try {
          await revokeSession(this.sessionToken);
        } catch {
          // Fail-soft: la sesion opaca expira sola a los 30 minutos.
        }
      }
      this.phase = "signed-out";
      this.sessionToken = "";
      this.expiresAt = "";
      this.supabaseUserId = "";
      this.fingerprint = "";
      this.x25519PublicKeyB64 = "";
      this.ed25519PublicKeyB64 = "";
      this.encryptionPrivateKeyB64 = "";
      this.signingPrivateKeyB64 = "";
      this.vault = readVault();
    },

    asignarIdentidad(identity: IdentityResponse): void {
      this.fingerprint = identity.fingerprint;
      this.x25519PublicKeyB64 = identity.x25519PublicKeyB64;
      this.ed25519PublicKeyB64 = identity.ed25519PublicKeyB64;
      this.encryptionPrivateKeyB64 = identity.encryptionPrivateKeyB64;
      this.signingPrivateKeyB64 = identity.signingPrivateKeyB64;
    },
  },
});

/**
 * Sujeto del JWT de Supabase (identidad del usuario en la plataforma).
 * Sin verificacion criptografica: la verificacion la hace el backend con
 * el secreto HS256; el cliente solo lee su propio identificador.
 */
function subjectFromJwt(jwt: string): string {
  const segments = jwt.split(".");
  if (segments.length < 2) {
    throw new Error("jwt_malformado");
  }
  const normalized = segments[1].replace(/-/g, "+").replace(/_/g, "/");
  const json = atob(normalized);
  const payload = JSON.parse(json) as { sub?: unknown };
  if (typeof payload.sub !== "string" || payload.sub === "") {
    throw new Error("jwt_sin_sujeto");
  }
  return payload.sub;
}
