import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const wsTarget = env.VITE_WS_TARGET || 'ws://localhost:1780';

  return {
    plugins: [react()],
    server: {
      host: '0.0.0.0',
      port: 5174,
      proxy: {
        '/ws': {
          target: wsTarget,
          ws: true,
          changeOrigin: true
        }
      }
    }
  };
});
