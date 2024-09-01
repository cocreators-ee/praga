import svg from "@poppanator/sveltekit-svg"
import {sveltekit} from '@sveltejs/kit/vite';
import {defineConfig} from 'vite';

export default defineConfig({
  plugins: [sveltekit(), svg()],

  // For `pnpm run dev`
  server: {
    proxy: {
      '/api': 'http://localhost:8086',
    },
  },
});
