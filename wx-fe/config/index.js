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
    enable: false,
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
      mode: 'browser',
    },
    devServer: {
      host: '0.0.0.0',
      port: 10088,
      open: false,
    },
  },
};

module.exports = config;
