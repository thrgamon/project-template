export default {
	api: {
		input: { target: './docs/swagger.json' },
		output: {
			target: 'src/lib/api/generated/client.ts',
			schemas: 'src/lib/api/generated/models',
			client: 'fetch',
			mode: 'tags-split',
			mock: true
		}
	}
};
