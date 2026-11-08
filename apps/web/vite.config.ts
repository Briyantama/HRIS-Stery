import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	// @stery/kit is linked from the sibling stery-kit checkout (file: dependency), outside this root.
	server: { fs: { allow: ['.', '../../../stery-kit/web'] } }
});
