<script lang="ts">
	import { cn } from "$lib/utils.js";

	let {
		class: className,
		open,
		onclose,
		children,
	}: {
		open: boolean;
		onclose?: () => void;
		class?: string;
		children?: import("svelte").Snippet;
	} = $props();

	let ref = $state<HTMLDivElement | null>(null);

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) onclose?.();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onclose?.();
	}
</script>

{#if open}
	<div
		class="fixed inset-0 z-50 bg-black/50 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0"
		data-state={open ? 'open' : 'closed'}
		onclick={handleBackdropClick}
		onkeydown={handleKeydown}
		role="presentation"
	>
		<div
			bind:this={ref}
			role="dialog"
			aria-modal="true"
			data-state={open ? 'open' : 'closed'}
			class={cn(
				"fixed top-[50%] left-[50%] z-50 grid w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] gap-4 rounded-lg border bg-background p-6 shadow-lg sm:max-w-lg",
				"data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 duration-200",
			className
		)}>
		{@render children?.()}
		</div>
	</div>
{/if}
