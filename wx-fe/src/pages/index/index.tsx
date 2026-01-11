import React, { useEffect, useMemo, useState } from 'react';
import { View, Text, Button, Swiper, SwiperItem, Image, Picker } from '@tarojs/components';
import Taro, { useRouter } from '@tarojs/taro';
import { getProducts } from '../../services/products';
import { listStores } from '../../services/stores';
import { listCategories } from '../../services/categories';
import { Product, Store, Category } from '../../services/types';
import { listBanners, Banner } from '../../services/banners';
import { getMeSummary } from '../../services/me';
import { getAboutContent } from '../../services/content';
import ContentRenderer from '../../components/ContentRenderer';
import ProductCard from '../../components/ProductCard';
import { listSiteConfigs } from '../../services/site';
import './index.scss';

export default function IndexPage() {
  const router = useRouter();

  // 运营位：轮播
  const [banners, setBanners] = useState<Banner[]>([]);

  // 商品展示区（简介/推荐/本地/版权电话）
  const [aboutMarkdown, setAboutMarkdown] = useState<string>('');
  const [recommended, setRecommended] = useState<Product[]>([]);
  const [localProducts, setLocalProducts] = useState<Product[]>([]);
  const [siteCopyright, setSiteCopyright] = useState<string>('');
  const [sitePhone, setSitePhone] = useState<string>('');

  // 会员横幅
  const [loggedIn, setLoggedIn] = useState(false);
  const [isMember, setIsMember] = useState(false);

  // 门店

  const [stores, setStores] = useState<Store[]>([]);
  const [currentStoreId, setCurrentStoreId] = useState<number | undefined>(undefined);
  const [loading, setLoading] = useState(false);
  const [resCheck, setResCheck] = useState<{ running: boolean; ok: number; total: number; errors: Array<{ url: string; status?: number; message: string }> }>({ running: false, ok: 0, total: 0, errors: [] });
  const [appPages, setAppPages] = useState<string[]>([]);

  // 商城：分区/上新/热销
  const [categories, setCategories] = useState<Category[]>([]);
  const [newProducts, setNewProducts] = useState<Product[]>([]);
  const [hotProducts, setHotProducts] = useState<Product[]>([]);

  useEffect(() => {
    // 初始化加载附近门店和商品列表
    void initFromParamsAndLoad();
    // 首页资源自检（chunk/js 是否可访问）
    void runResourceSelfCheck();
  }, []);

  useEffect(() => {
    // 当前门店变化时，刷新本地商品
    if (currentStoreId) {
      void fetchLocalProducts(currentStoreId);
    }
  }, [currentStoreId]);

  async function initFromParamsAndLoad() {
    const paramSidRaw = router?.params?.store_id;
    const paramSid = paramSidRaw ? Number(paramSidRaw) : NaN;
    if (!Number.isNaN(paramSid) && paramSid > 0) {
      setCurrentStoreId(paramSid);
      try { Taro.setStorageSync('current_store_id', String(paramSid)); } catch (_) {}
    }
    await loadHome();
  }

  async function loadHome() {
    setLoading(true);
    try {
      await Promise.all([fetchStores(), fetchOps(), fetchMembership(), fetchMallBlocks()]);
    } finally {
      setLoading(false);
    }
  }

  async function runResourceSelfCheck() {
    try {
      setResCheck({ running: true, ok: 0, total: 0, errors: [] });
      const origin = (globalThis as any)?.location?.origin || '';
      const assets = await buildAssetListFromAppJs(origin);
      // 解析 app.js 的 pages 列表
      void loadAppPagesInfo(origin).catch(() => {});
      const toUrl = (p: string) => {
        if (!origin) return p;
        try {
          const u = new URL(p, origin);
          return u.toString();
        } catch (_) {
          return origin.replace(/\/$/, '') + p;
        }
      };
      const total = assets.length;
      let ok = 0;
      const errors: Array<{ url: string; status?: number; message: string }> = [];
      for (const p of assets) {
        const url = toUrl(p);
        const head = await fetchWithTimeout(url, { method: 'HEAD', cache: 'no-store' }, 4000).catch((e) => e);
        if (head && head.ok) { ok++; continue; }
        const get = await fetchWithTimeout(url, { method: 'GET', cache: 'no-store' }, 6000).catch((e) => e);
        if (get && (get as Response).ok) { ok++; continue; }
        const status = (head && head.status) || (get && (get as any).status);
        const message = (head && head.statusText) || (get && (get as any).statusText) || 'network error';
        errors.push({ url, status, message });
      }
      setResCheck({ running: false, ok, total, errors });
    } catch (e: any) {
      setResCheck({ running: false, ok: 0, total: 0, errors: [{ url: 'self-check', message: e?.message || 'unknown error' }] });
    }
  }

  async function buildAssetListFromAppJs(origin: string): Promise<string[]> {
    const fallback = ['/js/app.js', '/chunk/22.js', '/chunk/200.js', '/chunk/402.js'];
    try {
      const appUrl = origin ? new URL('/js/app.js', origin).toString() : '/js/app.js';
      const res = await fetchWithTimeout(appUrl, { method: 'GET', cache: 'no-store' }, 5000);
      if (!res.ok) return fallback;
      const text = await res.text();
      const chunkIds = new Set<string>();
      const reChunk = /t\.e\((\d+)\)/g; // match dynamic import chunk ids
      let m: RegExpExecArray | null;
      while ((m = reChunk.exec(text))) {
        const id = m[1];
        if (id) chunkIds.add(`/chunk/${id}.js`);
        if (chunkIds.size > 50) break; // safety cap
      }
      const assets = ['/js/app.js', ...Array.from(chunkIds)];
      return assets.length > 1 ? assets : fallback;
    } catch (_) {
      return fallback;
    }
  }

  async function loadAppPagesInfo(origin: string): Promise<void> {
    try {
      const appUrl = origin ? new URL('/js/app.js', origin).toString() : '/js/app.js';
      const res = await fetchWithTimeout(appUrl, { method: 'GET', cache: 'no-store' }, 5000);
      if (!res.ok) return;
      const text = await res.text();
      const pages: string[] = [];
      const m = text.match(/"pages"\s*:\s*\[(.*?)\]/s);
      if (m && m[1]) {
        const inside = m[1];
        const reStr = /"([^"]+)"/g;
        let mm: RegExpExecArray | null;
        while ((mm = reStr.exec(inside))) {
          pages.push(mm[1]);
          if (pages.length > 200) break; // safety cap
        }
      }
      if (pages.length > 0) setAppPages(pages);
    } catch (_) {
      // ignore parse errors
    }
  }

  function fetchWithTimeout(url: string, init: RequestInit, timeoutMs: number): Promise<Response> {
    const controller = new AbortController();
    const id = setTimeout(() => controller.abort(), timeoutMs);
    return fetch(url, { ...init, signal: controller.signal }).finally(() => clearTimeout(id));
  }

  async function fetchOps() {
    try {
      const [bannerList, aboutText, siteList, recommendedRes] = await Promise.all([
        listBanners(10).catch(() => []),
        getAboutContent().catch(() => ''),
        listSiteConfigs({ keys: ['site_copyright', 'site_phone'] }).catch(() => []),
        getProducts({ page: 1, limit: 6 }).catch(() => ({ data: [] } as any)),
      ]);

      setBanners(Array.isArray(bannerList) ? bannerList : []);
      setAboutMarkdown(typeof aboutText === 'string' ? aboutText : '');

      const map: Record<string, string> = {};
      for (const it of siteList || []) {
        if (it?.config_key) map[it.config_key] = it.config_value ?? '';
      }
      setSiteCopyright(map['site_copyright'] || '');
      setSitePhone(map['site_phone'] || '');

      const maybe: any = recommendedRes;
      let items: Product[] = [];
      if (Array.isArray(maybe?.data)) items = maybe.data;
      else if (Array.isArray(maybe?.items)) items = maybe.items;
      else if (Array.isArray(maybe)) items = maybe;
      setRecommended(items);
    } catch (e) {
      console.error('load banners failed', e);
    }
  }

  async function fetchLocalProducts(storeId: number) {
    try {
      const res = await getProducts({ page: 1, limit: 6, store_id: storeId });
      const maybe: any = res;
      let items: Product[] = [];
      if (Array.isArray(maybe?.data)) items = maybe.data;
      else if (Array.isArray(maybe?.items)) items = maybe.items;
      else if (Array.isArray(maybe)) items = maybe;
      setLocalProducts(items);
    } catch (e) {
      console.error('load local products failed', e);
    }
  }

  async function fetchMembership() {
    // 不强制要求该接口存在：失败则视为未登录/非会员
    try {
      const s: any = await getMeSummary();

      // 兼容两种后端返回形态：
      // A) { user, wallet, points, ... }
      // B) { user_id, nickname, wallet_balance, points, coupons, membership }
      const hasUser = Boolean(s?.user?.id) || Boolean(s?.user_id);
      setLoggedIn(hasUser);

      const level: string =
        String(
          s?.membership ||
          s?.data?.membership ||
          s?.membership_level ||
          s?.membershipLevel ||
          '',
        ).trim();
      setIsMember(Boolean(level) && level !== 'visitor' && level !== 'none');
    } catch (_err) {
      setLoggedIn(false);
      setIsMember(false);
    }
  }

  async function fetchStores() {
    try {
      const res = await listStores({ page: 1, limit: 20 });
      const maybe: any = res;
      let items: Store[] = [];
      if (Array.isArray(maybe?.data)) items = maybe.data;
      else if (Array.isArray(maybe?.items)) items = maybe.items;
      else if (Array.isArray(maybe)) items = maybe;
      setStores(items);
      if (!currentStoreId && items.length > 0) setCurrentStoreId(items[0].id);
    } catch (e) {
      console.error('load stores failed', e);
    }
  }

  const storePickerRange = useMemo(() => ['附近门店', ...stores.map((s) => s.name)], [stores]);
  const storePickerIndex = useMemo(() => {
    if (!currentStoreId) return 0;
    const idx = stores.findIndex((s) => s.id === currentStoreId);
    return idx >= 0 ? idx + 1 : 0;
  }, [stores, currentStoreId]);

  function handleStorePickerChange(e: any) {
    const index = Number(e?.detail?.value ?? 0);
    const storeId = index === 0 ? undefined : stores[index - 1]?.id;
    setCurrentStoreId(storeId);
    try {
      if (storeId) Taro.setStorageSync('current_store_id', String(storeId));
      else Taro.removeStorageSync('current_store_id');
    } catch (_) {}
  }

  function goProductDetail(id: number, storeId?: number) {
    const storeQuery = storeId ? `&store_id=${storeId}` : '';
    Taro.navigateTo({ url: `/pages/product-detail/index?id=${id}${storeQuery}` }).catch(() => {});
  }

  function goLogin() {
    Taro.navigateTo({ url: '/pages/login/index' }).catch(() => {});
  }

  function ensureLoggedIn() {
    // 首页首次渲染可能尚未完成 fetchMembership，这里额外用 token 兜底判断
    let hasToken = false;
    try {
      hasToken = Boolean(Taro.getStorageSync('token'));
    } catch (_) {}
    if (loggedIn || hasToken) return true;
    Taro.showToast({ title: '请先登录', icon: 'none' });
    goLogin();
    return false;
  }

  function goMembership() {
    if (!loggedIn) {
      Taro.showToast({ title: '请先登录', icon: 'none' });
      goLogin();
      return;
    }
    Taro.navigateTo({ url: '/pages/membership/index' }).catch(() => {});
  }

  function goActivities() {
    Taro.navigateTo({ url: '/pages/activities/index' }).catch(() => {});
  }

  function goShareCenter() {
    if (!ensureLoggedIn()) return;
    Taro.navigateTo({ url: '/pages/share/index' }).catch(() => {});
  }

  function goCoupons() {
    if (!ensureLoggedIn()) return;
    Taro.navigateTo({ url: '/pages/coupons/index' }).catch(() => {});
  }

  function goMall() {
    const sid = currentStoreId;
    const storeQuery = sid ? `store_id=${sid}` : '';
    const q = storeQuery ? `?${storeQuery}` : '';
    Taro.navigateTo({ url: `/pages/category/index${q}` }).catch(() => {});
  }

  function goDineIn() {
    const sid = currentStoreId;
    if (!sid) {
      Taro.showToast({ title: '请先选择门店', icon: 'none' });
      return;
    }
    Taro.navigateTo({ url: `/pages/product-list/index?store_id=${sid}` }).catch(() => {});
  }

  function goDelivery() {
    const sid = currentStoreId;
    if (!sid) {
      Taro.showToast({ title: '请先选择门店', icon: 'none' });
      return;
    }
    Taro.navigateTo({ url: `/pages/product-list/index?store_id=${sid}` }).catch(() => {});
  }

  function onBannerClick(b: Banner) {
    const linkType = b.link_type || 1;
    const linkURL = (b.link_url || '').trim();
    if (!linkURL) return;

    // 1:无链接 2:商品详情 3:分类页 4:外部链接
    if (linkType === 2) {
      // link_url 约定：优先支持纯数字商品ID；其次支持包含 id=xx 的 query
      const asNumber = Number(linkURL);
      if (!Number.isNaN(asNumber) && asNumber > 0) {
        goProductDetail(asNumber, currentStoreId);
        return;
      }
      const m = linkURL.match(/(^|[?&])id=(\d+)/);
      if (m && m[2]) {
        goProductDetail(Number(m[2]), currentStoreId);
      }
      return;
    }

    if (linkType === 3) {
      const asNumber = Number(linkURL);
      if (!Number.isNaN(asNumber) && asNumber > 0) {
        const sid = currentStoreId;
        const storeQuery = sid ? `&store_id=${sid}` : '';
        Taro.navigateTo({ url: `/pages/category/index?category_id=${asNumber}${storeQuery}` }).catch(() => {});
        return;
      }
      const m = linkURL.match(/(^|[?&])category_id=(\d+)/);
      if (m && m[2]) {
        const sid = currentStoreId;
        const storeQuery = sid ? `&store_id=${sid}` : '';
        Taro.navigateTo({ url: `/pages/category/index?category_id=${Number(m[2])}${storeQuery}` }).catch(() => {});
      }
      return;
    }

    if (linkType === 4) {
      // 小程序环境下外链通常需 WebView 承载
      const encoded = encodeURIComponent(linkURL);
      Taro.navigateTo({ url: `/pages/webview/index?url=${encoded}` }).catch(async () => {
        try {
          await Taro.setClipboardData({ data: linkURL });
          Taro.showToast({ title: '链接已复制', icon: 'none' });
        } catch (_) {
          Taro.showToast({ title: '无法打开外链', icon: 'none' });
        }
      });
    }
  }

  async function fetchMallBlocks() {
    try {
      const [catRes, prodRes] = await Promise.all([
        listCategories({ status: 1 }).catch(() => ({ data: [] } as any)),
        getProducts({ page: 1, limit: 50 }).catch(() => ({ data: [] } as any)),
      ]);

      const catMaybe: any = catRes;
      const catItems: Category[] = Array.isArray(catMaybe?.data)
        ? catMaybe.data
        : Array.isArray(catMaybe?.items)
          ? catMaybe.items
          : Array.isArray(catMaybe)
            ? catMaybe
            : [];
      setCategories(catItems);

      const prodMaybe: any = prodRes;
      const items: Product[] = Array.isArray(prodMaybe?.data)
        ? prodMaybe.data
        : Array.isArray(prodMaybe?.items)
          ? prodMaybe.items
          : Array.isArray(prodMaybe)
            ? prodMaybe
            : [];

      const createdDesc = [...items].sort((a, b) => {
        const ta = Date.parse(String((a as any).created_at || a.created_at || ''));
        const tb = Date.parse(String((b as any).created_at || b.created_at || ''));
        if (!Number.isFinite(ta) || !Number.isFinite(tb)) return b.id - a.id;
        return tb - ta;
      });

      const now = Date.now();
      const in30Days = (p: Product) => {
        if ((p as any).is_new === true || p.is_new === true) return true;
        const t = Date.parse(String((p as any).created_at || p.created_at || ''));
        if (!Number.isFinite(t)) return false;
        return now-t <= 30 * 24 * 60 * 60 * 1000;
      };
      const newList: Product[] = [];
      for (const p of createdDesc) {
        if (in30Days(p)) newList.push(p);
        if (newList.length >= 4) break;
      }
      if (newList.length < 4) {
        for (const p of createdDesc) {
          if (newList.find((x) => x.id === p.id)) continue;
          newList.push(p);
          if (newList.length >= 4) break;
        }
      }
      setNewProducts(newList);

      const hotDesc = [...items].sort((a, b) => {
        const sa = Number((a as any).sales ?? a.sales ?? 0);
        const sb = Number((b as any).sales ?? b.sales ?? 0);
        if (Number.isFinite(sa) && Number.isFinite(sb) && (sa !== 0 || sb !== 0)) return sb - sa;
        // 兜底：is_hot 优先，其次按 id
        const ha = (a as any).is_hot === true || a.is_hot === true;
        const hb = (b as any).is_hot === true || b.is_hot === true;
        if (ha !== hb) return hb ? 1 : -1;
        return b.id - a.id;
      });
      setHotProducts(hotDesc.slice(0, 4));
    } catch (e) {
      console.error('load mall blocks failed', e);
    }
  }

  const mallPartitions = useMemo(() => {
    const byName = (names: string[]) => {
      const found = categories.find((c) => names.some((n) => (c.name || '').includes(n)));
      return found?.id;
    };
    return [
      { key: 'gift', title: '礼品茶', categoryId: byName(['礼品', '送礼']) },
      { key: 'health', title: '健康茶', categoryId: byName(['健康']) },
      { key: 'staple', title: '口粮茶', categoryId: byName(['口粮']) },
      { key: 'bulk', title: '散茶', categoryId: byName(['散茶', '散装']) },
      { key: 'invest', title: '投资收藏', categoryId: byName(['投资', '收藏']) },
    ];
  }, [categories]);

  function goPartition(categoryId?: number) {
    const sid = currentStoreId;
    const storeQuery = sid ? `&store_id=${sid}` : '';
    if (categoryId) {
      Taro.navigateTo({ url: `/pages/category/index?category_id=${categoryId}${storeQuery}` }).catch(() => {});
      return;
    }
    const q = storeQuery ? `?${storeQuery.replace(/^&/, '')}` : '';
    Taro.navigateTo({ url: `/pages/category/index${q}` }).catch(() => {});
  }

  function productCover(p: Product): string {
    const raw = (p as any).image_url || (p as any).cover || p.images || '';
    const first = String(raw).split(',').map((s) => s.trim()).filter(Boolean)[0];
    return first || 'https://dummyimage.com/300x300/dcdcdc/333333&text=Tea';
  }

  return (
    <View data-testid="page-index" style={{ padding: 12, backgroundColor: '#f5f6f8', minHeight: '100vh' }}>
      {/* 顶部状态条：用于可视化加载与排障 */}
      <View style={{ backgroundColor: '#fff', borderRadius: 12, padding: 8, marginBottom: 12 }}>
        <Text style={{ color: '#333' }}>{loading ? '首页加载中…' : '首页就绪'}</Text>
        <View style={{ marginTop: 6 }}>
          {resCheck.running ? (
            <Text style={{ color: '#666' }}>资源自检进行中…</Text>
          ) : resCheck.errors.length > 0 ? (
            <View>
              <Text style={{ color: '#d4380d' }}>资源检查失败 {resCheck.ok}/{resCheck.total}</Text>
              {resCheck.errors.map((e) => (
                <Text key={e.url} style={{ display: 'block', color: '#d4380d', fontSize: 12 }}>
                  {e.url} → {e.status ?? ''} {e.message}
                </Text>
              ))}
            </View>
          ) : (
            <Text style={{ color: '#389e0d' }}>资源检查通过 {resCheck.ok}/{resCheck.total}</Text>
          )}
          {appPages.length > 0 && (
            <View style={{ marginTop: 6 }}>
              <Text style={{ color: '#666' }}>已注册页面：{appPages.length}</Text>
              <Text style={{ display: 'block', color: '#999', fontSize: 12 }}>
                {appPages.slice(0, 12).join(', ')}{appPages.length > 12 ? ' …' : ''}
              </Text>
            </View>
          )}
        </View>
      </View>
      {/* 首页：Balance 主图 / 轮播 */}
      {banners.length > 0 && (
        <View style={{ marginBottom: 12, borderRadius: 12, overflow: 'hidden' }}>
          <Swiper
            circular
            autoplay
            interval={4000}
            style={{ height: 160 }}
          >
            {banners.map((b) => (
              <SwiperItem key={b.id}>
                <View onClick={() => onBannerClick(b)} style={{ width: '100%', height: 160 }}>
                  <Image
                    src={b.image_url}
                    mode="aspectFill"
                    style={{ width: '100%', height: 160, borderRadius: 8 }}
                  />
                </View>
              </SwiperItem>
            ))}
          </Swiper>
        </View>
      )}

      {/* 首页：会员横幅 */}
      {!isMember && (
        <View style={{ backgroundColor: '#fff', borderRadius: 12, padding: 12, marginBottom: 12 }}>
          {!loggedIn ? (
            <View style={{ display: 'flex', flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }}>
              <Text style={{ color: '#333' }}>登录后可同步会员权益与优惠</Text>
              <Button size="mini" type="primary" onClick={goLogin}>去登录</Button>
            </View>
          ) : (
            <View style={{ display: 'flex', flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }}>
              <Text style={{ color: '#333' }}>开通会员享受优惠</Text>
              <Button size="mini" type="primary" onClick={goMembership}>开通/查看</Button>
            </View>
          )}
        </View>
      )}

      {/* 首页：金刚区（堂食/外卖/商城） */}
      <View style={{ backgroundColor: '#fff', borderRadius: 12, padding: 12, marginBottom: 12 }}>
        <View style={{ display: 'flex', flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
          <Text style={{ fontSize: 16, fontWeight: 'bold' }}>快捷入口</Text>
          <Picker mode="selector" range={storePickerRange} onChange={handleStorePickerChange} value={storePickerIndex}>
            <View style={{ padding: '6px 10px', borderWidth: 1, borderStyle: 'solid', borderColor: '#eee', borderRadius: 16 }}>
              <Text style={{ color: '#666', fontSize: 12 }}>{storePickerRange[storePickerIndex] || '附近门店'}</Text>
            </View>
          </Picker>
        </View>

        <View style={{ display: 'flex', flexDirection: 'row', gap: 10, marginTop: 12 }}>
          <View onClick={goDineIn} style={{ flex: 1, padding: 12, borderRadius: 10, backgroundColor: '#f6ffed' }}>
            <Text style={{ fontWeight: 'bold', display: 'block' }}>堂食</Text>
            <Text style={{ color: '#666', fontSize: 12 }}>选择门店后下单</Text>
          </View>
          <View onClick={goDelivery} style={{ flex: 1, padding: 12, borderRadius: 10, backgroundColor: '#e6f7ff' }}>
            <Text style={{ fontWeight: 'bold', display: 'block' }}>外卖</Text>
            <Text style={{ color: '#666', fontSize: 12 }}>支持配送地址</Text>
          </View>
          <View onClick={goMall} style={{ flex: 1, padding: 12, borderRadius: 10, backgroundColor: '#fff7e6' }}>
            <Text style={{ fontWeight: 'bold', display: 'block' }}>商城</Text>
            <Text style={{ color: '#666', fontSize: 12 }}>分类/筛选</Text>
          </View>
        </View>
      </View>

      {/* 首页：标签区（1+2结构） */}
      <View style={{ backgroundColor: '#fff', borderRadius: 12, padding: 12, marginBottom: 12 }}>
        <Text style={{ fontSize: 16, fontWeight: 'bold' }}>标签区</Text>
        <View style={{ display: 'flex', flexDirection: 'row', gap: 10, marginTop: 12 }}>
          <View onClick={goActivities} style={{ flex: 1, padding: 12, borderRadius: 10, backgroundColor: '#f5f6f8' }}>
            <Text style={{ fontWeight: 'bold', display: 'block' }}>活动</Text>
            <Text style={{ color: '#666', fontSize: 12 }}>报名/活动详情</Text>
          </View>
          <View style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 10 }}>
            <View onClick={goShareCenter} style={{ padding: 12, borderRadius: 10, backgroundColor: '#f5f6f8' }}>
              <Text style={{ fontWeight: 'bold', display: 'block' }}>邀请好友</Text>
              <Text style={{ color: '#666', fontSize: 12 }}>得现金奖励</Text>
            </View>
            <View onClick={goCoupons} style={{ padding: 12, borderRadius: 10, backgroundColor: '#f5f6f8' }}>
              <Text style={{ fontWeight: 'bold', display: 'block' }}>优惠券</Text>
              <Text style={{ color: '#666', fontSize: 12 }}>平台/门店权益</Text>
            </View>
          </View>
        </View>
      </View>

      {/* 首页-商城：分区 */}
      <View style={{ backgroundColor: '#fff', borderRadius: 12, padding: 12, marginBottom: 12 }}>
        <Text style={{ fontSize: 16, fontWeight: 'bold' }}>商城分区</Text>
        <View
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(2, 1fr)',
            marginTop: 12,
            borderTop: '1px solid #f0f0f0',
            borderLeft: '1px solid #f0f0f0',
            borderRadius: 10,
            overflow: 'hidden',
          }}
        >
          {mallPartitions.map((p) => (
            <View
              key={p.key}
              onClick={() => goPartition(p.categoryId)}
              style={{
                padding: 12,
                borderRight: '1px solid #f0f0f0',
                borderBottom: '1px solid #f0f0f0',
              }}
            >
              <Text style={{ fontWeight: 'bold' }}>{p.title}</Text>
              <Text style={{ display: 'block', color: '#999', fontSize: 12 }}>点击进入</Text>
            </View>
          ))}
        </View>
      </View>

      {/* 推荐商品 */}
      <View style={{ backgroundColor: '#fff', borderRadius: 12, padding: 12, marginBottom: 12 }}>
        <Text style={{ fontSize: 16, fontWeight: 'bold' }}>推荐商品</Text>
        {recommended.length === 0 ? (
          <Text style={{ color: '#999' }}>暂无推荐商品</Text>
        ) : (
          <View className="grid" style={{ marginTop: 8 }}>
            {recommended.map((p) => (
              <ProductCard key={p.id} product={p} showCover onClick={() => goProductDetail(p.id, currentStoreId)} />
            ))}
          </View>
        )}
      </View>

      {/* 茶心阁简介区 */}
      <View style={{ backgroundColor: '#fff', borderRadius: 12, padding: 12, marginBottom: 12 }}>
        <View style={{ display: 'flex', flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
          <Text style={{ fontSize: 16, fontWeight: 'bold' }}>茶心阁简介</Text>
          <Button size="mini" onClick={() => Taro.navigateTo({ url: '/pages/about/index' }).catch(() => {})}>更多</Button>
        </View>
        <View style={{ marginTop: 8 }}>
          <ContentRenderer markdown={aboutMarkdown} />
        </View>
      </View>

      {/* 本地商品（当前门店） */}
      <View style={{ backgroundColor: '#fff', borderRadius: 12, padding: 12, marginBottom: 12 }}>
        <Text style={{ fontSize: 16, fontWeight: 'bold' }}>本地商品</Text>
        {localProducts.length === 0 ? (
          <Text style={{ color: '#999' }}>暂无本地商品</Text>
        ) : (
          <View className="grid" style={{ marginTop: 8 }}>
            {localProducts.map((p) => (
              <ProductCard key={p.id} product={p} showCover onClick={() => goProductDetail(p.id, currentStoreId)} />
            ))}
          </View>
        )}
      </View>

      {/* 技术支持&版权声明 */}
      <View style={{ marginTop: 4, padding: 12, backgroundColor: '#fff', borderRadius: 12 }}>
        <Text style={{ color: '#666' }}>技术支持：茶心阁</Text>
        {sitePhone && <Text style={{ color: '#666' }}> 客服电话：{sitePhone}</Text>}
        <View style={{ marginTop: 6 }}>
          <Text style={{ color: '#999' }}>{siteCopyright || 'Copyright © 茶心阁'}</Text>
        </View>
      </View>

      {loading && <Text style={{ color: '#999' }}>加载中...</Text>}
    </View>
  );
}
