<script lang="ts" module>
	import { cn, type WithElementRef } from "$lib/utils.js";
	import type { HTMLInputAttributes } from "svelte/elements";
	import { tv } from "tailwind-variants";

	export const inputVariants = tv({
		base: "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50",
		variants: {
			variant: {
				default: "border-input",
				error: "border-destructive focus-visible:ring-destructive",
			},
		},
		defaultVariants: {
			variant: "default",
		},
	});

	export type InputProps = WithElementRef<HTMLInputAttributes> & {
		variant?: "default" | "error";
	};
</script>

<script lang="ts">
	let {
		class: className,
		variant = "default",
		type = "text",
		ref = $bindable(null),
		value = $bindable(),
		...restProps
	}: InputProps = $props();
</script>

<input
	bind:this={ref}
	bind:value
	class={cn(inputVariants({ variant }), className)}
	{type}
	{...restProps}
/>