<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { useRouter } from "vue-router";
import BaseButton from "../components/ui/BaseButton.vue";
import BaseField from "../components/ui/BaseField.vue";
import BaseSkeleton from "../components/ui/BaseSkeleton.vue";
import EmptyState from "../components/ui/EmptyState.vue";
import { useSessionStore } from "../stores/session";
import { useMailboxStore } from "../stores/mailbox";
import {
  ClienteRealtime,
  urlRealtime,
  type EstadoRealtime,
} from "../shared/realtime";
import { layout, list, listRow, stack, statusClasses, typography } from "../styles/tokens";

const session = useSessionStore();
const mailbox = useMailboxStore();
const router = useRouter();

const estadoRealtime = ref<EstadoRealtime>("detenida");
const termino = ref("");
let realtime: ClienteRealtime | null = null;

const buscando = computed(() => termino.value.trim() !== "");
const itemsVisibles = computed(() =>
  buscando.value ? mailbox.buscar(termino.value) : mailbox.items,
);

onMounted(() => {
  void refrescar();
  if (session.phase === "unlocked" && session.sessionToken !== "") {
    realtime = new ClienteRealtime({
      token: session.sessionToken,
      alSenal: () => {
        void refrescar();
      },
      alEstado: (estado) => {
        estadoRealtime.value = estado;
      },
    });
    realtime.iniciar();
  }
});

onUnmounted(() => {
  realtime?.detener();
  realtime = null;
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

function conteo(n: number): string {
  return n === 1 ? "1 resultado" : n + " resultados";
}

function mensajeDeError(): string {
  if (mailbox.error.includes("limite_alcanzado")) {
    return "Demasiadas consultas; espere un momento e intente de nuevo.";
  }
  return "No se pudo cargar la bandeja; compruebe su conexión.";
}

const textoRealtime = computed(() => {
  if (estadoRealtime.value === "lista") {
    return "Conexión en tiempo real activa.";
  }
  if (estadoRealtime.value === "conectando" || estadoRealtime.value === "reintentando") {
    return "Reconectando el canal en tiempo real…";
  }
  return "Canal en tiempo real detenido; use Actualizar.";
});
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
    <p :class="typography.meta" aria-live="polite">{{ textoRealtime }}</p>
    <div :class="stack">
      <BaseField
        id="bandeja-busqueda"
        v-model="termino"
        label="Buscar"
        type="text"
        autocomplete="off"
        placeholder="Buscar en los sobres abiertos"
      />
      <p v-if="buscando" :class="typography.meta" role="status">
        {{ conteo(itemsVisibles.length) }} en los sobres abiertos de esta sesión.
      </p>
      <BaseSkeleton v-if="mailbox.status === 'cargando'" :lines="4" />
      <p v-else-if="mailbox.status === 'error'" :class="statusClasses('error')" role="alert">
        {{ mensajeDeError() }}
      </p>
      <EmptyState
        v-else-if="buscando && itemsVisibles.length === 0"
        title="Sin resultados"
        description="Ningún sobre abierto en esta sesión contiene ese término."
      />
      <EmptyState
        v-else-if="itemsVisibles.length === 0"
        title="Su bandeja está vacía"
        description="Cuando alguien le envíe un sobre cifrado aparecerá aquí. El servidor solo almacena ciphertext."
      />
      <ul v-else :class="list">
        <li v-for="item in itemsVisibles" :key="item.id">
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
