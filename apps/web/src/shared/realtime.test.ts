import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ClienteRealtime } from "./realtime";

vi.mock("./config", () => ({
  loadConfig: () => ({
    gatewayUrl: "http://localhost:8080",
    supabaseUrl: "http://localhost:54321",
    supabaseAnonKey: "anon",
  }),
}));

class WsStub {
  static instancias: WsStub[] = [];
  url: string;
  onopen: ((...args: unknown[]) => void) | null = null;
  onmessage: ((ev: { data: string }) => void) | null = null;
  onclose: ((ev: { code: number }) => void) | null = null;
  onerror: ((...args: unknown[]) => void) | null = null;
  enviados: string[] = [];

  constructor(url: string) {
    this.url = url;
    WsStub.instancias.push(this);
  }

  send(data: string): void {
    this.enviados.push(data);
  }

  close(..._args: unknown[]): void {
    this.onclose?.({ code: 1000 });
  }
}

beforeEach(() => {
  vi.useFakeTimers();
  vi.stubGlobal("WebSocket", WsStub);
  WsStub.instancias = [];
});

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

function nuevoCliente(alSenal: () => void, alEstado?: (estado: string) => void) {
  const cliente = new ClienteRealtime({ token: "tok-1", alSenal, alEstado });
  cliente.iniciar();
  return cliente;
}

describe("ClienteRealtime", () => {
  it("autentica con la sesion opaca como primer mensaje", () => {
    nuevoCliente(() => {});
    const ws = WsStub.instancias[0];
    expect(ws.url).toBe("ws://localhost:8080/api/v1/realtime");
    ws.onopen?.();
    expect(ws.enviados).toHaveLength(1);
    expect(JSON.parse(ws.enviados[0])).toEqual({ type: "auth", token: "tok-1" });
  });

  it("anuncia listo con ready y avisa al recibir la senal de sobre nuevo", () => {
    const senal = vi.fn();
    const estados: string[] = [];
    nuevoCliente(senal, (e) => estados.push(e));
    const ws = WsStub.instancias[0];
    ws.onopen?.();
    ws.onmessage?.({ data: JSON.stringify({ type: "ready" }) });
    ws.onmessage?.({ data: JSON.stringify({ type: "mailbox.envelope.stored" }) });
    ws.onmessage?.({ data: JSON.stringify({ type: "mailbox.envelope.stored" }) });
    expect(senal).toHaveBeenCalledTimes(2);
    expect(estados).toContain("lista");
  });

  it("reconecta con espera exponencial tras un cierre normal", () => {
    nuevoCliente(() => {});
    const primero = WsStub.instancias[0];
    primero.onclose?.({ code: 1006 });
    expect(WsStub.instancias).toHaveLength(1);
    vi.advanceTimersByTime(500);
    expect(WsStub.instancias).toHaveLength(2);
    const segundo = WsStub.instancias[1];
    segundo.onclose?.({ code: 1006 });
    vi.advanceTimersByTime(500);
    expect(WsStub.instancias).toHaveLength(2);
    vi.advanceTimersByTime(500);
    expect(WsStub.instancias).toHaveLength(3);
  });

  it("no reconecta cuando el servidor rechaza la sesion (4401)", () => {
    const estados: string[] = [];
    nuevoCliente(() => {}, (e) => estados.push(e));
    const ws = WsStub.instancias[0];
    ws.onclose?.({ code: 4401 });
    vi.advanceTimersByTime(60000);
    expect(WsStub.instancias).toHaveLength(1);
    expect(estados).toContain("detenida");
  });

  it("detener cierra y cancela la reconexion programada", () => {
    const cliente = nuevoCliente(() => {});
    const ws = WsStub.instancias[0];
    ws.onclose?.({ code: 1006 });
    cliente.detener();
    vi.advanceTimersByTime(60000);
    expect(WsStub.instancias).toHaveLength(1);
  });
});
