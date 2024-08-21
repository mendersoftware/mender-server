import autoprefixer from 'autoprefixer';
import { CleanWebpackPlugin } from 'clean-webpack-plugin';
import CopyWebpackPlugin from 'copy-webpack-plugin';
import ESLintPlugin from 'eslint-webpack-plugin';
import HtmlWebPackPlugin from 'html-webpack-plugin';
import MiniCssExtractPlugin from 'mini-css-extract-plugin';
import { createRequire } from 'module';
import path from 'path';
import TsconfigPathsPlugin from 'tsconfig-paths-webpack-plugin';
import { fileURLToPath } from 'url';
import webpack from 'webpack';

const require = createRequire(import.meta.url);
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export default (env, argv) => {
  const devPlugins = argv.mode === 'production' ? [] : [new ESLintPlugin()];
  return {
    devtool: 'source-map',
    node: {
      global: true
    },
    entry: './src/js/main.js',
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
            MiniCssExtractPlugin.loader,
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
    optimization: {
      minimize: argv.mode === 'production'
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
      new CopyWebpackPlugin({
        patterns: [
          { from: 'node_modules/monaco-editor/min/vs/', to: 'vs' },
          argv.mode !== 'production' && { from: 'node_modules/monaco-editor/min-maps/vs/', to: 'min-maps/vs' }
        ].filter(Boolean)
      }),
      new webpack.ProvidePlugin({
        process: 'process/browser',
        Buffer: ['buffer', 'Buffer']
      }),
      new webpack.DefinePlugin({
        ENV: JSON.stringify(argv.mode),
        XTERM_VERSION: JSON.stringify(require('./package.json').dependencies['@xterm/xterm']),
        XTERM_FIT_VERSION: JSON.stringify(require('./package.json').dependencies['@xterm/addon-fit']),
        XTERM_SEARCH_VERSION: JSON.stringify(require('./package.json').dependencies['@xterm/addon-search'])
      }),
      new HtmlWebPackPlugin({
        favicon: './src/favicon.svg',
        hash: true,
        template: './src/index.html'
      }),
      new MiniCssExtractPlugin({
        filename: '[name].css',
        chunkFilename: '[id].css'
      }),
      ...devPlugins
    ],
    resolve: {
      alias: {
        '@babel/runtime/helpers/esm': path.resolve(__dirname, 'node_modules/@babel/runtime/helpers/esm')
      },
      extensions: ['.ts', '.tsx', '.js', '.jsx'],
      fallback: {
        assert: require.resolve('assert/'),
        buffer: require.resolve('buffer/'),
        crypto: 'crypto-browserify',
        stream: require.resolve('stream-browserify'),
        util: require.resolve('util/'),
        'process/browser': require.resolve('process/browser')
      },
      plugins: [
        new TsconfigPathsPlugin({
          configFile: 'tsconfig.json',
          extensions: ['.ts', '.tsx', '.js', '.jsx']
        })
      ]
    },
    target: 'web'
  };
};
