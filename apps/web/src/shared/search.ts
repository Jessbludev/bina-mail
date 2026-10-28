/**
 * Busqueda local de Bina Mail (ADR-009): indice en RAM sobre los sobres que
 * el usuario ya descifro en esta sesion. Nunca se persiste, nunca sale del
 * dispositivo y el servidor no participa en la busqueda.
 */

/** Normaliza a minusculas sin diacriticos para comparar de forma estable. */
export function normalizar(texto: string): string {
  return texto
    .toLowerCase()
    .normalize("NFD")
    .replace(/p{Diacritic}/gu, "");
}

/** Indice de textos descifrados: id de sobre hacia texto normalizado. */
export class IndiceBusqueda {
  private readonly textos = new Map<string, string>();

  indexar(id: string, texto: string): void {
    this.textos.set(id, normalizar(texto));
  }

  /** Ids cuyo texto indexado contiene el termino normalizado. */
  buscar(termino: string): string[] {
    const objetivo = normalizar(termino).trim();
    if (objetivo === "") {
      return [];
    }
    const hallados: string[] = [];
    for (const [id, texto] of this.textos) {
      if (texto.includes(objetivo)) {
        hallados.push(id);
      }
    }
    return hallados;
  }

  contiene(id: string): boolean {
    return this.textos.has(id);
  }

  tamano(): number {
    return this.textos.size;
  }

  limpiar(): void {
    this.textos.clear();
  }
}
