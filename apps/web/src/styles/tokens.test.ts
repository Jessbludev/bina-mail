import { describe, expect, it } from "vitest";
import {
  buttonClasses,
  emptyState,
  fieldInputClasses,
  layout,
  navLinkClasses,
  spinner,
  statusClasses,
} from "./tokens";

describe("tokens de diseno", () => {
  it("button: combina base y variante con foco visible", () => {
    const primary = buttonClasses("primary");
    expect(primary).toContain("bg-sky-600");
    expect(primary).toContain("focus-visible:ring-2");
  });

  it("button: expone ghost y danger", () => {
    expect(buttonClasses("ghost")).toContain("hover:bg-neutral-800");
    expect(buttonClasses("danger")).toContain("bg-red-700");
  });

  it("spinner: respeta prefers-reduced-motion", () => {
    expect(spinner).toContain("motion-safe:animate-spin");
  });

  it("field: marca invalid solo cuando procede", () => {
    expect(fieldInputClasses(true)).toContain("border-red-500");
    expect(fieldInputClasses(false)).not.toContain("border-red-500");
  });

  it("status: expone info, success y error", () => {
    expect(statusClasses("success")).toContain("border-emerald-700");
    expect(statusClasses("error")).toContain("border-red-800");
  });

  it("nav: distingue el enlace activo", () => {
    expect(navLinkClasses(true)).toContain("text-sky-300");
    expect(navLinkClasses(false)).not.toContain("text-sky-300");
  });

  it("layout: fondo oscuro por defecto (mobile-first)", () => {
    expect(layout.page).toContain("bg-neutral-950");
    expect(layout.container).toContain("max-w-3xl");
  });

  it("emptyState: centro y texto legible", () => {
    expect(emptyState.box).toContain("flex-col");
    expect(emptyState.title).toContain("text-neutral-200");
  });
});
