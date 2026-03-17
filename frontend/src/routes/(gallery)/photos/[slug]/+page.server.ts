import { error } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { createServerApi } from '$lib/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params }) => {
	const apiUrl = env.API_URL || 'http://localhost:8080';
	const serverApi = createServerApi(apiUrl);

	try {
		const photo = await serverApi.getPhoto(params.slug);
		return { photo };
	} catch {
		throw error(404, 'Photo not found');
	}
};
