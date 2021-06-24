<template>
  <div class="mx-6 my-10">
    <tunnel v-for="tunnel in services" :key="tunnel" :tunnel="tunnel" @refresh="onRefresh"/>
  </div>
</template>
<script lang="ts">
import {defineComponent, onMounted, onUnmounted, ref} from "vue";
import Tunnel from "./Tunnel.vue";
import {CustomWindow} from "./types";

declare let window: CustomWindow;

export default defineComponent({
  components: {Tunnel},
  setup() {
    const services = ref<any>([]);
    let interval: number;
    onMounted(async () => {
      if (window.services) {
        services.value = await window.services();
        interval = setInterval(async () => {
          if (document.hasFocus()) {
            const v = await window.services();
            v.sort((a, b) => a.name.localeCompare(b.name))
            if (JSON.stringify(services.value) !== JSON.stringify(v)) {
              services.value = v
            }
          }

        }, 60000)
      }
    });
    onUnmounted(() => {
      clearInterval(interval)
    })
    const wait = (t: number) => new Promise((r) => setTimeout(r, t))
    const onRefresh = async () => {
      if (window.services) {
        await wait(2000)
        const v = await window.services();
        v.sort((a, b) => a.name.localeCompare(b.name))
        if (JSON.stringify(services.value) !== JSON.stringify(v)) {
          services.value = v
        }
      }
    }


    return {services, onRefresh};
  },
});
</script>

