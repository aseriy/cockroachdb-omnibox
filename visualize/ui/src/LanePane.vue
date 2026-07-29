<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

const props = defineProps({
  lane: String,
  scheme: Object,
  bufferSize: Number,
})

const records = ref([])
const lineHeight = ref(0)
const charWidth = ref(0)
const viewportRows = ref(0)
const cols = ref(1)
const typeWidth = ref(0)
const scrollRow = ref(0)
const follow = ref(true)

const view = ref(null)
const probe = ref(null)

let source = null
let observer = null
let seq = 0
let wheelDebt = 0

function rowsOf(r) {
  const width = cols.value
  const ind = typeWidth.value + 1
  const len = ind + r.msg.length
  if (len <= width) {
    return 1
  }
  return 1 + Math.ceil((len - width) / Math.max(1, width - ind))
}

const totalRows = computed(() => records.value.reduce((t, r) => t + rowsOf(r), 0))

const maxScroll = computed(() => Math.max(0, totalRows.value - viewportRows.value))

const paint = computed(() => {
  const rs = records.value
  let skip = scrollRow.value
  let i = 0
  while (i < rs.length && skip >= rowsOf(rs[i])) {
    skip -= rowsOf(rs[i])
    i++
  }
  const items = []
  let covered = -skip
  while (i < rs.length && covered < viewportRows.value + 1) {
    items.push(rs[i])
    covered += rowsOf(rs[i])
    i++
  }
  return { items, offset: skip }
})

const indent = computed(() => ({
  paddingLeft: `${typeWidth.value + 1}ch`,
  textIndent: `-${typeWidth.value + 1}ch`,
}))

function color(type) {
  return props.scheme.messages[type] ?? props.scheme.default_color
}

function setScroll(row) {
  scrollRow.value = Math.min(Math.max(0, row), maxScroll.value)
  follow.value = scrollRow.value >= maxScroll.value
}

function onWheel(event) {
  wheelDebt += event.deltaY
  const lines = Math.trunc(wheelDebt / lineHeight.value)
  if (!lines) {
    return
  }
  wheelDebt -= lines * lineHeight.value
  setScroll(scrollRow.value + lines)
}

function size() {
  viewportRows.value = Math.max(0, Math.floor(view.value.clientHeight / lineHeight.value))
  cols.value = Math.max(1, Math.floor((view.value.clientWidth - 16) / charWidth.value))
  if (follow.value) {
    scrollRow.value = maxScroll.value
  }
}

onMounted(() => {
  lineHeight.value = probe.value.offsetHeight
  charWidth.value = probe.value.offsetWidth
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
    let evicted = null
    if (records.value.length > props.bufferSize) {
      evicted = records.value.shift()
      if (evicted.type.length >= typeWidth.value) {
        typeWidth.value = records.value.reduce((w, r) => Math.max(w, r.type.length), 0)
      }
    }
    if (follow.value) {
      scrollRow.value = maxScroll.value
    } else if (evicted) {
      setScroll(scrollRow.value - rowsOf(evicted))
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
    <div ref="view" class="view" @wheel.prevent="onWheel">
      <span ref="probe" class="probe">X</span>
      <div class="content" :style="{ marginTop: `${-paint.offset * lineHeight}px` }">
        <div v-for="r in paint.items" :key="r.seq" class="line" :style="indent"><span :style="{ color: color(r.type) }">{{ r.type.padEnd(typeWidth + 1) }}</span>{{ r.msg }}</div>
      </div>
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
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
