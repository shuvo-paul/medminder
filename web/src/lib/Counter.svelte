<script lang="ts">
  let count = $state(0)
</script>

<div class="flex items-center gap-4">
  <button class="btn preset-filled" onclick={() => count--}>-</button>
  <span class="text-xl font-bold">{count}</span>
  <button class="btn preset-filled" onclick={() => count++}>+</button>
</div>
