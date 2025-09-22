import { defineConfig, type Plugin } from 'vite'
import react from '@vitejs/plugin-react'

const inlineAssets = (): Plugin => ({
  name: 'inline-assets',
  enforce: 'post',
  apply: 'build',
  generateBundle(_options, bundle) {
    let htmlAssetKey: string | undefined
    const cssSnippets: string[] = []
    const jsSnippets: string[] = []

    for (const [fileName, chunk] of Object.entries(bundle)) {
      if (fileName.endsWith('.html') && chunk.type === 'asset') {
        htmlAssetKey = fileName
        continue
      }

      if (fileName.endsWith('.css') && chunk.type === 'asset') {
        cssSnippets.push(typeof chunk.source === 'string' ? chunk.source : chunk.source.toString())
        delete bundle[fileName]
        continue
      }

      if (fileName.endsWith('.js') && chunk.type === 'chunk') {
        jsSnippets.push(chunk.code)
        delete bundle[fileName]
      }
    }

    if (!htmlAssetKey) {
      return
    }

    const htmlAsset = bundle[htmlAssetKey]
    if (!htmlAsset || htmlAsset.type !== 'asset') {
      return
    }

    let html = typeof htmlAsset.source === 'string' ? htmlAsset.source : htmlAsset.source.toString()

    html = html.replace(/<link rel="modulepreload"[^>]*>\s*/g, '')

    if (cssSnippets.length > 0) {
      const styles = `<style>${cssSnippets.join('\n')}</style>`
      html = html.replace(/<link rel="stylesheet"[^>]*>\s*/g, '')
      html = html.replace('</head>', `${styles}</head>`)
    }

    if (jsSnippets.length > 0) {
      const scripts = `<script type="module">${jsSnippets.join('\n')}</script>`
      html = html.replace(/<script[^>]*type="module"[^>]*src="[^"]+"[^>]*><\/script>\s*/g, '')
      html = html.replace('</body>', `${scripts}</body>`)
    }

    htmlAsset.source = html
  }
})

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), inlineAssets()],
  server: {
    port: 3000,
    open: true,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080/',
        changeOrigin: true,
        // rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
  build: {
    cssCodeSplit: false,
    modulePreload: false,
    outDir: 'dist',
    emptyOutDir: true,
    assetsInlineLimit: 100_000_000,
    rollupOptions: {
      output: {
        manualChunks: undefined,
      },
    },
  },
})
