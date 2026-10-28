<script setup lang="ts">
import { computed } from "vue";
import { buttonClasses, spinner, type ButtonVariant } from "../../styles/tokens";

const props = withDefaults(
  defineProps<{
    variant?: ButtonVariant;
    type?: "button" | "submit";
    loading?: boolean;
    disabled?: boolean;
  }>(),
  { variant: "primary", type: "button", loading: false, disabled: false },
);

const classes = computed(() => buttonClasses(props.variant));
</script>

<template>
  <button
    :class="classes"
    :type="props.type"
    :disabled="props.disabled || props.loading"
    :aria-busy="props.loading"
  >
    <span v-if="props.loading" :class="spinner" aria-hidden="true"></span>
    <slot />
  </button>
</template>
