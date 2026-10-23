import { beforeEach, describe, expect, it } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { useSessionStore } from "./session";

describe("session store", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("inicia bloqueada", () => {
    expect(useSessionStore().status).toBe("locked");
  });

  it("desbloquea y bloquea", () => {
    const session = useSessionStore();
    session.unlock();
    expect(session.status).toBe("unlocked");
    session.lock();
    expect(session.status).toBe("locked");
  });
});
