import { createRouter, createWebHistory } from "vue-router";
import { useSessionStore } from "./stores/session";

/**
 * Rutas de la Fase 4 (ADR-009): la URL es la unica fuente de verdad de la
 * vista; ninguna ruta transporta token ni datos en la query.
 * La bandeja y la lectura llegan con la tanda siguiente.
 */
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
  ],
});

router.beforeEach((to) => {
  const session = useSessionStore();
  const fase = (to.meta.fase ?? "anonimo") as "anonimo" | "sesion";
  if (fase === "anonimo" && session.phase !== "signed-out") {
    return "/desbloqueo";
  }
  if (fase === "sesion" && session.phase === "signed-out") {
    return "/acceso";
  }
  return true;
});

export default router;
