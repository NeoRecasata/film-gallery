import { error } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { createServerApi } from '$lib/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params }) => {
	const apiUrl = env.API_URL || 'http://localhost:8080';
	const serverApi = createServerApi(apiUrl);

	try {
		const roll = await serverApi.getPublicRoll(params.slug);
		return { roll };
	} catch {
		throw error(404, 'Roll not found');
	}
};
