import react from '@vitejs/plugin-react';
import { fileURLToPath } from 'url';
import { defineConfig } from 'vite';

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [react()],
    server: {
        // @TODO: This is a temporary proxy while the new frontend is not integrated
        proxy: {
            '/api': {
                target: 'http://localhost:8001',
                changeOrigin: true,
                rewrite: (path) => {
                    return path.replace(/^\/api/, '');
                },
            },
        },
    },
    resolve: {
        alias: [
            {
                find: '@',
                replacement: fileURLToPath(new URL('./src', import.meta.url)),
            },
        ],
    },
});
