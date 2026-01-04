import api, { unwrapResponse } from './api';

export type ContentKey = 'content_about' | 'content_help' | 'content_privacy' | 'content_terms';

export interface ContentPagesResp {
  items: Array<{ key: ContentKey; value: string }>;
}

export async function getContentPagesByKeys(keys: ContentKey[]): Promise<Record<ContentKey, string>> {
  const res = await api.get('/api/v1/content/pages', { params: { keys: keys.join(',') } });
  const data = unwrapResponse<ContentPagesResp>(res);
  const map = {} as Record<ContentKey, string>;
  for (const k of keys) map[k] = '';
  for (const it of data.items || []) {
    (map as any)[it.key] = it.value ?? '';
  }
  return map;
}

export async function getPrivacyContent(): Promise<string> {
  const map = await getContentPagesByKeys(['content_privacy']);
  return map['content_privacy'] || '';
}

export async function getAboutContent(): Promise<string> {
  const map = await getContentPagesByKeys(['content_about']);
  return map['content_about'] || '';
}

export async function getHelpContent(): Promise<string> {
  const map = await getContentPagesByKeys(['content_help']);
  return map['content_help'] || '';
}

export async function getTermsContent(): Promise<string> {
  const map = await getContentPagesByKeys(['content_terms']);
  return map['content_terms'] || '';
}
