<script lang="ts">
	import { page } from '$app/stores';
	import { cn } from '$lib/utils';
	import LayoutDashboard from '@lucide/svelte/icons/layout-dashboard';
	import Pill from '@lucide/svelte/icons/pill';
	import CalendarClock from '@lucide/svelte/icons/calendar-clock';
	import UsersRound from '@lucide/svelte/icons/users-round';
	import UserRound from '@lucide/svelte/icons/user-round';

	const navItems = [
		{ href: '/', label: 'Dashboard', icon: LayoutDashboard, match: (p: string) => p === '/' },
		{ href: '/medications', label: 'Medications', icon: Pill, match: (p: string) => p.startsWith('/medications') },
		{ href: '/schedule', label: 'Schedule', icon: CalendarClock, match: (p: string) => p.startsWith('/schedule') },
		{ href: '/profiles', label: 'Profiles', icon: UsersRound, match: (p: string) => p.startsWith('/profiles') },
		{ href: '/account', label: 'Account', icon: UserRound, match: (p: string) => p.startsWith('/account') },
	];
</script>

<nav
	class="fixed bottom-0 left-0 right-0 z-50 border-t border-border bg-background/95 backdrop-blur-sm"
	style="padding-bottom: env(safe-area-inset-bottom, 0px);"
>
	<ul class="mx-auto flex h-16 max-w-md items-stretch">
		{#each navItems as item}
			{@const active = item.match($page.url.pathname)}
			<li class="flex-1">
				<a
					href={item.href}
					class={cn(
						'flex h-full flex-col items-center justify-center gap-0.5 transition-colors duration-150',
						active ? 'text-primary' : 'text-muted-foreground'
					)}
					aria-current={active ? 'page' : undefined}
				>
					<span
						class={cn(
							'flex items-center justify-center rounded-full p-1.5 transition-all duration-200',
							active && 'bg-primary/10'
						)}
					>
						<svelte:component
							this={item.icon}
							class={cn('size-5 transition-transform duration-200', active && 'scale-110')}
						/>
					</span>
					<span class="text-[10px] font-medium leading-none tracking-wide">{item.label}</span>
				</a>
			</li>
		{/each}
	</ul>
</nav>
