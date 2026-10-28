<script setup lang="ts">
import { emptyState } from "../../styles/tokens";

defineProps<{ title: string; description?: string }>();
</script>

<template>
  <div :class="emptyState.box">
    <p :class="emptyState.title">{{ title }}</p>
    <p v-if="description" :class="emptyState.description">{{ description }}</p>
    <slot />
  </div>
</template>
