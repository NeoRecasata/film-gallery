export interface Photo {
	id: string;
	title: string | null;
	description: string | null;
	slug: string;
	film_stock: string | null;
	camera: string | null;
	lens: string | null;
	taken_at: string | null;
	published: boolean;
	width: number;
	height: number;
	file_size: number;
	blur_hash: string | null;
	sort_order: number;
	created_at: string;
	updated_at: string;
	urls: Record<string, string>;
}

export interface PhotosResponse {
	data: Photo[];
	next_cursor: string | null;
}

export interface Collection {
	id: string;
	title: string;
	slug: string;
	description: string | null;
	cover_photo: string | null;
	sort_order: number;
	created_at: string;
	updated_at: string;
	photos?: Photo[];
	photo_count?: number;
	cover_url?: string | null;
}

export interface SiteSettings {
	site_title: string;
	site_description: string;
	photographer_name: string;
	about_text: string;
	social_links: SocialLink[];
	accent_color: string;
}

export interface SocialLink {
	platform: string;
	url: string;
}

export interface User {
	id: string;
	email: string;
	created_at: string;
}

export interface ApiError {
	error: string;
}
