<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import BottomNav from '$lib/components/BottomNav.svelte';

	let { children } = $props();

	onMount(() => {
		if (!localStorage.getItem('access_token')) {
			goto('/login');
		}
	});
</script>

<main class="min-h-screen pb-20 pt-12">
	<div class="mx-auto max-w-md">
		{@render children()}
	</div>
</main>

<BottomNav />
