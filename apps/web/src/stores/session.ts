import { defineStore } from "pinia";

export type SessionStatus = "locked" | "unlocked";

/** Estado de sesión del cliente. Nunca contiene claves ni material derivado. */
export const useSessionStore = defineStore("session", {
  state: (): { status: SessionStatus } => ({ status: "locked" }),
  actions: {
    unlock() {
      this.status = "unlocked";
    },
    lock() {
      this.status = "locked";
    },
  },
});
