import fs from 'node:fs';
import fsp from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);

if (!globalThis.self) {
  globalThis.self = globalThis;
}

const { ResolverFactory, CachedInputFileSystem } = require('enhanced-resolve');
const esbuild = await import('esbuild-wasm/esm/browser.js');

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const rootDir = path.resolve(__dirname, '..');
const distDir = path.join(rootDir, 'dist');
const publicDir = path.join(rootDir, 'public');
const emptyModulePath = path.join(__dirname, 'empty-module.js');
const virtualRoot = '/@fs';

const fileSystem = new CachedInputFileSystem(fs, 4000);
const resolver = ResolverFactory.createResolver({
  fileSystem,
  extensions: ['.jsx', '.js', '.mjs', '.cjs', '.json', '.css'],
  extensionAlias: {
    '.js': ['.js', '.jsx', '.mjs'],
    '.mjs': ['.mjs', '.js'],
    '.cjs': ['.cjs', '.js'],
  },
  conditionNames: ['browser', 'import', 'require', 'default'],
  mainFields: ['browser', 'module', 'main'],
  aliasFields: ['browser'],
  exportsFields: ['exports'],
  mainFiles: ['index'],
  modules: ['node_modules'],
  symlinks: false,
});

const textLoaders = new Map([
  ['.css', 'css'],
  ['.js', 'js'],
  ['.jsx', 'jsx'],
  ['.mjs', 'js'],
  ['.cjs', 'js'],
  ['.json', 'json'],
]);

const binaryLoaders = new Map([
  ['.png', 'file'],
  ['.jpg', 'file'],
  ['.jpeg', 'file'],
  ['.gif', 'file'],
  ['.svg', 'file'],
  ['.webp', 'file'],
  ['.ico', 'file'],
  ['.woff', 'file'],
  ['.woff2', 'file'],
  ['.ttf', 'file'],
  ['.eot', 'file'],
  ['.otf', 'file'],
]);

function toVirtualPath(realPath) {
  const normalized = path.resolve(realPath).replace(/\\/g, '/');
  const driveMatch = normalized.match(/^([A-Za-z]):\/(.*)$/);
  if (driveMatch) {
    return `${virtualRoot}/${driveMatch[1].toLowerCase()}/${driveMatch[2]}`;
  }
  return `${virtualRoot}${normalized.startsWith('/') ? '' : '/'}${normalized}`;
}

function fromVirtualPath(virtualPath) {
  if (!virtualPath.startsWith(virtualRoot)) {
    return virtualPath;
  }

  // Keep leading slash for POSIX paths.
  // Example: /@fs/home/user/project -> /home/user/project
  const relativePath = virtualPath.slice(virtualRoot.length);

  // Windows drive form: /@fs/c/Users/... -> C:\Users\...
  const driveMatch = relativePath.match(/^\/([a-z])\/(.*)$/);
  if (driveMatch) {
    return `${driveMatch[1].toUpperCase()}:\\${driveMatch[2].replace(/\//g, '\\')}`;
  }

  return relativePath.replace(/\//g, path.sep);
}

function resolveRequest(context, request) {
  return new Promise((resolve, reject) => {
    resolver.resolve({}, context, request, {}, (error, result) => {
      if (error) {
        reject(error);
        return;
      }
      resolve(result);
    });
  });
}

function isExternalRequest(request) {
  return /^(https?:)?\/\//.test(request);
}

function normalizeOutputPath(filePath) {
  const normalized = filePath.replace(/\//g, path.sep);
  if (/^[\\/][A-Za-z]:[\\/]/.test(normalized)) {
    return normalized.slice(1);
  }
  return normalized;
}

const wasmBinary = await fsp.readFile(require.resolve('esbuild-wasm/esbuild.wasm'));
const wasmModule = await WebAssembly.compile(wasmBinary);
await esbuild.initialize({ wasmModule, worker: false });

await fsp.mkdir(__dirname, { recursive: true });
await fsp.writeFile(emptyModulePath, 'export default {}\n', 'utf8');

try {
  await fsp.rm(distDir, { recursive: true, force: true });
  await fsp.mkdir(distDir, { recursive: true });

  const result = await esbuild.build({
    entryPoints: [toVirtualPath(path.join(rootDir, 'src', 'main.jsx'))],
    bundle: true,
    format: 'esm',
    splitting: false,
    platform: 'browser',
    target: ['es2020'],
    outfile: path.join(distDir, 'assets', 'app.js'),
    assetNames: 'assets/[name]-[hash]',
    loader: Object.fromEntries([...textLoaders, ...binaryLoaders]),
    minify: true,
    legalComments: 'none',
    sourcemap: false,
    jsx: 'automatic',
    define: {
      'process.env.NODE_ENV': '"production"',
      global: 'globalThis',
    },
    plugins: [
      {
        name: 'formal-resolver',
        setup(build) {
          build.onResolve({ filter: /.*/ }, async (args) => {
            if (args.path.startsWith('data:') || isExternalRequest(args.path)) {
              return { path: args.path, external: true };
            }

            if (args.path === false) {
              return { path: toVirtualPath(emptyModulePath), namespace: 'local' };
            }

            if (args.path.startsWith(virtualRoot)) {
              return { path: args.path, namespace: 'local' };
            }

            if (path.isAbsolute(args.path)) {
              return { path: toVirtualPath(args.path), namespace: 'local' };
            }

            const context = args.resolveDir ? fromVirtualPath(args.resolveDir) : rootDir;
            try {
              const resolved = await resolveRequest(context, args.path);
              if (resolved === false) {
                return { path: toVirtualPath(emptyModulePath), namespace: 'local' };
              }
              return { path: toVirtualPath(resolved), namespace: 'local' };
            } catch (error) {
              throw new Error(`Failed to resolve ${args.path} from ${context}: ${error.message}`);
            }
          });

          build.onLoad({ filter: /.*/, namespace: 'local' }, async (args) => {
            const realPath = fromVirtualPath(args.path);
            const extension = path.extname(realPath).toLowerCase();
            const loader = textLoaders.get(extension) || binaryLoaders.get(extension) || 'file';

            if (binaryLoaders.has(extension)) {
              return {
                contents: await fsp.readFile(realPath),
                loader,
                resolveDir: toVirtualPath(path.dirname(realPath)),
              };
            }

            return {
              contents: await fsp.readFile(realPath, 'utf8'),
              loader,
              resolveDir: toVirtualPath(path.dirname(realPath)),
            };
          });
        },
      },
    ],
    write: false,
  });

  for (const outputFile of result.outputFiles) {
    const outputPath = normalizeOutputPath(outputFile.path);
    await fsp.mkdir(path.dirname(outputPath), { recursive: true });
    await fsp.writeFile(outputPath, outputFile.contents);
  }

  if (fs.existsSync(publicDir)) {
    await fsp.cp(publicDir, distDir, { recursive: true });
  }

  const cssFile = result.outputFiles.find((file) => normalizeOutputPath(file.path).endsWith('.css'));
  const html = [
    '<!doctype html>',
    '<html lang="en">',
    '  <head>',
    '    <meta charset="UTF-8" />',
    '    <meta name="viewport" content="width=device-width, initial-scale=1.0" />',
    '    <title>ApiHub</title>',
    cssFile ? `    <link rel="stylesheet" href="/${path.relative(distDir, normalizeOutputPath(cssFile.path)).replace(/\\/g, '/')}" />` : '',
    '  </head>',
    '  <body>',
    '    <div id="root"></div>',
    '    <script type="module" src="/assets/app.js"></script>',
    '  </body>',
    '</html>',
  ].filter(Boolean).join('\n') + '\n';
  await fsp.writeFile(path.join(distDir, 'index.html'), html, 'utf8');

  console.log('Formal production bundle written to web/dist');
} finally {
  await fsp.rm(emptyModulePath, { force: true });
  esbuild.stop();
}
