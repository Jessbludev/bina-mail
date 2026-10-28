import { beforeEach, describe, expect, it } from "vitest";
import { IndiceBusqueda, normalizar } from "./search";

describe("normalizar", () => {
  it("pasa a minusculas y quita diacriticos", () => {
    expect(normalizar("Café con Leche")).toBe("cafe con leche");
    expect(normalizar("SOBRE")).toBe("sobre");
  });
});

describe("IndiceBusqueda", () => {
  let indice: IndiceBusqueda;

  beforeEach(() => {
    indice = new IndiceBusqueda();
  });

  it("encuentra textos con terminos sin diacriticos", () => {
    indice.indexar("sobre-1", "El café está listo");
    indice.indexar("sobre-2", "Reunion de manana");
    expect(indice.buscar("CAFE")).toEqual(["sobre-1"]);
    expect(indice.buscar("manana")).toEqual(["sobre-2"]);
  });

  it("devuelve vacio con termino en blanco", () => {
    indice.indexar("sobre-1", "texto");
    expect(indice.buscar("")).toEqual([]);
    expect(indice.buscar("   ")).toEqual([]);
  });

  it("reemplaza el texto al reindexar y limpia por completo", () => {
    indice.indexar("sobre-1", "primera version");
    indice.indexar("sobre-1", "segunda version");
    expect(indice.buscar("primera")).toEqual([]);
    expect(indice.buscar("segunda")).toEqual(["sobre-1"]);
    indice.limpiar();
    expect(indice.tamano()).toBe(0);
    expect(indice.contiene("sobre-1")).toBe(false);
  });

  it("cumple el SLO de busqueda con 500 sobres (p95 < 50 ms por consulta)", () => {
    const palabras = ["alpha", "beta", "gamma", "delta"];
    for (let i = 0; i < 500; i += 1) {
      indice.indexar("sobre-" + i, "mensaje " + palabras[i % 4] + " numero " + i);
    }
    const tiempos: number[] = [];
    for (let i = 0; i < 20; i += 1) {
      const inicio = performance.now();
      indice.buscar(i % 2 === 0 ? "beta" : "no-existe-" + i);
      tiempos.push(performance.now() - inicio);
    }
    tiempos.sort((a, b) => a - b);
    const p95 = tiempos[Math.floor(tiempos.length * 0.95) - 1];
    expect(p95).toBeLessThan(50);
  });
});
