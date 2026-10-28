import { createRouter, createWebHistory } from "vue-router";
import { useSessionStore } from "./stores/session";

/**
 * Rutas de la Fase 4 (ADR-009): la URL es la unica fuente de verdad de la
 * vista; ninguna ruta transporta token ni datos en la query.
 */
type FaseRuta = "anonimo" | "sesion" | "desbloqueado";

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", redirect: "/acceso" },
    {
      path: "/acceso",
      name: "acceso",
      component: () => import("./views/AccesoView.vue"),
      meta: { fase: "anonimo" },
    },
    {
      path: "/desbloqueo",
      name: "desbloqueo",
      component: () => import("./views/DesbloqueoView.vue"),
      meta: { fase: "sesion" },
    },
    {
      path: "/bandeja",
      name: "bandeja",
      component: () => import("./views/BandejaView.vue"),
      meta: { fase: "desbloqueado" },
    },
    {
      path: "/bandeja/:id",
      name: "lectura",
      component: () => import("./views/LecturaView.vue"),
      meta: { fase: "desbloqueado" },
    },
    {
      path: "/redaccion",
      name: "redaccion",
      component: () => import("./views/RedaccionView.vue"),
      meta: { fase: "desbloqueado" },
    },
  ],
});

/** Guard por fase: la sesion decide que ruta es valida; la URL nunca manda. */
router.beforeEach((to) => {
  const session = useSessionStore();
  const fase = (to.meta.fase ?? "anonimo") as FaseRuta;
  if (fase === "desbloqueado") {
    if (session.phase === "unlocked") {
      return true;
    }
    return session.phase === "signed-out" ? "/acceso" : "/desbloqueo";
  }
  if (fase === "sesion") {
    if (session.phase === "signed-out") {
      return "/acceso";
    }
    if (session.phase === "unlocked") {
      return "/bandeja";
    }
    return true;
  }
  // Ruta anonima: con sesion activa no hay nada que hacer aqui.
  if (session.phase === "unlocked") {
    return "/bandeja";
  }
  if (session.phase === "signed-in") {
    return "/desbloqueo";
  }
  return true;
});

export default router;
