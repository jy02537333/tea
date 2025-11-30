"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.default = Thumbnail;
const jsx_runtime_1 = require("react/jsx-runtime");
const react_1 = require("react");
const components_1 = require("@tarojs/components");
function Thumbnail({ src, width = 40, height = 40, radius = 4, skeleton = true, shimmerHighlight = '#f5f5f5', shimmerBg = '#eaeaea', shimmerSpeed = 1.2, loading = 'lazy', animateScale = true, }) {
    const [loaded, setLoaded] = (0, react_1.useState)(false);
    const containerStyle = { width: `${width}px`, height: `${height}px`, borderRadius: `${radius}px`, marginRight: '6px', overflow: 'hidden' };
    const imgStyle = { width: '100%', height: '100%', opacity: loaded ? 1 : 0, transform: loaded ? 'scale(1)' : (animateScale ? 'scale(1.03)' : 'scale(1)'), transition: 'opacity 320ms ease-in-out, transform 320ms ease-in-out', backgroundColor: '#f0f0f0' };
    // css vars for global stylesheet - components can set root vars
    const cssVars = {
        ['--shimmer-duration']: `${shimmerSpeed}s`,
        ['--shimmer-bg']: shimmerBg,
        ['--shimmer-highlight']: shimmerHighlight,
    };
    return ((0, jsx_runtime_1.jsxs)(components_1.View, { style: { ...containerStyle, ...cssVars }, children: [skeleton ? (0, jsx_runtime_1.jsx)(components_1.View, { className: "thumbnail-skeleton", style: { width: '100%', height: '100%' } }) : null, src ? (0, jsx_runtime_1.jsx)(components_1.Image, { src: src, style: imgStyle, mode: "aspectFill", lazyLoad: loading === 'lazy', onLoad: () => setLoaded(true) }) : null] }));
}
