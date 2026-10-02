import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit({ adapter: adapter({ precompress: true, strict: true }) })],
	server: {
		host: '0.0.0.0',
		proxy: {
			'/api': { target: process.env.API_URL || 'http://localhost:8080', changeOrigin: true }
		}
	}
});
