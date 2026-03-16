import { error } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { createServerApi } from '$lib/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params }) => {
	const apiUrl = env.API_URL || 'http://localhost:8080';
	const serverApi = createServerApi(apiUrl);

	try {
		const collection = await serverApi.getCollection(params.slug);
		return { collection };
	} catch {
		throw error(404, 'Collection not found');
	}
};
