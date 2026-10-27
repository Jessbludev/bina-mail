import { describe, expect, it } from "vitest";
import { deriveKEK } from "./argon2id";

describe("deriveKEK (Argon2id, ADR-004)", () => {
  const salt = new Uint8Array(16).fill(7);

  it("es determinista con la misma sal y produce 32 bytes", () => {
    const a = deriveKEK("frase-secreta", salt);
    const b = deriveKEK("frase-secreta", salt);
    expect(a).toHaveLength(32);
    expect(b).toEqual(a);
  });

  it("produce salidas distintas con sal distinta", () => {
    const a = deriveKEK("frase-secreta", salt);
    const b = deriveKEK("frase-secreta", new Uint8Array(16).fill(9));
    expect(b).not.toEqual(a);
  });

  it("rechaza sales menores de 16 bytes", () => {
    expect(() => deriveKEK("frase-secreta", new Uint8Array(8))).toThrow();
  });
});
