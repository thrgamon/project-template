import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		host: '0.0.0.0',
		proxy: {
			'/api': { target: process.env.API_URL || 'http://localhost:8080', changeOrigin: true }
		}
	}
});
