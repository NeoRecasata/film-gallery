import { redirect } from '@sveltejs/kit';
import { api, getAccessToken } from '$lib/api';
import type { LayoutLoad } from './$types';

export const ssr = false; // Admin is client-side only

export const load: LayoutLoad = async ({ url }) => {
	// Allow login and setup pages without auth
	if (url.pathname === '/admin/login' || url.pathname === '/admin/setup') {
		return {};
	}

	// Try to refresh if no access token
	if (!getAccessToken()) {
		const refreshed = await api.refresh();
		if (!refreshed) {
			throw redirect(302, '/admin/login');
		}
	}

	return {};
};
