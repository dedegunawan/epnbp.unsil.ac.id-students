import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react-swc";
import path from "path";

export default defineConfig(({ mode }) => {
  // Load env vars from .env file in current working directory
  // In Docker, this will be /app/.env
  const env = loadEnv(mode, process.cwd(), '');

  // Debug: log env vars in development
  if (mode === 'development') {
    console.log('Vite Config - Mode:', mode);
    console.log('Vite Config - VITE_API_URL:', env.VITE_API_URL);
    console.log('Vite Config - VITE_BASE_URL:', env.VITE_BASE_URL);
  }

  return {
    base: env.VITE_BASE_URL || '/',
    server: {
      host: "::",
      port: 3000,
      proxy: {
        '/api': {
          target: env.VITE_API_URL || 'http://localhost:8080',
          changeOrigin: true,
          rewrite: path => path.replace(/^\/api/, '/api'),
        },
      },
    },
    plugins: [react()],
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
      },
    },
    build: {
      outDir: 'dist',
      sourcemap: false,
      // Ensure assets are referenced correctly with base path
      assetsDir: 'assets',
      rollupOptions: {
        output: {
          // Ensure chunk names are consistent
          chunkFileNames: 'assets/js/[name]-[hash].js',
          entryFileNames: 'assets/js/[name]-[hash].js',
          assetFileNames: 'assets/[ext]/[name]-[hash].[ext]',
        },
      },
    },
  };
});



