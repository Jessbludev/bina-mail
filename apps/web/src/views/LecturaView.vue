<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import BasePanel from "../components/ui/BasePanel.vue";
import BaseSkeleton from "../components/ui/BaseSkeleton.vue";
import { useSessionStore } from "../stores/session";
import { useMailboxStore } from "../stores/mailbox";
import { layout, navLinkClasses, stack, statusClasses, typography } from "../styles/tokens";

const route = useRoute();
const session = useSessionStore();
const mailbox = useMailboxStore();

const id = computed(() => String(route.params.id ?? ""));
const sobre = computed(() => mailbox.porId(id.value));

const texto = ref("");
const cargando = ref(false);
const error = ref("");

onMounted(() => {
  void preparar();
});

async function preparar(): Promise<void> {
  if (session.phase !== "unlocked" || session.encryptionPrivateKeyB64 === "") {
    error.value = "sesion_bloqueada";
    return;
  }
  if (mailbox.items.length === 0) {
    cargando.value = true;
    await mailbox.refrescar(session.sessionToken, session.fingerprint);
    cargando.value = false;
    if (mailbox.status === "error") {
      error.value = "bandeja_error";
      return;
    }
  }
  await abrir();
}

async function abrir(): Promise<void> {
  if (sobre.value === null) {
    error.value = "sobre_inexistente";
    return;
  }
  cargando.value = true;
  error.value = "";
  try {
    texto.value = await mailbox.abrir(sobre.value.id, session.encryptionPrivateKeyB64);
  } catch (e) {
    error.value = e instanceof Error ? e.message : "error_inesperado";
  } finally {
    cargando.value = false;
  }
}

function remitenteDe(fp: string): string {
  return fp === "" ? "Remitente anónimo" : fp.slice(0, 16) + "…";
}

function fechaDe(created: string): string {
  const ms = Date.parse(created);
  return Number.isNaN(ms) ? "" : new Date(ms).toLocaleString();
}

function mensajeDeError(): string {
  if (error.value === "sesion_bloqueada") {
    return "Su identidad está bloqueada; desbloquéela para leer el sobre.";
  }
  if (error.value === "bandeja_error") {
    return "No se pudo cargar la bandeja; compruebe su conexión.";
  }
  if (error.value === "sobre_inexistente") {
    return "El sobre solicitado no está en su bandeja.";
  }
  if (error.value.includes("fingerprint no coincide")) {
    return "El sobre no pasó la verificación de integridad.";
  }
  if (error.value.includes("ninguna envoltura")) {
    return "Este dispositivo no puede descifrar el sobre.";
  }
  return "No se pudo descifrar el sobre.";
}
</script>

<template>
  <div :class="layout.container">
    <BasePanel title="Sobre cifrado">
      <div v-if="sobre !== null" :class="stack">
        <p :class="typography.meta">
          {{ remitenteDe(sobre.sender_fingerprint) }} · {{ fechaDe(sobre.created_at) }}
        </p>
        <BaseSkeleton v-if="cargando" :lines="5" />
        <p v-else-if="error !== ''" :class="statusClasses('error')" role="alert">
          {{ mensajeDeError() }}
        </p>
        <p v-else :class="typography.body">{{ texto }}</p>
      </div>
      <div v-else :class="stack">
        <BaseSkeleton v-if="cargando" :lines="3" />
        <p v-else :class="statusClasses('error')" role="alert">{{ mensajeDeError() }}</p>
      </div>
      <div :class="layout.row">
        <RouterLink to="/bandeja" :class="navLinkClasses(false)">Volver a la bandeja</RouterLink>
      </div>
    </BasePanel>
  </div>
</template>
