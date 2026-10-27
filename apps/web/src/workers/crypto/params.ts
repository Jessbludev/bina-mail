/**
 * Parametros criptograficos de Bina Mail (ADR-004).
 * Cambiar cualquier valor exige incrementar PARAMS_VERSION y reenvolver claves.
 */
export const PARAMS_VERSION = 1;

export const ARGON2ID_PARAMS = {
  t: 3,
  m: 65536,
  p: 4,
  dkLen: 32,
} as const;

export const KEY_SIZE = 32;
export const XCHACHA_NONCE_SIZE = 24;
