export interface TaroThumbnailProps {
    src?: string | null;
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
export default function Thumbnail({ src, width, height, radius, skeleton, shimmerHighlight, shimmerBg, shimmerSpeed, loading, animateScale, }: TaroThumbnailProps): any;
