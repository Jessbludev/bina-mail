<script setup lang="ts">
import { computed, ref } from "vue";
import { RouterLink } from "vue-router";
import BaseButton from "../components/ui/BaseButton.vue";
import BaseField from "../components/ui/BaseField.vue";
import BasePanel from "../components/ui/BasePanel.vue";
import { ApiError, lookupKey, storeEnvelope } from "../shared/api";
import { cryptoRequest } from "../shared/cryptoBridge";
import type { EnvelopeJson } from "../workers/crypto/envelope";
import { useSessionStore } from "../stores/session";
import {
  field,
  fieldInputClasses,
  layout,
  navLinkClasses,
  stack,
  statusClasses,
  typography,
} from "../styles/tokens";

const session = useSessionStore();

const destino = ref("");
const cuerpo = ref("");
const ocupado = ref(false);
const aviso = ref<{ tipo: "success" | "error"; texto: string } | null>(null);

const puedeEnviar = computed(
  () => destino.value.trim() !== "" && cuerpo.value !== "" && !ocupado.value,
);

async function enviar(): Promise<void> {
  if (!puedeEnviar.value || session.phase !== "unlocked" || session.sessionToken === "") {
    return;
  }
  ocupado.value = true;
  aviso.value = null;
  const huella = destino.value.trim();
  try {
    const clave = await lookupKey(session.sessionToken, huella);
    if (!clave.active) {
      aviso.value = { tipo: "error", texto: "La clave del destinatario no está activa." };
      return;
    }
    const sobre = await cryptoRequest<EnvelopeJson>({
      type: "seal",
      plaintext: cuerpo.value,
      recipientPublicKeyB64: clave.x25519_public_key,
      recipientFingerprint: huella,
      senderFingerprint: session.fingerprint,
    });
    await storeEnvelope(session.sessionToken, sobre);
    cuerpo.value = "";
    aviso.value = {
      tipo: "success",
      texto: "Sobre cifrado y almacenado; el destinatario lo verá en su bandeja.",
    };
  } catch (error) {
    aviso.value = { tipo: "error", texto: mensajeDe(error) };
  } finally {
    ocupado.value = false;
  }
}

function mensajeDe(error: unknown): string {
  if (error instanceof ApiError) {
    if (error.status === 404) {
      return "Ningún usuario tiene publicada esa huella.";
    }
    if (error.code === "fingerprint_invalido") {
      return "La huella debe ser de 64 caracteres hexadecimales.";
    }
    if (error.code === "limite_alcanzado" || error.status === 429) {
      return "Demasiados envíos; espere un momento.";
    }
    return "El servicio no está disponible ahora; vuelva a intentarlo.";
  }
  return "Ocurrió un error inesperado; vuelva a intentarlo.";
}
</script>

<template>
  <div :class="layout.container">
    <BasePanel title="Nuevo sobre">
      <p :class="typography.subheading">
        El mensaje se cifra en su navegador con la clave pública del destinatario;
        el servidor solo recibe ciphertext.
      </p>
      <form :class="stack" novalidate @submit.prevent="enviar">
        <BaseField
          id="redaccion-destino"
          v-model="destino"
          label="Huella del destinatario"
          type="text"
          autocomplete="off"
          required
        />
        <div>
          <label for="redaccion-cuerpo" :class="field.label">Mensaje</label>
          <textarea
            id="redaccion-cuerpo"
            v-model="cuerpo"
            :class="fieldInputClasses(false)"
            rows="8"
            required
          ></textarea>
        </div>
        <p v-if="aviso" :class="statusClasses(aviso.tipo)" role="status">{{ aviso.texto }}</p>
        <div :class="layout.row">
          <BaseButton type="submit" :loading="ocupado" :disabled="!puedeEnviar">
            Enviar
          </BaseButton>
          <RouterLink to="/bandeja" :class="navLinkClasses(false)">Volver a la bandeja</RouterLink>
        </div>
      </form>
    </BasePanel>
  </div>
</template>
