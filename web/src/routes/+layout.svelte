<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';

	let { children } = $props();
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<header class="fixed top-0 right-0 left-0 z-50 border-b border-border bg-background/95 backdrop-blur-sm">
	<div class="mx-auto flex h-12 max-w-md items-center justify-end px-4">
		<ThemeToggle />
	</div>
</header>

<main>
	<div class="mx-auto max-w-md">
		{@render children()}
	</div>
</main>
