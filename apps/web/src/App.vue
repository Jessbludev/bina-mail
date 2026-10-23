<script setup lang="ts">
import { useSessionStore } from "./stores/session";
import { buttonClasses } from "./styles/tokens";

const session = useSessionStore();

function toggleSession() {
  if (session.status === "unlocked") {
    session.lock();
  } else {
    session.unlock();
  }
}
</script>

<template>
  <main class="flex min-h-screen items-center justify-center bg-neutral-950">
    <div class="space-y-4 text-center">
      <h1 class="text-2xl font-semibold text-neutral-50">Bina Mail</h1>
      <p class="text-sm text-neutral-400">Estado de sesión: {{ session.status }}</p>
      <button :class="buttonClasses('primary')" type="button" @click="toggleSession">
        {{ session.status === "unlocked" ? "Bloquear" : "Desbloquear" }}
      </button>
    </div>
  </main>
</template>
