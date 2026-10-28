<script setup lang="ts">
import { RouterLink, RouterView, useRoute, useRouter } from "vue-router";
import { useSessionStore } from "./stores/session";
import { a11y, buttonClasses, layout, nav, navLinkClasses, shell } from "./styles/tokens";

const session = useSessionStore();
const route = useRoute();
const router = useRouter();

async function bloquear() {
  session.lock();
  await router.replace("/desbloqueo");
}

async function cerrarSesion() {
  await session.cerrarSesion();
  await router.replace("/acceso");
}
</script>

<template>
  <div :class="layout.page">
    <a href="#contenido" :class="a11y.skipLink">Saltar al contenido</a>
    <header v-if="session.phase !== 'signed-out'" :class="nav.header">
      <div :class="nav.row">
        <div :class="nav.actions">
          <p :class="nav.brand">Bina Mail</p>
          <RouterLink
            v-if="session.phase === 'unlocked'"
            :class="navLinkClasses(route.path.startsWith('/bandeja'))"
            to="/bandeja"
          >
            Bandeja
          </RouterLink>
          <RouterLink
            v-if="session.phase === 'unlocked'"
            :class="navLinkClasses(route.path === '/redaccion')"
            to="/redaccion"
          >
            Redacción
          </RouterLink>
        </div>
        <div :class="nav.actions">
          <button
            v-if="session.phase === 'unlocked'"
            :class="buttonClasses('ghost')"
            type="button"
            @click="bloquear"
          >
            Bloquear
          </button>
          <button :class="buttonClasses('ghost')" type="button" @click="cerrarSesion">
            Cerrar sesión
          </button>
        </div>
      </div>
    </header>
    <main id="contenido" :class="shell.main">
      <RouterView />
    </main>
  </div>
</template>
