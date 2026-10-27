/**
 * Fingerprint BLAKE3 en hex minuscula (64 caracteres).
 * Vector oficial de la entrada vacia cubierto en primitives.test.ts.
 */
import { blake3 } from "@noble/hashes/blake3";
import { bytesToHex } from "@noble/hashes/utils";

export function fingerprint(data: Uint8Array): string {
  return bytesToHex(blake3(data));
}
