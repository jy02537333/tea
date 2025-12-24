export interface WithdrawRemarkFields {
  phase?: string;
  currency?: string;
  amount_cents?: number | string;
  fee_cents?: number | string;
  net_cents?: number | string;
}

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
    // ignore invalid JSON
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
