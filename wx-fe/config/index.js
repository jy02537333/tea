const path = require('path');

module.exports = {
  projectName: 'wx-fe',
  date: '2025-12-01',
  designWidth: 750,
  deviceRatio: {
    '640': 2.34,
    '750': 1,
    '828': 1.81
  },
  sourceRoot: 'src',
  outputRoot: 'dist',
  plugins: [],
  defineConstants: {},
  copy: {
    patterns: [],
    options: {}
  },
  framework: 'react',
  // Disable prebundle to avoid webpack prebundle plugin errors in CI/local
  compiler: {
    prebundle: {
      enable: false
    }
  },
  h5: {
    publicPath: '/',
    staticDirectory: 'static',
    prebundle: {
      enable: false
    }
  },
  mini: {}
};
