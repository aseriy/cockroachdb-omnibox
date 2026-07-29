<script setup>
import { computed, onMounted, ref } from 'vue'
import LanePane from './LanePane.vue'

const cfg = ref(null)

const scheme = computed(() => {
  const schemes = cfg.value.schemes
  for (const name in schemes) {
    if (schemes[name].active) {
      return schemes[name]
    }
  }
  return Object.values(schemes)[0]
})

const lanes = computed(() => cfg.value.lanes.slice(0, 3))

onMounted(async () => {
  cfg.value = await (await fetch('/api/config')).json()
})
</script>

<template>
  <div
    v-if="cfg"
    class="lanes"
    :class="cfg.orientation"
    :style="{ background: scheme.background, color: scheme.foreground }"
  >
    <LanePane
      v-for="lane in lanes"
      :key="lane.name"
      :lane="lane.name"
      :scheme="scheme"
      :buffer-size="cfg.buffer_size"
    />
  </div>
</template>

<style>
html,
body,
#app {
  height: 100%;
  margin: 0;
}

.lanes {
  display: flex;
  height: 100%;
  font-family: monospace;
}

.lanes.columns {
  flex-direction: row;
}

.lanes.rows {
  flex-direction: column;
}

.lanes.columns > .pane + .pane {
  border-left: 1px solid color-mix(in srgb, currentColor 25%, transparent);
}

.lanes.rows > .pane + .pane {
  border-top: 1px solid color-mix(in srgb, currentColor 25%, transparent);
}
</style>
