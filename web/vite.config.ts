import { defineConfig, type Plugin } from 'vite'
import react from '@vitejs/plugin-react'
import { minify as minifyHtml } from 'html-minifier-terser'

const inlineAssets = (): Plugin => ({
  name: 'inline-assets',
  enforce: 'post',
  apply: 'build',
  async generateBundle(_options, bundle) {
    let htmlAssetKey: string | undefined
    const cssSnippets: string[] = []
    const jsChunks: { fileName: string; code: string }[] = []

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
        jsChunks.push({ fileName, code: chunk.code })
      }
    }

    if (!htmlAssetKey) return
    const htmlAsset = bundle[htmlAssetKey]
    if (!htmlAsset || htmlAsset.type !== 'asset') return

    let html = typeof htmlAsset.source === 'string' ? htmlAsset.source : htmlAsset.source.toString()

    html = html.replace(/<link rel="modulepreload"[^>]*>\s*/g, '')

    if (cssSnippets.length > 0) {
      const styles = `<style>${cssSnippets.join('\n')}</style>`
      html = html.replace(/<link rel="stylesheet"[^>]*>\s*/g, '')
      html = html.replace(/<\/head>/, (m) => `${styles}${m}`)
    }

    if (jsChunks.length === 1) {
      const [onlyChunk] = jsChunks
      const scripts = `<script type="module">${onlyChunk.code}</script>`
      html = html.replace(/<script[^>]*type="module"[^>]*src="[^"]+"[^>]*><\/script>\s*/g, '')
      html = html.replace(/<\/body>/, (m) => `${scripts}${m}`)
      delete bundle[onlyChunk.fileName]
    } else if (jsChunks.length > 1) {
      this.warn(
        `inline-assets: skipped inlining JavaScript because ${jsChunks.length} chunks were produced; consider enabling inlineDynamicImports if you want a single inline bundle.`
      )
    }

    htmlAsset.source = await minifyHtml(html, {
      collapseWhitespace: true,
      removeComments: true,
      removeRedundantAttributes: true,
      removeEmptyAttributes: true,
      useShortDoctype: true,
      minifyJS: true,
      minifyCSS: true,
      keepClosingSlash: true,
    })
  }
})


export default defineConfig({
  plugins: [react(), inlineAssets()],

  // ✅ 把 esbuild 放在顶层，而不是 build 里
  esbuild: {
    drop: ['console', 'debugger'],
    legalComments: 'none',
  },

  server: {
    port: 3000,
    open: true,
    proxy: {
      '/api': {
        target: 'http://192.168.10.53:6211/',
        // target: 'http://127.0.0.1:8080/',
        changeOrigin: true,
      },
      '/files': {
        target: 'http://192.168.10.53:6211/',
        //  target: 'http://127.0.0.1:8080/',
        changeOrigin: true,
      },
    },
  },
  build: {
    cssCodeSplit: false,
    modulePreload: false,
    outDir: 'dist',
    emptyOutDir: true,
    assetsInlineLimit: 100_000_000,
    minify: 'esbuild',
    rollupOptions: {
      output: {
        inlineDynamicImports: true,
        manualChunks: undefined,
      },
    },
  },
})
