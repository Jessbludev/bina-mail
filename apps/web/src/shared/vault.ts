/**
 * Vault local del dispositivo: guarda solo el bloque sellado y las claves
 * publicas. La frase de contrasena y las claves privadas nunca se escriben
 * aqui (criterio zero-knowledge, MAP: "ninguna clave privada sale del
 * dispositivo").
 */
export interface VaultRecord {
  readonly saltB64: string;
  readonly identitySealedB64: string;
  readonly x25519PublicKeyB64: string;
  readonly ed25519PublicKeyB64: string;
  readonly fingerprint: string;
}

const VAULT_KEY = "bina.v1.identity";

function isRecord(value: unknown): value is VaultRecord {
  if (typeof value !== "object" || value === null) {
    return false;
  }
  const record = value as Record<string, unknown>;
  return (
    typeof record.saltB64 === "string" &&
    typeof record.identitySealedB64 === "string" &&
    typeof record.x25519PublicKeyB64 === "string" &&
    typeof record.ed25519PublicKeyB64 === "string" &&
    typeof record.fingerprint === "string" &&
    record.saltB64 !== "" &&
    record.identitySealedB64 !== "" &&
    record.x25519PublicKeyB64 !== "" &&
    record.ed25519PublicKeyB64 !== "" &&
    record.fingerprint !== ""
  );
}

export function readVault(): VaultRecord | null {
  try {
    const raw = window.localStorage.getItem(VAULT_KEY);
    if (!raw) {
      return null;
    }
    const parsed: unknown = JSON.parse(raw);
    return isRecord(parsed) ? parsed : null;
  } catch {
    return null;
  }
}

export function writeVault(record: VaultRecord): void {
  window.localStorage.setItem(VAULT_KEY, JSON.stringify(record));
}

export function clearVault(): void {
  window.localStorage.removeItem(VAULT_KEY);
}
