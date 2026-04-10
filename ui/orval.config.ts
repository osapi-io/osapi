import { defineConfig } from 'orval'

export default defineConfig({
  osapi: {
    input: {
      target: './src/sdk/gen/api.yaml',
    },
    output: {
      target: './src/sdk/gen/client.ts',
      schemas: './src/sdk/gen/schemas',
      client: 'fetch',
      mode: 'tags-split',
      baseUrl: false,
      override: {
        mutator: {
          path: './src/sdk/fetch.ts',
          name: 'apiFetch',
        },
      },
    },
  },
})
