// Use runtime require to avoid TypeScript including package source in admin-fe compilation
const _shared = require('../../../packages/ui-thumbnail/react') as any;
export default (_shared && _shared.__esModule) ? _shared.default : _shared;
