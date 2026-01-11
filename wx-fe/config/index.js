const config = {
  projectName: 'wx-fe',
  date: '2025-12-06',
  designWidth: 375,
  deviceRatio: {
    640: 2.34 / 2,
    750: 2,
    828: 1.81 / 2,
  },
  sourceRoot: 'src',
  outputRoot: 'dist',
  plugins: ['@tarojs/plugin-html'],
  framework: 'react',
  compiler: {
    type: 'webpack5',
    prebundle: {
      enable: false,
    },
  },
  cache: {
    enable: true,
  },
  defineConstants: {
    WX_API_BASE_URL: JSON.stringify(process.env.WX_API_BASE_URL || ''),
  },
  copy: {
    patterns: [],
    options: {},
  },
  mini: {
    postcss: {
      pxtransform: {
        enable: true,
        config: {},
      },
      cssModules: {
        enable: false,
      },
    },
  },
  h5: {
    publicPath: '/',
    staticDirectory: 'static',
    router: {
      mode: 'hash',
    },
    devServer: {
      host: '0.0.0.0',
      port: 10088,
      open: false,
    },
    webpackChain(chain) {
      chain.optimization.splitChunks({
        chunks: 'all',
        cacheGroups: {
          react: {
            test: /[\\/]node_modules[\\/](react|react-dom)[\\/]/,
            name: 'react',
            priority: 30,
            enforce: true,
          },
          taro: {
            test: /[\\/]node_modules[\\/]@tarojs[\\/]/,
            name: 'taro',
            priority: 25,
            enforce: true,
          },
          utils: {
            test: /[\\/]node_modules[\\/](qrcode|marked|dompurify)[\\/]/,
            name: 'utils',
            priority: 20,
            enforce: true,
          },
          vendors: {
            test: /[\\/]node_modules[\\/]/,
            name: 'vendors',
            priority: 10,
          },
        },
      });

      // Filter noisy cache serialization warnings
      chain.set('ignoreWarnings', [
        (warning) => {
          const msg = (warning && (warning.message || warning)) || '';
          return /No serializer registered for NullDependency/.test(msg);
        },
      ]);
      chain.set('stats', {
        warningsFilter: [/No serializer registered for NullDependency/],
      });
      chain.set('infrastructureLogging', { level: 'error' });
    },
  },
};

module.exports = config;
