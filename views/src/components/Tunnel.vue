<template>
  <div class="flex justify-between my-6">
    <div class="flex items-center">
      <status :online="tunnel.running"/>
      <span class="ml-4 text-gray-600 text-md">{{ tunnel.name }}</span>
    </div>
    <toggle :name="tunnel.name" :value="tunnel.running" @input="toggle"/>
  </div>
</template>
<script lang="ts">
import {defineComponent, PropType} from "vue";
import Toggle from "./Toggle.vue";
import Status from "./Status.vue";
import {CustomWindow, Tunnel} from "./types";

declare let window: CustomWindow

export default defineComponent({
  components: {Status, Toggle},
  props: {
    tunnel: {
      type: Object as PropType<Tunnel>,
      default: () => ({})
    },
  },
  emits: ['refresh'],
  setup({tunnel}, {emit}) {
    const toggle = () => {
      if (window.toggle) {
        window.toggle(tunnel.name, !tunnel.running)
        emit('refresh')
      }
    }

    return {toggle}
  },
});
</script>

