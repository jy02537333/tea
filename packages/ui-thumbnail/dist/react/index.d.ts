import './thumbnail.css';
export interface ReactThumbnailProps {
    src?: string | null;
    alt?: string;
    width?: number;
    height?: number;
    radius?: number;
    skeleton?: boolean;
    shimmerHighlight?: string;
    shimmerBg?: string;
    shimmerSpeed?: number;
    loading?: 'lazy' | 'eager';
    animateScale?: boolean;
}
export default function Thumbnail({ src, alt, width, height, radius, skeleton, shimmerHighlight, shimmerBg, shimmerSpeed, loading, animateScale, }: ReactThumbnailProps): import("react/jsx-runtime").JSX.Element;
