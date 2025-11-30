"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.default = Thumbnail;
const jsx_runtime_1 = require("react/jsx-runtime");
const react_1 = require("react");
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
