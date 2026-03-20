<script lang="ts">
  let count = $state(0)
</script>

<div class="flex items-center gap-4">
  <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700" onclick={() => count--}>-</button>
  <span class="text-xl font-bold">{count}</span>
  <button class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700" onclick={() => count++}>+</button>
</div>
