<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

const props = defineProps({
  lane: String,
  scheme: Object,
  bufferSize: Number,
})

const records = ref([])
const lineHeight = ref(0)
const viewportRows = ref(0)
const typeWidth = ref(0)

const view = ref(null)
const probe = ref(null)

let source = null
let observer = null
let seq = 0

const visible = computed(() =>
  records.value.slice(Math.max(0, records.value.length - viewportRows.value))
)

function color(type) {
  return props.scheme.messages[type] ?? props.scheme.default_color
}

function size() {
  viewportRows.value = Math.max(0, Math.floor(view.value.clientHeight / lineHeight.value))
}

onMounted(() => {
  lineHeight.value = probe.value.offsetHeight
  size()
  observer = new ResizeObserver(size)
  observer.observe(view.value)

  source = new EventSource(`/api/stream/${encodeURIComponent(props.lane)}`)
  source.onmessage = (event) => {
    const record = JSON.parse(event.data)
    record.seq = seq++
    if (record.type.length > typeWidth.value) {
      typeWidth.value = record.type.length
    }
    records.value.push(record)
    if (records.value.length > props.bufferSize) {
      const evicted = records.value.shift()
      if (evicted.type.length >= typeWidth.value) {
        typeWidth.value = records.value.reduce((w, r) => Math.max(w, r.type.length), 0)
      }
    }
  }
  // A dropped stream stays silently blank; recovery arrives with RESET.
  source.onerror = () => source.close()
})

onBeforeUnmount(() => {
  source.close()
  observer.disconnect()
})
</script>

<template>
  <section class="pane">
    <header>{{ lane }}</header>
    <div ref="view" class="view">
      <span ref="probe" class="probe">X</span>
      <div v-for="r in visible" :key="r.seq" class="line" :style="{ height: lineHeight + 'px' }"><span :style="{ color: color(r.type) }">{{ r.type.padEnd(typeWidth + 1) }}</span>{{ r.msg }}</div>
    </div>
  </section>
</template>

<style scoped>
.pane {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
  min-height: 0;
}

header {
  font-weight: bold;
  padding: 4px 8px;
}

.view {
  flex: 1;
  overflow: hidden;
  padding: 0 8px;
  position: relative;
}

.probe {
  position: absolute;
  visibility: hidden;
  white-space: pre;
}

.line {
  white-space: pre;
  overflow: hidden;
}
</style>
