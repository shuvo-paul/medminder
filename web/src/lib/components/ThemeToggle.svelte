<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import Sun from '@lucide/svelte/icons/sun';
	import Moon from '@lucide/svelte/icons/moon';
	import Monitor from '@lucide/svelte/icons/monitor';

	type Theme = 'system' | 'light' | 'dark';

	let theme = $state<Theme>('system');

	function applyTheme(t: Theme) {
		const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
		const isDark = t === 'dark' || (t === 'system' && prefersDark);
		document.documentElement.classList.toggle('dark', isDark);
		localStorage.setItem('theme', t);
		theme = t;
	}

	onMount(() => {
		const stored = (localStorage.getItem('theme') as Theme) ?? 'system';
		theme = stored;

		const mq = window.matchMedia('(prefers-color-scheme: dark)');
		const handler = () => {
			if (theme === 'system') applyTheme('system');
		};
		mq.addEventListener('change', handler);
		return () => mq.removeEventListener('change', handler);
	});

	function cycle() {
		const next: Theme = theme === 'system' ? 'light' : theme === 'light' ? 'dark' : 'system';
		applyTheme(next);
	}
</script>

<Button variant="ghost" size="icon" onclick={cycle} aria-label="Toggle theme">
	{#if theme === 'light'}
		<Sun class="size-4" />
	{:else if theme === 'dark'}
		<Moon class="size-4" />
	{:else}
		<Monitor class="size-4" />
	{/if}
</Button>
