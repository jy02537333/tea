// Use runtime require to load local dist so Taro's resolver treats it as project file
const _shared = require('../../../packages/ui-thumbnail/dist/taro');
export default (_shared && _shared.__esModule) ? _shared.default : _shared;
