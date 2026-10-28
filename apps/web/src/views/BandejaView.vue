<script setup lang="ts">
import { onMounted } from "vue";
import { useRouter } from "vue-router";
import BaseButton from "../components/ui/BaseButton.vue";
import BaseSkeleton from "../components/ui/BaseSkeleton.vue";
import EmptyState from "../components/ui/EmptyState.vue";
import { useSessionStore } from "../stores/session";
import { useMailboxStore } from "../stores/mailbox";
import { layout, list, listRow, stack, statusClasses, typography } from "../styles/tokens";

const session = useSessionStore();
const mailbox = useMailboxStore();
const router = useRouter();

onMounted(() => {
  void refrescar();
});

async function refrescar(): Promise<void> {
  if (session.phase !== "unlocked" || session.sessionToken === "") {
    return;
  }
  await mailbox.refrescar(session.sessionToken, session.fingerprint);
}

async function abrir(id: string): Promise<void> {
  await router.push("/bandeja/" + encodeURIComponent(id));
}

function remitente(fp: string): string {
  return fp === "" ? "Remitente anónimo" : fp.slice(0, 16) + "…";
}

function fechaDe(created: string): string {
  const ms = Date.parse(created);
  return Number.isNaN(ms) ? "" : new Date(ms).toLocaleString();
}

function tamanoDe(bytes: number): string {
  return (bytes / 1024).toFixed(1) + " KiB";
}

function mensajeDeError(): string {
  if (mailbox.error.includes("limite_alcanzado")) {
    return "Demasiadas consultas; espere un momento e intente de nuevo.";
  }
  return "No se pudo cargar la bandeja; compruebe su conexión.";
}
</script>

<template>
  <div :class="layout.container">
    <div :class="layout.row">
      <h1 :class="typography.heading">Bandeja de entrada</h1>
      <BaseButton
        variant="ghost"
        :disabled="mailbox.status === 'cargando'"
        @click="refrescar"
      >
        Actualizar
      </BaseButton>
    </div>
    <div :class="stack">
      <BaseSkeleton v-if="mailbox.status === 'cargando'" :lines="4" />
      <p v-else-if="mailbox.status === 'error'" :class="statusClasses('error')" role="alert">
        {{ mensajeDeError() }}
      </p>
      <EmptyState
        v-else-if="mailbox.items.length === 0"
        title="Su bandeja está vacía"
        description="Cuando alguien le envíe un sobre cifrado aparecerá aquí. El servidor solo almacena ciphertext."
      />
      <ul v-else :class="list">
        <li v-for="item in mailbox.items" :key="item.id">
          <button :class="listRow" type="button" @click="abrir(item.id)">
            <span :class="typography.subheading">{{ remitente(item.sender_fingerprint) }}</span>
            <span :class="typography.meta">
              {{ fechaDe(item.created_at) }} · {{ tamanoDe(item.size_bytes) }}
            </span>
          </button>
        </li>
      </ul>
    </div>
  </div>
</template>
