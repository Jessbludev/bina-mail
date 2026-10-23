/**
 * Clases Tailwind seguras y tipadas, definidas en TypeScript (criterio del proyecto).
 * Prohibido escribir clases sueltas en plantillas: solo variantes tipadas de aquí.
 */
export const button = {
  base: "inline-flex items-center justify-center rounded-md px-4 py-2 text-sm font-medium focus:outline-none focus-visible:ring-2 focus-visible:ring-sky-500",
  variants: {
    primary: "bg-sky-600 text-white hover:bg-sky-500",
    ghost: "text-neutral-300 hover:bg-neutral-800",
  },
} as const;

export type ButtonVariant = keyof typeof button.variants;

export function buttonClasses(variant: ButtonVariant): string {
  return [button.base, button.variants[variant]].join(" ");
}
