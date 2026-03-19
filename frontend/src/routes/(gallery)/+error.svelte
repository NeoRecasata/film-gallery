<script lang="ts">
	import { page } from '$app/state';
</script>

<div class="max-w-7xl mx-auto px-4 py-32 text-center">
	<h1 class="text-6xl font-medium mb-4">{page.status}</h1>
	<p class="text-text-muted mb-8">{page.error?.message || 'Page not found'}</p>
	<a href="/" class="text-sm text-text-muted hover:text-text transition-colors">&larr; Back to gallery</a>
</div>
