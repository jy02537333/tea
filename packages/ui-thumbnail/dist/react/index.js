// wrapper that re-exports the source entry so bundlers can process TSX
module.exports = require('../../react/index.tsx');
"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
exports.default = Thumbnail;
const jsx_runtime_1 = require("react/jsx-runtime");
const react_1 = __importStar(require("react"));
require("./thumbnail.css");
function Thumbnail({ src, alt = '', width = 40, height = 40, radius = 4, skeleton = true, shimmerHighlight = '#f5f5f5', shimmerBg = '#eaeaea', shimmerSpeed = 1.2, loading = 'lazy', animateScale = true, }) {
    const [loaded, setLoaded] = (0, react_1.useState)(false);
    const style = { width, height, borderRadius: radius, overflow: 'hidden', marginRight: 6 };
    const cssVars = {
        ['--af-shimmer-bg']: shimmerBg,
        ['--af-shimmer-highlight']: shimmerHighlight,
        ['--af-shimmer-duration']: `${shimmerSpeed}s`,
    };
    return ((0, jsx_runtime_1.jsxs)("div", { style: { ...style, ...cssVars }, className: `af-thumbnail-container ${loaded ? 'af-loaded' : ''}`, children: [skeleton && (0, jsx_runtime_1.jsx)("div", { className: "af-thumbnail-skeleton" }), src ? ((0, jsx_runtime_1.jsx)("img", { src: src, alt: alt, style: { width: '100%', height: '100%', objectFit: 'cover', opacity: loaded ? 1 : 0, transform: loaded ? 'scale(1)' : (animateScale ? 'scale(1.03)' : 'scale(1)'), transition: 'opacity 320ms ease, transform 320ms ease' }, loading: loading, onLoad: () => setLoaded(true) })) : null] }));
}
