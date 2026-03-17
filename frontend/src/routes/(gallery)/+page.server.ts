import { env } from '$env/dynamic/private';
import { createServerApi } from '$lib/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	const apiUrl = env.API_URL || 'http://localhost:8080';
	const serverApi = createServerApi(apiUrl);

	try {
		const photosResponse = await serverApi.getPhotos(undefined, 20, true);
		return { photosResponse };
	} catch {
		return { photosResponse: { data: [], next_cursor: null } };
	}
};
