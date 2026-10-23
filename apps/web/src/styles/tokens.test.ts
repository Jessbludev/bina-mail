import { describe, expect, it } from "vitest";
import { buttonClasses } from "./tokens";

describe("buttonClasses", () => {
  it("combina base y variante tipada", () => {
    expect(buttonClasses("primary")).toContain("bg-sky-600");
    expect(buttonClasses("primary")).toContain("rounded-md");
  });

  it("expone la variante ghost", () => {
    expect(buttonClasses("ghost")).toContain("hover:bg-neutral-800");
  });
});
