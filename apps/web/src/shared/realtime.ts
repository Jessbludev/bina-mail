/**
 * Cliente del canal realtime (ADR-009): un unico WebSocket al gateway con la
 * sesion opaca como primer mensaje; el servidor nunca envia contenido.
 */
import { loadConfig } from "./config";

export type EstadoRealtime = "detenida" | "conectando" | "lista" | "reintentando";

export interface OpcionesRealtime {
  readonly token: string;
  readonly alSenal: () => void;
  readonly alEstado?: (estado: EstadoRealtime) => void;
}

const REINTENTO_BASE_MS = 500;
const REINTENTO_MAX_MS = 15000;

export class ClienteRealtime {
  private ws: WebSocket | null = null;
  private reintentos = 0;
  private detenido = false;
  private temporizador: ReturnType<typeof setTimeout> | null = null;

  constructor(private readonly opciones: OpcionesRealtime) {}

  iniciar(): void {
    this.detenido = false;
    this.conectar();
  }

  detener(): void {
    this.detenido = true;
    if (this.temporizador !== null) {
      clearTimeout(this.temporizador);
      this.temporizador = null;
    }
    if (this.ws !== null) {
      const ws = this.ws;
      this.ws = null;
      ws.close(1000, "cliente_detenido");
    }
    this.anunciar("detenida");
  }

  private conectar(): void {
    if (this.detenido) {
      return;
    }
    this.anunciar("conectando");
    const ws = new WebSocket(urlRealtime());
    this.ws = ws;
    ws.onopen = () => {
      // Primer mensaje: autenticacion con la sesion opaca; nunca en la URL.
      ws.send(JSON.stringify({ type: "auth", token: this.opciones.token }));
    };
    ws.onmessage = (ev: MessageEvent<string>) => {
      let tipo = "";
      try {
        const dato = JSON.parse(ev.data) as { type?: unknown };
        if (typeof dato.type === "string") {
          tipo = dato.type;
        }
      } catch {
        // Mensaje no JSON: se ignora; el canal solo transporta avisos.
        return;
      }
      if (tipo === "mailbox.envelope.stored") {
        this.opciones.alSenal();
      } else if (tipo === "ready") {
        this.reintentos = 0;
        this.anunciar("lista");
      }
    };
    ws.onclose = (ev: CloseEvent) => {
      this.ws = null;
      if (this.detenido) {
        return;
      }
      if (ev.code === 4401) {
        // Sesion rechazada: reintentar seria inutil hasta nueva sesion.
        this.anunciar("detenida");
        return;
      }
      this.reintentos += 1;
      this.anunciar("reintentando");
      const espera = Math.min(
        REINTENTO_BASE_MS * 2 ** (this.reintentos - 1),
        REINTENTO_MAX_MS,
      );
      this.temporizador = setTimeout(() => {
        this.conectar();
      }, espera);
    };
    ws.onerror = () => {
      ws.close();
    };
  }

  private anunciar(estado: EstadoRealtime): void {
    this.opciones.alEstado?.(estado);
  }
}

export function urlRealtime(): string {
  const base = loadConfig().gatewayUrl;
  return base.replace(/^http/, "ws") + "/api/v1/realtime";
}
