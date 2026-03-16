import type { Photo, PhotosResponse, Collection, SiteSettings, ApiError } from './types';

let accessToken: string | null = null;

export function setAccessToken(token: string | null) {
	accessToken = token;
}

export function getAccessToken(): string | null {
	return accessToken;
}

class ApiClient {
	private baseUrl: string;

	constructor(baseUrl: string = '') {
		this.baseUrl = baseUrl;
	}

	private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
		const headers: Record<string, string> = {
			...(options.headers as Record<string, string> || {})
		};

		if (accessToken) {
			headers['Authorization'] = `Bearer ${accessToken}`;
		}

		// Don't set Content-Type for FormData (browser sets boundary)
		if (!(options.body instanceof FormData)) {
			headers['Content-Type'] = 'application/json';
		}

		const res = await fetch(`${this.baseUrl}${path}`, {
			...options,
			headers,
			credentials: 'include' // send cookies (refresh token)
		});

		if (res.status === 401 && accessToken) {
			// Try to refresh
			const refreshed = await this.refresh();
			if (refreshed) {
				headers['Authorization'] = `Bearer ${accessToken}`;
				const retry = await fetch(`${this.baseUrl}${path}`, {
					...options,
					headers,
					credentials: 'include'
				});
				if (!retry.ok) {
					const err: ApiError = await retry.json();
					throw new Error(err.error);
				}
				if (retry.status === 204) return undefined as T;
				return retry.json();
			}
			throw new Error('Session expired');
		}

		if (!res.ok) {
			const err: ApiError = await res.json();
			throw new Error(err.error);
		}

		if (res.status === 204) return undefined as T;
		return res.json();
	}

	async refresh(): Promise<boolean> {
		try {
			const res = await fetch(`${this.baseUrl}/api/auth/refresh`, {
				method: 'POST',
				credentials: 'include'
			});
			if (!res.ok) return false;
			const data = await res.json();
			accessToken = data.access_token;
			return true;
		} catch {
			return false;
		}
	}

	// Public endpoints
	async getPhotos(cursor?: string, limit = 20): Promise<PhotosResponse> {
		const params = new URLSearchParams({ limit: String(limit) });
		if (cursor) params.set('cursor', cursor);
		return this.request(`/api/photos?${params}`);
	}

	async getPhoto(slug: string): Promise<Photo> {
		return this.request(`/api/photos/${slug}`);
	}

	async getCollections(): Promise<Collection[]> {
		return this.request('/api/collections');
	}

	async getCollection(slug: string): Promise<Collection> {
		return this.request(`/api/collections/${slug}`);
	}

	async getSiteSettings(): Promise<SiteSettings> {
		return this.request('/api/site');
	}

	// Auth
	async setup(email: string, password: string): Promise<{ access_token: string; user_id: string }> {
		return this.request('/api/auth/setup', {
			method: 'POST',
			body: JSON.stringify({ email, password })
		});
	}

	async login(email: string, password: string): Promise<{ access_token: string }> {
		return this.request('/api/auth/login', {
			method: 'POST',
			body: JSON.stringify({ email, password })
		});
	}

	async logout(): Promise<void> {
		await this.request('/api/auth/logout', { method: 'POST' });
		accessToken = null;
	}

	async changePassword(currentPassword: string, newPassword: string): Promise<void> {
		await this.request('/api/auth/change-password', {
			method: 'POST',
			body: JSON.stringify({ current_password: currentPassword, new_password: newPassword })
		});
	}

	// Admin - Photos
	async uploadPhoto(file: File, metadata?: Record<string, string>): Promise<Photo> {
		const form = new FormData();
		form.append('photo', file);
		if (metadata) {
			for (const [key, value] of Object.entries(metadata)) {
				if (value) form.append(key, value);
			}
		}
		return this.request('/api/admin/photos', { method: 'POST', body: form });
	}

	async updatePhoto(id: string, data: Partial<Photo>): Promise<Photo> {
		return this.request(`/api/admin/photos/${id}`, {
			method: 'PATCH',
			body: JSON.stringify(data)
		});
	}

	async deletePhoto(id: string): Promise<void> {
		await this.request(`/api/admin/photos/${id}`, { method: 'DELETE' });
	}

	async reorderPhotos(orders: { id: string; sort_order: number }[]): Promise<void> {
		await this.request('/api/admin/photos/reorder', {
			method: 'POST',
			body: JSON.stringify({ orders })
		});
	}

	// Admin - Collections
	async createCollection(data: { title: string; description?: string }): Promise<Collection> {
		return this.request('/api/admin/collections', {
			method: 'POST',
			body: JSON.stringify(data)
		});
	}

	async updateCollection(id: string, data: Partial<Collection>): Promise<Collection> {
		return this.request(`/api/admin/collections/${id}`, {
			method: 'PATCH',
			body: JSON.stringify(data)
		});
	}

	async deleteCollection(id: string): Promise<void> {
		await this.request(`/api/admin/collections/${id}`, { method: 'DELETE' });
	}

	async setCollectionPhotos(id: string, photos: { photo_id: string; sort_order: number }[]): Promise<void> {
		await this.request(`/api/admin/collections/${id}/photos`, {
			method: 'PUT',
			body: JSON.stringify({ photos })
		});
	}

	// Admin - Settings
	async updateSiteSettings(settings: Partial<SiteSettings>): Promise<void> {
		await this.request('/api/admin/settings', {
			method: 'PATCH',
			body: JSON.stringify(settings)
		});
	}
}

export const api = new ApiClient();

// Server-side client with explicit base URL
export function createServerApi(baseUrl: string): ApiClient {
	return new ApiClient(baseUrl);
}
