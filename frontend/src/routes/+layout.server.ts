import { env } from '$env/dynamic/private';
import { createServerApi } from '$lib/api';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async () => {
	const apiUrl = env.API_URL || 'http://localhost:8080';
	const serverApi = createServerApi(apiUrl);

	try {
		const settings = await serverApi.getSiteSettings();
		return { settings };
	} catch {
		return {
			settings: {
				site_title: 'Gallery',
				site_description: '',
				photographer_name: '',
				about_text: '',
				social_links: [],
				accent_color: '#f5f5f5'
			}
		};
	}
};
