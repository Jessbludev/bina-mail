<script setup lang="ts">
import { skeleton, a11y } from "../../styles/tokens";

defineProps<{ lines?: number }>();
const props = { lines: 3 };
void props;
</script>

<template>
  <div role="status">
    <span :class="a11y.srOnly">Cargando…</span>
    <div
      v-for="line in (typeof $attrs['data-lines'] === 'undefined' ? 3 : Number($attrs['data-lines']))"
      :key="line"
      :class="skeleton"
      :style="{ height: '0.75rem', marginBottom: '0.5rem' }"
      aria-hidden="true"
    ></div>
  </div>
</template>
