import { env } from '$env/dynamic/private';
import { createServerApi } from '$lib/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	const apiUrl = env.API_URL || 'http://localhost:8080';
	const serverApi = createServerApi(apiUrl);

	try {
		const rolls = await serverApi.getPublicRolls();
		// Fetch each roll's photos in parallel
		const rollsWithPhotos = await Promise.all(
			rolls.map(async (roll) => {
				try {
					const detail = await serverApi.getPublicRoll(roll.slug);
					return detail;
				} catch {
					return { ...roll, photos: [] };
				}
			})
		);
		return { rolls: rollsWithPhotos };
	} catch {
		return { rolls: [] };
	}
};
