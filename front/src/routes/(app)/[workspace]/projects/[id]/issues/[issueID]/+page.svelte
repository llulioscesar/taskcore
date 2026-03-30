<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import type { PageData } from './$types';
	import { issues as issuesApi, type UpdateIssueBody } from '$lib/api';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { data }: { data: PageData } = $props();

	const t = $derived.by(() => {
		i18n.locale;
		return {
			title: m.issue_detail_title(),
			description: m.issue_detail_description(),
			priority: m.issue_detail_priority(),
			assignee: m.issue_detail_assignee(),
			dueDate: m.issue_detail_due_date(),
			save: m.issue_detail_save(),
			saving: m.issue_detail_saving(),
			saved: m.issue_detail_saved(),
			unassigned: m.issue_detail_unassigned(),
			noDate: m.issue_detail_no_date()
		};
	});

	let title = $state('');
	let description = $state('');
	let priority = $state('medium');
	let assigneeId = $state('');
	let dueDate = $state('');
	let saving = $state(false);
	let saved = $state(false);
	let error = $state('');

	$effect(() => {
		title = data.issue.title;
		description = data.issue.description ?? '';
		priority = data.issue.priority;
		assigneeId = data.issue.assignee_id ?? '';
		dueDate = data.issue.due_date ? data.issue.due_date.slice(0, 10) : '';
	});

	async function handleSave() {
		error = '';
		saving = true;
		saved = false;
		try {
			const body: UpdateIssueBody = {
				title,
				description: description || undefined,
				priority,
				assignee_id: assigneeId || null,
				due_date: dueDate || null
			};
			const updated = await issuesApi.update(data.project.id, data.issue.id, body);
			data.issue = updated;
			saved = true;
			setTimeout(() => { saved = false; }, 2000);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to save';
		} finally {
			saving = false;
		}
	}
</script>

<div class="mx-auto max-w-2xl space-y-6">
	<Card.Root>
		<Card.Header>
			<div class="flex items-center gap-2 text-sm text-muted-foreground">
				<span>{data.project.key}-{data.issue.number}</span>
			</div>
		</Card.Header>
		<Card.Content class="space-y-4">
			<div class="space-y-1.5">
				<label for="issue-title" class="text-sm font-medium">{t.title}</label>
				<Input id="issue-title" bind:value={title} required />
			</div>

			<div class="space-y-1.5">
				<label for="issue-desc" class="text-sm font-medium">{t.description}</label>
				<textarea
					id="issue-desc"
					bind:value={description}
					rows="5"
					class="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm shadow-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
				></textarea>
			</div>

			<Separator />

			<div class="grid grid-cols-2 gap-4">
				<div class="space-y-1.5">
					<label for="issue-priority" class="text-sm font-medium">{t.priority}</label>
					<select
						id="issue-priority"
						bind:value={priority}
						class="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
					>
						<option value="low">Low</option>
						<option value="medium">Medium</option>
						<option value="high">High</option>
						<option value="critical">Critical</option>
					</select>
				</div>

				<div class="space-y-1.5">
					<label for="issue-assignee" class="text-sm font-medium">{t.assignee}</label>
					<select
						id="issue-assignee"
						bind:value={assigneeId}
						class="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
					>
						<option value="">{t.unassigned}</option>
						{#each data.members as member}
							<option value={member.user_id}>{member.user_id}</option>
						{/each}
					</select>
				</div>
			</div>

			<div class="space-y-1.5">
				<label for="issue-due" class="text-sm font-medium">{t.dueDate}</label>
				<Input id="issue-due" type="date" bind:value={dueDate} />
			</div>

			{#if error}
				<p class="text-sm text-destructive">{error}</p>
			{/if}

			<div class="flex items-center gap-3 pt-2">
				<Button onclick={handleSave} disabled={saving || !title.trim()}>
					{saving ? t.saving : t.save}
				</Button>
				{#if saved}
					<span class="text-sm text-green-600">{t.saved}</span>
				{/if}
			</div>
		</Card.Content>
	</Card.Root>
</div>
