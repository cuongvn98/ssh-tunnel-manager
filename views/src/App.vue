<template>
  <settings />
</template>

<script lang="ts">
import { defineComponent } from "vue";
import Settings from "./components/Settings.vue";

export default defineComponent({
  name: "App",
  components: { Settings },
});
</script>

<style>
</style>
