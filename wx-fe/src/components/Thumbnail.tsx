// runtime require of built dist entry so bundlers use package's main
const _shared = require('../../../packages/ui-thumbnail/dist/taro') as any;
export default (_shared && _shared.__esModule) ? _shared.default : _shared;
