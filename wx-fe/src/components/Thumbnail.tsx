// runtime require to avoid including package TS sources in wx-fe tsc program
const _shared = require('../../../packages/ui-thumbnail/taro') as any;
export default (_shared && _shared.__esModule) ? _shared.default : _shared;
