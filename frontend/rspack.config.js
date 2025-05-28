import { rspack } from '@rspack/core';
import autoprefixer from 'autoprefixer';
import { CleanWebpackPlugin } from 'clean-webpack-plugin';
import CompressionPlugin from 'compression-webpack-plugin';
import ESLintPlugin from 'eslint-rspack-plugin';
import { createRequire } from 'module';
import path from 'path';
import { fileURLToPath } from 'url';
import LicensePlugin from 'webpack-license-plugin';

const require = createRequire(import.meta.url);
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export default (env, argv) => {
  const plugins =
    argv.mode === 'production'
      ? [
          new LicensePlugin({
            outputFilename: 'licenses.json',
            excludedPackageTest: packageName => packageName.startsWith('@northern.tech'),
            replenishDefaultLicenseTexts: true
          }),
          new CompressionPlugin({
            filename: '[path][base].gz'
          })
        ]
      : [new ESLintPlugin({ extensions: ['js', 'ts', 'tsx'] })];
  return {
    devtool: 'source-map',
    node: {
      global: true
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
          test: /\.(less|css)$/,
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
            },
            'less-loader'
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
      publicPath: '/ui/'
    },
    plugins: [
      new CleanWebpackPlugin({
        cleanOnceBeforeBuildPatterns: ['**/*', '!env.js'],
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
