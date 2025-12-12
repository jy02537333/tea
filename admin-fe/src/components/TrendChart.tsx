import React from 'react';

export interface TrendPoint {
  date: string;
  order_count: number;
  paid_order_count: number;
  sales_amount: number;
}

export interface TrendChartProps {
  data: TrendPoint[];
  height?: number;
  colors?: {
    order?: string;
    paid?: string;
  };
}

// 轻量级 SVG 折线图（双折线：订单数/已支付订单数），无需额外依赖
export default function TrendChart({ data, height = 260, colors }: TrendChartProps) {
  const padding = { top: 16, right: 24, bottom: 32, left: 40 };
  const width = 800; // 宽度交由容器控制，这里用于坐标换算
  const innerW = width - padding.left - padding.right;
  const innerH = height - padding.top - padding.bottom;

  const dates = data.map((d) => d.date);
  const maxY = Math.max(1, ...data.map((d) => Math.max(d.order_count, d.paid_order_count)));

  const x = (i: number) => (innerW * i) / Math.max(1, data.length - 1);
  const y = (v: number) => innerH - (v / maxY) * innerH;

  const toPath = (values: number[]) =>
    values
      .map((v, i) => `${i === 0 ? 'M' : 'L'} ${padding.left + x(i)} ${padding.top + y(v)}`)
      .join(' ');

  const orderPath = toPath(data.map((d) => d.order_count));
  const paidPath = toPath(data.map((d) => d.paid_order_count));

  const ticks = 4;
  const yTicks = Array.from({ length: ticks + 1 }, (_, i) => Math.round((maxY * i) / ticks));

  const colorOrder = colors?.order || '#1677ff';
  const colorPaid = colors?.paid || '#52c41a';

  return (
    <div style={{ width: '100%', overflowX: 'auto' }}>
      <svg width={width} height={height} role="img" aria-label="订单趋势折线图">
        {/* 坐标系 */}
        <g transform={`translate(${padding.left}, ${padding.top})`}>
          {/* Y 轴网格与标签 */}
          {yTicks.map((t) => (
            <g key={t}>
              <line
                x1={0}
                x2={innerW}
                y1={y(t)}
                y2={y(t)}
                stroke="#eee"
              />
              <text x={-8} y={y(t)} textAnchor="end" dominantBaseline="middle" fill="#888" fontSize={12}>
                {t}
              </text>
            </g>
          ))}

          {/* X 轴标签（首末与中间点） */}
          {data.map((d, i) => {
            const shouldShow = i === 0 || i === data.length - 1 || i === Math.floor(data.length / 2);
            if (!shouldShow) return null;
            return (
              <text
                key={d.date}
                x={x(i)}
                y={innerH + 20}
                textAnchor="middle"
                fill="#888"
                fontSize={12}
              >
                {d.date.slice(5)}
              </text>
            );
          })}

          {/* 折线：订单总数 */}
          <path d={orderPath} fill="none" stroke={colorOrder} strokeWidth={2} />
          {/* 折线：已支付订单数 */}
          <path d={paidPath} fill="none" stroke={colorPaid} strokeWidth={2} />

          {/* 数据点（小圆） */}
          {data.map((d, i) => (
            <circle key={`o-${i}`} cx={x(i)} cy={y(d.order_count)} r={2.5} fill={colorOrder} />
          ))}
          {data.map((d, i) => (
            <circle key={`p-${i}`} cx={x(i)} cy={y(d.paid_order_count)} r={2.5} fill={colorPaid} />
          ))}
        </g>

        {/* 图例 */}
        <g transform={`translate(${padding.left}, ${padding.top - 6})`}>
          <rect x={0} y={-14} width={10} height={2} fill={colorOrder} />
          <text x={16} y={-10} fill="#555" fontSize={12}>
            订单数
          </text>
          <rect x={64} y={-14} width={10} height={2} fill={colorPaid} />
          <text x={80} y={-10} fill="#555" fontSize={12}>
            已支付订单数
          </text>
        </g>
      </svg>
    </div>
  );
}
