<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { page } from '$app/state';
	import { currentUser, logout } from '$lib/stores/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import LogOutIcon from '@lucide/svelte/icons/log-out';

	let { children } = $props();

	const isWorkspaceRoute = $derived(!!page.params.workspace);
</script>

{#if isWorkspaceRoute}
	{@render children()}
{:else}
	<div class="bg-background min-h-screen">
		<header class="border-b">
			<div class="mx-auto flex h-14 max-w-screen-xl items-center justify-between px-6">
				<a href="/" class="flex items-center gap-2 font-semibold">
					<div class="bg-primary text-primary-foreground flex size-6 items-center justify-center rounded-md text-xs">
						T
					</div>
					Tookly
				</a>
				{#if $currentUser}
					<div class="flex items-center gap-3">
						<span class="text-muted-foreground text-sm">{$currentUser.email}</span>
						<Button variant="ghost" size="sm" onclick={() => logout()}>
							<LogOutIcon class="size-4" />
						</Button>
					</div>
				{/if}
			</div>
		</header>
		{@render children()}
	</div>
{/if}
