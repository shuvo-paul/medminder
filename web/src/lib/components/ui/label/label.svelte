<script lang="ts" module>
	import { cn } from "$lib/utils.js";
	import type { HTMLLabelAttributes } from "svelte/elements";

	export type LabelProps = HTMLLabelAttributes;
</script>

<script lang="ts">
	let { class: className, ...restProps }: LabelProps = $props();
</script>

<label class={cn("text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70", className)} {...restProps}></label>