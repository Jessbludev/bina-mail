<script setup lang="ts">
import { computed, ref } from "vue";
import BaseButton from "../components/ui/BaseButton.vue";
import BaseField from "../components/ui/BaseField.vue";
import BasePanel from "../components/ui/BasePanel.vue";
import { ApiError } from "../shared/api";
import { useSessionStore } from "../stores/session";
import { layout, stack, statusClasses, typography } from "../styles/tokens";

const session = useSessionStore();

const creando = computed(() => session.vault === null);
const passphrase = ref("");
const confirmacion = ref("");
const ocupado = ref(false);
const aviso = ref<{ tipo: "success" | "error"; texto: string } | null>(null);

const titulo = computed(() =>
  creando.value ? "Crear su identidad cifrada" : "Desbloquear su identidad",
);
const descripcion = computed(() =>
  creando.value
    ? "La frase descifra su clave privada solo en este dispositivo; el servidor jamás la recibe."
    : "Introduzca su frase para usar la identidad de este dispositivo en esta sesión.",
);
const errorConfirmacion = computed(() =>
  creando.value &&
  passphrase.value !== "" &&
  confirmacion.value !== "" &&
  passphrase.value !== confirmacion.value
    ? "Las frases no coinciden"
    : "",
);

async function enviar() {
  if (ocupado.value || errorConfirmacion.value !== "") {
    return;
  }
  ocupado.value = true;
  aviso.value = null;
  try {
    if (creando.value) {
      await session.crearIdentidad(passphrase.value);
      aviso.value = {
        tipo: "success",
        texto: "Identidad creada y clave pública publicada en el directorio.",
      };
    } else {
      await session.desbloquear(passphrase.value);
      aviso.value = { tipo: "success", texto: "Identidad desbloqueada en este dispositivo." };
    }
    passphrase.value = "";
    confirmacion.value = "";
  } catch (error) {
    aviso.value = { tipo: "error", texto: mensajeDe(error) };
  } finally {
    ocupado.value = false;
  }
}

function mensajeDe(error: unknown): string {
  if (error instanceof Error) {
    if (error.message === "frase_corta") {
      return "La frase debe tener al menos 8 caracteres.";
    }
    if (error.message === "frase_invalida") {
      return "La frase no corresponde a la identidad guardada en este dispositivo.";
    }
    if (error.message === "fingerprint_no_coincide") {
      return "La identidad descifrada no coincide con este dispositivo.";
    }
  }
  if (error instanceof ApiError) {
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
      <p :class="typography.subheading">{{ descripcion }}</p>
      <form :class="stack" novalidate @submit.prevent="enviar">
        <BaseField
          id="desbloqueo-frase"
          v-model="passphrase"
          label="Frase de identidad"
          type="password"
          autocomplete="new-password"
          required
        />
        <BaseField
          v-if="creando"
          id="desbloqueo-confirmacion"
          v-model="confirmacion"
          label="Confirmar frase"
          type="password"
          autocomplete="new-password"
          required
          :error="errorConfirmacion"
        />
        <p v-if="aviso" :class="statusClasses(aviso.tipo)" role="status">{{ aviso.texto }}</p>
        <p v-if="session.fingerprint !== ''" :class="typography.subheading">
          Huella de su clave: {{ session.fingerprint.slice(0, 16) }}…
        </p>
        <div :class="layout.row">
          <BaseButton type="submit" :loading="ocupado">
            {{ creando ? "Crear identidad" : "Desbloquear" }}
          </BaseButton>
        </div>
      </form>
    </BasePanel>
  </div>
</template>
