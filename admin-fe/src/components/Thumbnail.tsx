// Load built package entry (dist) so bundlers and runtime use package main
const _shared = require('../../../packages/ui-thumbnail/dist/react') as any;
export default (_shared && _shared.__esModule) ? _shared.default : _shared;
