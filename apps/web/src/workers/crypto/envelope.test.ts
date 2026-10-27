import { describe, expect, it } from "vitest";
import { buildEnvelope, openEnvelope } from "./envelope";
import { fingerprint } from "./blake3";
import { generateIdentity } from "./x25519";
import { fromB64, toB64 } from "./b64";

describe("sobre E2EE", () => {
  const bob = generateIdentity();
  const bobId = fingerprint(bob.publicKey);
  const recipient = { keyId: bobId, publicKey: bob.publicKey };

  it("ida y vuelta entre dos usuarios", () => {
    const envelope = buildEnvelope("hola bob", recipient, {
      senderFingerprint: "1111111111111111111111111111111111111111111111111111111111111111",
    });
    const plaintext = openEnvelope(envelope, bob.privateKey);
    expect(new TextDecoder().decode(plaintext)).toBe("hola bob");
    expect(envelope.sender_fingerprint).toBe(
      "1111111111111111111111111111111111111111111111111111111111111111",
    );
  });

  it("rechaza la apertura con clave ajena", () => {
    const mallory = generateIdentity();
    const envelope = buildEnvelope("hola bob", recipient);
    expect(() => openEnvelope(envelope, mallory.privateKey)).toThrow();
  });

  it("detecta manipulacion del ciphertext", () => {
    const envelope = buildEnvelope("hola bob", recipient);
    const bytes = fromB64(envelope.ciphertext);
    bytes[0] = (bytes[0] ?? 0) ^ 1;
    const tampered = { ...envelope, ciphertext: toB64(bytes) };
    expect(() => openEnvelope(tampered, bob.privateKey)).toThrow();
  });
});
