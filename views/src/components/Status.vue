<template>
  <div
    class="w-2 h-2 rounded-md"
    :class="{ 'bg-green-500': online, 'bg-gray-500': !online }"
  ></div>
</template>>
<script lang="ts">
import { defineComponent } from "vue";

export default defineComponent({
  props: {
    online: Boolean,
  },
  setup() {},
});
</script>

