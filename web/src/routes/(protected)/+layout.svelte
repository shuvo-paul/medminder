<script lang="ts">
	import BottomNav from '$lib/components/BottomNav.svelte';

	let { children } = $props();
</script>

<main class="min-h-screen pb-20 pt-12">
	<div class="mx-auto max-w-md">
		{@render children()}
	</div>
</main>

<BottomNav />
