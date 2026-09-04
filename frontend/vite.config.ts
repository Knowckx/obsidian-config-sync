import { svelte } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';
import { resolve } from 'node:path';
import wails from "@wailsio/runtime/plugins/vite";

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [
        svelte(), 
        tailwindcss(),
        wails("./bindings")
    ],
    resolve: {
        alias: [
            { find: '@bindings', replacement: resolve(process.cwd(), 'bindings') },
            { find: '@', replacement: resolve(process.cwd(), 'src') }
        ]
    },
    server: {
        host: "127.0.0.1",
        port: Number(process.env.WAILS_VITE_PORT) || 33005,
        strictPort: true,
    },
});
