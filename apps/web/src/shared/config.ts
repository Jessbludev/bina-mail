/**
 * Configuracion de entorno de la aplicacion web.
 * Falla de forma explicita y temprana si falta alguna variable.
 */
export interface AppConfig {
  readonly gatewayUrl: string;
  readonly supabaseUrl: string;
  readonly supabaseAnonKey: string;
}

export function loadConfig(): AppConfig {
  const gatewayUrl = import.meta.env.VITE_GATEWAY_URL;
  const supabaseUrl = import.meta.env.VITE_SUPABASE_URL;
  const supabaseAnonKey = import.meta.env.VITE_SUPABASE_ANON_KEY;
  if (!gatewayUrl || !supabaseUrl || !supabaseAnonKey) {
    throw new Error(
      "config: faltan VITE_GATEWAY_URL, VITE_SUPABASE_URL o VITE_SUPABASE_ANON_KEY",
    );
  }
  return { gatewayUrl, supabaseUrl, supabaseAnonKey };
}
