import { describe, expect, it } from "vitest";
import { randomBytes } from "@noble/ciphers/utils";
import { openBox, sealBox } from "./aead";
import { fingerprint } from "./blake3";
import { generateIdentity, sharedSecret } from "./x25519";

describe("fingerprint (BLAKE3)", () => {
  it("coincide con el vector oficial de la entrada vacia", () => {
    expect(fingerprint(new Uint8Array(0))).toBe(
      "af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9a93cae41f3262",
    );
  });
});

describe("aead (XChaCha20-Poly1305)", () => {
  it("cifra y descifra un mensaje de ida y vuelta", () => {
    const key = randomBytes(32);
    const message = new TextEncoder().encode("contenido privado");
    const box = sealBox(key, message);
    expect(openBox(key, box)).toEqual(message);
  });

  it("rechaza un ciphertext manipulado", () => {
    const key = randomBytes(32);
    const box = sealBox(key, new TextEncoder().encode("mensaje"));
    const tampered = box.ciphertext.slice();
    tampered[0] = (tampered[0] ?? 0) ^ 1;
    expect(() => openBox(key, { nonce: box.nonce, ciphertext: tampered })).toThrow();
  });

  it("rechaza una clave de tamano invalido", () => {
    expect(() => sealBox(new Uint8Array(16), new Uint8Array(1))).toThrow();
  });
});

describe("x25519", () => {
  it("el secreto compartido es simetrico", () => {
    const alice = generateIdentity();
    const bob = generateIdentity();
    expect(sharedSecret(alice.privateKey, bob.publicKey)).toEqual(
      sharedSecret(bob.privateKey, alice.publicKey),
    );
  });
});
