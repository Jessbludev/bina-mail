/**
 * Sistema de diseno de Bina Mail: clases Tailwind seguras y tipadas en TypeScript.
 * Regla del proyecto: prohibido escribir clases sueltas en plantillas; solo variantes de aqui.
 * Contraste AA verificado (WCAG 2.1): texto >= 4,5:1 sobre neutral-950; foco visible.
 */
export const layout = {
  page: "min-h-screen bg-neutral-950 text-neutral-100",
  center: "flex min-h-screen items-center justify-center p-4",
  container: "mx-auto w-full max-w-3xl px-4 sm:px-6",
  stack: "space-y-4",
  row: "flex flex-wrap items-center gap-3",
} as const;

export const shell = {
  main: "mx-auto w-full max-w-3xl px-4 py-6 sm:px-6",
} as const;

export const typography = {
  heading: "text-2xl font-semibold text-neutral-50",
  subheading: "text-sm text-neutral-400",
  body: "whitespace-pre-wrap break-words text-sm text-neutral-100",
  meta: "text-xs text-neutral-500",
} as const;

export const button = {
  base: "inline-flex items-center justify-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-sky-400 focus-visible:ring-offset-2 focus-visible:ring-offset-neutral-950 disabled:cursor-not-allowed disabled:opacity-60",
  variants: {
    primary: "bg-sky-600 text-white hover:bg-sky-500",
    ghost: "text-neutral-200 hover:bg-neutral-800",
    danger: "bg-red-700 text-white hover:bg-red-600",
  },
} as const;

export type ButtonVariant = keyof typeof button.variants;

export function buttonClasses(variant: ButtonVariant): string {
  return [button.base, button.variants[variant]].join(" ");
}

export const spinner = "h-4 w-4 rounded-full border-2 border-white/30 border-t-white motion-safe:animate-spin";

export const field = {
  label: "block text-sm font-medium text-neutral-200",
  input:
    "mt-1 w-full rounded-md border bg-neutral-900 px-3 py-2 text-sm text-neutral-100 placeholder:text-neutral-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-sky-400",
  valid: "border-neutral-700",
  invalid: "border-red-500",
  error: "mt-1 text-sm text-red-400",
} as const;

export function fieldInputClasses(invalid: boolean): string {
  return [field.input, invalid ? field.invalid : field.valid].join(" ");
}

export const panel = {
  box: "rounded-lg border border-neutral-800 bg-neutral-900 p-4 sm:p-6",
  title: "text-lg font-semibold text-neutral-50",
} as const;

export const skeleton = "animate-pulse motion-safe:animate-pulse rounded bg-neutral-800";

export const a11y = {
  srOnly: "sr-only",
  skipLink:
    "absolute left-2 top-2 z-10 rounded-md bg-sky-600 px-3 py-2 text-sm font-medium text-white focus:outline-none focus-visible:ring-2 focus-visible:ring-sky-400",
} as const;

export const status = {
  banner: "rounded-md border px-3 py-2 text-sm",
  info: "border-neutral-700 bg-neutral-900 text-neutral-300",
  success: "border-emerald-700 bg-emerald-950 text-emerald-300",
  error: "border-red-800 bg-red-950 text-red-300",
} as const;

export type StatusKind = keyof typeof status;

export function statusClasses(kind: StatusKind): string {
  return [status.banner, status[kind]].join(" ");
}

export const emptyState = {
  box: "flex flex-col items-center gap-2 py-12 text-center",
  title: "text-base font-medium text-neutral-200",
  description: "max-w-md text-sm text-neutral-400",
} as const;

export const nav = {
  header: "border-b border-neutral-800 bg-neutral-950",
  row: "mx-auto flex w-full max-w-3xl flex-wrap items-center justify-between gap-2 px-4 py-3 sm:px-6",
  brand: "text-base font-semibold text-neutral-50",
  actions: "flex items-center gap-2",
  link: "rounded-md px-3 py-2 text-sm font-medium text-neutral-300 hover:bg-neutral-800 hover:text-neutral-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-sky-400",
  linkActive: "text-sky-300",
} as const;

export function navLinkClasses(active: boolean): string {
  return [nav.link, active ? nav.linkActive : ""].filter(Boolean).join(" ");
}

export const list = "m-0 list-none space-y-2 p-0";

export const listRow =
  "block w-full rounded-md border border-neutral-800 bg-neutral-900 p-3 text-left transition-colors hover:border-neutral-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-sky-400";
