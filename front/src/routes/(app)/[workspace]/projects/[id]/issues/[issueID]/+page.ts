// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1

import { issues, projects, issueTypes, workspaces } from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const [issue, project] = await Promise.all([
		issues.get(params.id, params.issueID),
		projects.get(params.id)
	]);

	const [types, members] = await Promise.all([
		issueTypes.list(params.id),
		workspaces.members.list(project.workspace_id)
	]);

	return {
		issue,
		project,
		issueTypes: types ?? [],
		members: members ?? [],
		breadcrumb: [{ label: project.name }, { label: issue.title }]
	};
};
