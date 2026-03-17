import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error';

export interface Toast {
	id: number;
	message: string;
	type: ToastType;
}

let nextId = 0;

function createToastStore() {
	const { subscribe, update } = writable<Toast[]>([]);

	function add(message: string, type: ToastType = 'success', duration = 3000) {
		const id = nextId++;
		update((toasts) => [...toasts, { id, message, type }]);
		setTimeout(() => {
			update((toasts) => toasts.filter((t) => t.id !== id));
		}, duration);
	}

	function dismiss(id: number) {
		update((toasts) => toasts.filter((t) => t.id !== id));
	}

	return {
		subscribe,
		success: (message: string) => add(message, 'success'),
		error: (message: string) => add(message, 'error'),
		dismiss
	};
}

export const toasts = createToastStore();
