<script setup lang="ts">
import { computed } from "vue";
import { panel } from "../../styles/tokens";

const props = defineProps<{ title?: string }>();

const headingId = computed(() => (props.title ? "panel-" + props.title.toLowerCase().replace(/[^a-z0-9]+/g, "-") : undefined));
</script>

<template>
  <section :class="panel.box" :aria-labelledby="headingId">
    <h2 v-if="props.title" :id="headingId" :class="panel.title">{{ props.title }}</h2>
    <slot />
  </section>
</template>
