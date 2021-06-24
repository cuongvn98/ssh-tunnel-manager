<template>
  <label :for="checkboxId" class="inline-flex items-center cursor-pointer"
         :class="{'opacity-50': disabled}"
  >
    <span class="relative">
      <span class="block w-6 h-4 bg-gray-300 rounded-full shadow-inner"></span>
      <span :class="customClass">
        <input
            :id="checkboxId"
            type="checkbox"
            class="absolute opacity-0 w-0 h-0"
            v-model="value"
            :disabled="disabled"
        />
      </span>
    </span>
  </label>
</template>>
<script lang="ts">
import {computed, defineComponent, ref, watch} from "vue";

export default defineComponent({
  props: {
    value: Boolean,
    name: String,
  },
  computed: {
    checkboxId: function (): string {
      return `checkbox:${this.name}`;
    },
  },
  emits: ["input"],
  setup(props, {emit}) {
    const disabled = ref(false)
    const value = ref(props.value);

    const customClass = computed(() =>
        value.value
            ? `absolute block w-2 h-2 mt-1 ml-1 rounded-full shadow inset-y-0 left-0 focus-within:shadow-outline
        transition-transform duration-300 ease-in-out bg-blue-500 transform translate-x-full`
            : "absolute block w-2 h-2 mt-1 ml-1 bg-white rounded-full shadow inset-y-0 left-0 focus-within:shadow-outline transition-transform duration-300 ease-in-out"
    );

    watch(value, () => {
      disabled.value = true
      emit("input", value.value)
    });

    watch(() => props.value, () => {
      disabled.value = false
    })


    return {value, customClass, disabled};
  },
});
</script>

