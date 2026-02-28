export default {
	api: {
		input: { target: './docs/swagger.json' },
		output: {
			target: 'src/lib/api/generated/client.ts',
			schemas: 'src/lib/api/generated/models',
			client: 'react-query',
			mode: 'tags-split',
			mock: {
				type: 'msw',
				delay: false,
			},
			override: {
				mutator: {
					path: 'src/lib/api/mutator/custom-fetch.ts',
					name: 'customFetch',
				},
			},
			biome: true,
		},
	},
};
