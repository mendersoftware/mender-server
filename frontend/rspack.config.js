import { rspack } from '@rspack/core';
import ReactRefreshPlugin from '@rspack/plugin-react-refresh';
import { sentryWebpackPlugin } from '@sentry/webpack-plugin';
import autoprefixer from 'autoprefixer';
import { CleanWebpackPlugin } from 'clean-webpack-plugin';
import ESLintPlugin from 'eslint-rspack-plugin';
import { createRequire } from 'module';
import path from 'path';
import { fileURLToPath } from 'url';
import LicensePlugin from 'webpack-license-plugin';

const require = createRequire(import.meta.url);
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const publicPath = '/ui/';

export default (env, argv) => {
  const plugins =
    argv.mode === 'production'
      ? [
          new LicensePlugin({
            outputFilename: 'licenses.json',
            excludedPackageTest: packageName => packageName.startsWith('@northern.tech'),
            replenishDefaultLicenseTexts: true
          })
        ]
      : [
          new ESLintPlugin({
            configType: 'flat',
            extensions: ['js', 'ts', 'tsx'],
            failOnWarning: false,
            failOnError: false,
            emitWarning: true,
            emitError: true
          }),
          new ReactRefreshPlugin()
        ];
  const { GIT_COMMIT_SHA, SENTRY_AUTH_TOKEN, SENTRY_ORG, SENTRY_URL } = process.env;
  if (SENTRY_URL && SENTRY_AUTH_TOKEN && argv.mode === 'production') {
    plugins.push(
      sentryWebpackPlugin({
        authToken: SENTRY_AUTH_TOKEN,
        org: SENTRY_ORG,
        project: 'mender-frontend',
        release: { name: `mender-frontend@${GIT_COMMIT_SHA}` },
        sourcemaps: { ignore: ['./dist/vs'], assets: './dist' },
        telemetry: false,
        url: SENTRY_URL
      })
    );
  }
  return {
    devtool: 'source-map',
    node: {
      global: true
    },
    devServer: {
      host: '0.0.0.0',
      port: 8080,
      hot: 'only',
      liveReload: false,
      server: 'https',
      static: [{ directory: path.join(__dirname, 'dist'), publicPath: '/ui', serveIndex: false }],
      allowedHosts: 'all',
      historyApiFallback: {
        index: `${publicPath}index.html`,
        verbose: true,
        rewrites: [
          { from: /^\/ui\/env\.js$/, to: false },
          { from: /^\/ui\/.*$/, to: `${publicPath}index.html` }
        ]
      },
      devMiddleware: { publicPath, writeToDisk: false },
      setupMiddlewares: (middlewares, devServer) => {
        devServer.app.get(['/tags.json', '/versions.json'], (req, res) =>
          res.sendFile(path.join(__dirname, `dist${req.path}`), err => err && res.status(404).send('File not found'))
        );
        return middlewares;
      },
      proxy: [{ context: ['/api'], target: 'https://docker.mender.io', secure: false, changeOrigin: true }]
    },
    entry: './src/js/main.tsx',
    module: {
      rules: [
        {
          test: /\.m?[jt]sx?$/,
          exclude: [/node_modules/, /\.test\./, /__snapshots__/],
          resolve: { fullySpecified: false },
          loader: 'esbuild-loader',
          options: {
            loader: 'tsx',
            jsx: 'automatic'
          }
        },
        {
          test: /\.css$/,
          sideEffects: true,
          use: [
            rspack.CssExtractRspackPlugin.loader,
            {
              loader: 'css-loader',
              options: {
                sourceMap: true,
                url: true
              }
            },
            {
              loader: 'postcss-loader',
              options: {
                postcssOptions: { plugins: [autoprefixer({})] },
                sourceMap: true
              }
            },
            {
              loader: 'esbuild-loader',
              options: {
                loader: 'css',
                minify: true
              }
            }
          ]
        },
        {
          test: /\.(png|jpe?g|gif|eot|ttf|woff|woff2)$/i,
          exclude: [/node_modules/, /\.test\./, /__snapshots__/],
          type: 'asset'
        },
        {
          test: /\.svg$/i,
          exclude: [/node_modules/, /\.test\./, /__snapshots__/],
          issuer: /\.[jt]sx?$/,
          use: ['@svgr/webpack']
        }
      ]
    },
    output: {
      filename: '[name].min.js',
      hashFunction: 'xxhash64',
      path: path.resolve(__dirname, 'dist'),
      publicPath,
      sourceMapFilename: '[file].map'
    },
    plugins: [
      new CleanWebpackPlugin({
        cleanOnceBeforeBuildPatterns: ['**/*', '!env.js', '!tags.json', '!versions.json'],
        cleanAfterEveryBuildPatterns: ['!assets/fonts/*', '!assets/img/*']
      }),
      new rspack.CopyRspackPlugin({
        patterns: [
          { from: 'node_modules/monaco-editor/min/vs/', to: 'vs' },
          argv.mode !== 'production' && { from: 'node_modules/monaco-editor/min-maps/vs/', to: 'min-maps/vs' }
        ].filter(Boolean)
      }),
      new rspack.ProvidePlugin({
        process: 'process/browser',
        Buffer: ['buffer', 'Buffer']
      }),
      new rspack.DefinePlugin({
        ENV: JSON.stringify(argv.mode),
        XTERM_VERSION: JSON.stringify(require('./package.json').dependencies['@xterm/xterm']),
        XTERM_FIT_VERSION: JSON.stringify(require('./package.json').dependencies['@xterm/addon-fit']),
        XTERM_SEARCH_VERSION: JSON.stringify(require('./package.json').dependencies['@xterm/addon-search'])
      }),
      new rspack.HtmlRspackPlugin({
        favicon: './src/favicon.svg',
        hash: true,
        template: './src/index.html'
      }),
      new rspack.CssExtractRspackPlugin({
        filename: '[name].css',
        chunkFilename: '[id].css'
      }),
      ...plugins
    ],
    resolve: {
      alias: {
        '@babel/runtime/helpers/esm': path.resolve(__dirname, 'node_modules/@babel/runtime/helpers/esm')
      },
      extensions: ['.ts', '.tsx', '.js', '.jsx'],
      fallback: {
        assert: require.resolve('assert/'),
        buffer: require.resolve('buffer/'),
        stream: require.resolve('stream-browserify'),
        util: require.resolve('util/'),
        vm: require.resolve('vm-browserify'),
        'process/browser': require.resolve('process/browser')
      },
      tsConfig: path.resolve(__dirname, 'tsconfig.json')
    },
    target: 'web',
    profile: true
  };
};
