<script setup lang="ts">
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import BaseButton from "../components/ui/BaseButton.vue";
import BaseField from "../components/ui/BaseField.vue";
import BasePanel from "../components/ui/BasePanel.vue";
import { ApiError } from "../shared/api";
import { useSessionStore } from "../stores/session";
import { layout, stack, statusClasses, typography } from "../styles/tokens";

const session = useSessionStore();
const router = useRouter();

const modo = ref<"login" | "registro">("login");
const email = ref("");
const password = ref("");
const ocupado = ref(false);
const aviso = ref<{ texto: string } | null>(null);

const titulo = computed(() => (modo.value === "login" ? "Acceder a Bina Mail" : "Crear cuenta en Bina Mail"));
const textoBoton = computed(() => (modo.value === "login" ? "Iniciar sesión" : "Crear cuenta"));
const textoAlterno = computed(() => (modo.value === "login" ? "No tengo cuenta" : "Ya tengo cuenta"));

function cambiarModo() {
  modo.value = modo.value === "login" ? "registro" : "login";
  aviso.value = null;
}

async function enviar() {
  if (ocupado.value) {
    return;
  }
  aviso.value = null;
  ocupado.value = true;
  try {
    await session.iniciarSesion(email.value, password.value, modo.value === "registro");
    await router.replace("/desbloqueo");
  } catch (error) {
    aviso.value = { texto: mensajeDe(error) };
  } finally {
    ocupado.value = false;
  }
}

function mensajeDe(error: unknown): string {
  if (error instanceof ApiError) {
    if (error.status === 401) {
      return "Credenciales incorrectas o cuenta inexistente.";
    }
    if (error.status === 429) {
      return "Demasiados intentos; espere un momento.";
    }
    return "El servicio no está disponible ahora; vuelva a intentarlo.";
  }
  return "Ocurrió un error inesperado; vuelva a intentarlo.";
}
</script>

<template>
  <div :class="layout.center">
    <BasePanel :title="titulo">
      <p :class="typography.subheading">
        Mensajería interna cifrada de extremo a extremo; el servidor nunca ve el contenido.
      </p>
      <form :class="stack" novalidate @submit.prevent="enviar">
        <BaseField
          id="acceso-email"
          v-model="email"
          label="Correo"
          type="email"
          autocomplete="email"
          required
        />
        <BaseField
          id="acceso-password"
          v-model="password"
          label="Contraseña"
          type="password"
          autocomplete="current-password"
          required
        />
        <p v-if="aviso" :class="statusClasses('error')" role="alert">{{ aviso.texto }}</p>
        <div :class="layout.row">
          <BaseButton type="submit" :loading="ocupado">{{ textoBoton }}</BaseButton>
          <BaseButton variant="ghost" @click="cambiarModo">{{ textoAlterno }}</BaseButton>
        </div>
      </form>
    </BasePanel>
  </div>
</template>
