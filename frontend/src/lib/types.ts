export interface Photo {
	id: string;
	title: string | null;
	description: string | null;
	slug: string;
	film_stock: string | null;
	camera: string | null;
	lens: string | null;
	taken_at: string | null;
	hidden: boolean;
	roll_id: string;
	location: string | null;
	roll_slug?: string;
	roll_title?: string;
	prev_slug?: string | null;
	next_slug?: string | null;
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

export interface Roll {
	id: string;
	title: string;
	slug: string;
	description: string | null;
	camera: string | null;
	film_stock: string | null;
	lens: string | null;
	location: string | null;
	shot_at: string | null;
	published: boolean;
	cover_photo_id: string | null;
	sort_order: number;
	created_at: string;
	updated_at: string;
	photos?: Photo[];
	photo_count?: number;
	cover_url?: string | null;
}

export interface AdminStats {
	roll_count: number;
	photo_count: number;
	collection_count: number;
	storage_bytes: number;
}

export interface UploadResult {
	uploaded: Photo[];
	failed: { filename: string; error: string }[];
}

export interface ApiError {
	error: string;
}
