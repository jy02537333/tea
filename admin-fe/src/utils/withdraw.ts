export interface WithdrawRemarkFields {
  phase?: string;
  currency?: string;
  amount_cents?: number | string;
  fee_cents?: number | string;
  net_cents?: number | string;
}

import type { ColumnsType } from 'antd/es/table';
import type { StoreWithdrawRecord } from '../services/stores';

export function parseWithdrawRemark(remark?: string): WithdrawRemarkFields | null {
  try {
    const obj = remark ? JSON.parse(remark) : null;
    if (obj && typeof obj === 'object') {
      return {
        phase: (obj as any).phase,
        currency: (obj as any).currency,
        amount_cents: (obj as any).amount_cents,
        fee_cents: (obj as any).fee_cents,
        net_cents: (obj as any).net_cents,
      };
    }
  } catch {
    // ignore
  }
  return null;
}

export function getRemarkField(
  remark: string | undefined,
  key: keyof WithdrawRemarkFields,
  fallback: string | number = '-',
): string | number {
  const obj = parseWithdrawRemark(remark);
  const val = obj?.[key];
  return val != null && val !== '' ? val : fallback;
}

export function getRemarkFieldsForCsv(remark?: string): Array<string | number> {
  const obj = parseWithdrawRemark(remark);
  const phase = obj?.phase ?? '-';
  const currency = obj?.currency ?? '-';
  const amountCents = obj?.amount_cents ?? '-';
  const feeCents = obj?.fee_cents ?? '-';
  const netCents = obj?.net_cents ?? '-';
  return [phase, currency, amountCents, feeCents, netCents];
}

export function buildWithdrawRemarkColumns(): ColumnsType<StoreWithdrawRecord> {
  return [
    {
      title: '阶段',
      key: 'remark_phase',
      width: 100,
      render: (_, record) => getRemarkField(record.remark, 'phase'),
    },
    {
      title: '币种',
      key: 'remark_currency',
      width: 80,
      render: (_, record) => getRemarkField(record.remark, 'currency'),
    },
    {
      title: '金额(分)',
      key: 'remark_amount_cents',
      width: 120,
      render: (_, record) => getRemarkField(record.remark, 'amount_cents'),
    },
    {
      title: '手续费(分)',
      key: 'remark_fee_cents',
      width: 120,
      render: (_, record) => getRemarkField(record.remark, 'fee_cents'),
    },
    {
      title: '实付(分)',
      key: 'remark_net_cents',
      width: 120,
      render: (_, record) => getRemarkField(record.remark, 'net_cents'),
    },
  ];
}
