/**
 * Cliente tipado del gateway BFF y de Supabase Auth.
 * Ninguna funcion registra contenido de mensajes (politica zero-log, ADR-005).
 */
import { loadConfig } from "./config";
import type { EnvelopeJson } from "../workers/crypto/envelope";

const { gatewayUrl, supabaseUrl, supabaseAnonKey } = loadConfig();

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
  ) {
    super(`API ${status}: ${code}`);
  }
}

async function requestJson<T>(url: string, init: RequestInit): Promise<T> {
  const response = await fetch(url, init);
  if (!response.ok) {
    let code = "error";
    try {
      const body: unknown = await response.json();
      if (typeof body === "object" && body !== null && "error" in body) {
        const errorField = (body as { error: unknown }).error;
        if (typeof errorField === "string") {
          code = errorField;
        }
      }
    } catch {
      // Cuerpo sin JSON: se conserva el codigo generico.
    }
    throw new ApiError(response.status, code);
  }
  return (await response.json()) as T;
}

export interface AuthSession {
  readonly token: string;
  readonly expires_at: string;
}

export async function signUp(email: string, password: string): Promise<void> {
  await requestJson<unknown>(`${supabaseUrl}/auth/v1/signup`, {
    method: "POST",
    headers: jsonHeaders({ apikey: supabaseAnonKey }),
    body: JSON.stringify({ email, password }),
  });
}

export async function signIn(email: string, password: string): Promise<string> {
  const result = await requestJson<{ access_token: string }>(
    `${supabaseUrl}/auth/v1/token?grant_type=password`,
    {
      method: "POST",
      headers: jsonHeaders({ apikey: supabaseAnonKey }),
      body: JSON.stringify({ email, password }),
    },
  );
  return result.access_token;
}

export function exchangeSession(supabaseJwt: string): Promise<AuthSession> {
  return requestJson<AuthSession>(`${gatewayUrl}/api/v1/auth/session`, {
    method: "POST",
    headers: jsonHeaders(),
    body: JSON.stringify({ supabase_jwt: supabaseJwt }),
  });
}

export interface PublicKeyRecord {
  readonly user_id: string;
  readonly x25519_public_key: string;
  readonly ed25519_signing_public_key: string;
  readonly fingerprint: string;
  readonly active: boolean;
}

export interface PublishedKey {
  readonly user_id: string;
  readonly x25519_public_key: string;
  readonly ed25519_signing_public_key: string;
  readonly fingerprint: string;
}

export function publishKey(sessionToken: string, key: PublishedKey): Promise<{ ok: boolean }> {
  return requestJson<{ ok: boolean }>(`${gatewayUrl}/api/v1/keys`, {
    method: "POST",
    headers: jsonHeaders({ Authorization: `Bearer ${sessionToken}` }),
    body: JSON.stringify(key),
  });
}

export function lookupKey(sessionToken: string, fingerprint: string): Promise<PublicKeyRecord> {
  return requestJson<PublicKeyRecord>(
    `${gatewayUrl}/api/v1/keys/${encodeURIComponent(fingerprint)}`,
    {
      method: "GET",
      headers: jsonHeaders({ Authorization: `Bearer ${sessionToken}` }),
    },
  );
}

export function storeEnvelope(
  sessionToken: string,
  envelope: EnvelopeJson,
): Promise<{ id: string }> {
  return requestJson<{ id: string }>(`${gatewayUrl}/api/v1/envelopes`, {
    method: "POST",
    headers: jsonHeaders({ Authorization: `Bearer ${sessionToken}` }),
    body: JSON.stringify(envelope),
  });
}

export function listEnvelopes(
  sessionToken: string,
  recipientFingerprint: string,
  limit = 50,
): Promise<EnvelopeJson[]> {
  const query = new URLSearchParams({
    recipient_fingerprint: recipientFingerprint,
    limit: String(limit),
  });
  return requestJson<EnvelopeJson[]>(`${gatewayUrl}/api/v1/envelopes?${query}`, {
    method: "GET",
    headers: jsonHeaders({ Authorization: `Bearer ${sessionToken}` }),
  });
}

function jsonHeaders(extra: Record<string, string>): Record<string, string> {
  return { "Content-Type": "application/json", ...extra };
}
