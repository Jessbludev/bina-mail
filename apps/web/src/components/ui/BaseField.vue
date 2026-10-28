<script setup lang="ts">
import { computed } from "vue";
import { field, fieldInputClasses } from "../../styles/tokens";

const props = withDefaults(
  defineProps<{
    id: string;
    label: string;
    type?: "text" | "email" | "password";
    modelValue?: string;
    error?: string;
    autocomplete?: string;
    required?: boolean;
    placeholder?: string;
  }>(),
  { type: "text", modelValue: "", error: "", autocomplete: "", required: false, placeholder: "" },
);

const emit = defineEmits<{ "update:modelValue": [value: string] }>();

const inputClasses = computed(() => fieldInputClasses(props.error !== ""));
const errorId = computed(() => props.id + "-error");
</script>

<template>
  <div>
    <label :class="field.label" :for="props.id">{{ props.label }}</label>
    <input
      :id="props.id"
      :class="inputClasses"
      :type="props.type"
      :value="props.modelValue"
      :aria-invalid="props.error !== ''"
      :aria-describedby="props.error !== '' ? errorId : undefined"
      :autocomplete="props.autocomplete || undefined"
      :required="props.required"
      :placeholder="props.placeholder || undefined"
      @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
    <p v-if="props.error !== ''" :id="errorId" :class="field.error">{{ props.error }}</p>
  </div>
</template>
