import React, { useEffect, useMemo, useState } from 'react';
import { View, Text, Button, Canvas, Image as TaroImage, Picker } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { getMeSummary } from '../../services/me';
import type { MeSummary } from '../../services/types';
import { buildShareLink, getWxaCode, listSharePosterTemplates } from '../../services/share';
import { getStore } from '../../services/stores';
// 动态引入 QRCode 以减少首屏体积
import { DEFAULT_SHARE_TEMPLATE } from '../../constants/share';

export default function SharePage() {
  const [summary, setSummary] = useState<MeSummary | null>(null);
  const [loading, setLoading] = useState(false);
  const [posterPath, setPosterPath] = useState<string | null>(null);
  const [useBgImage, setUseBgImage] = useState<boolean>(false);
  const [posterTemplates, setPosterTemplates] = useState<Array<{ id: string; title?: string; image_url: string }>>([]);
  const [posterTemplateIndex, setPosterTemplateIndex] = useState<number>(0); // 0 = pure

  useEffect(() => {
    try {
      const saved = Taro.getStorageSync('share_useBgImage');
      if (saved === '1' || saved === true) {
        setUseBgImage(true);
      } else if (saved === '0' || saved === false) {
        setUseBgImage(false);
      } else {
        setUseBgImage(DEFAULT_SHARE_TEMPLATE === 'bg');
      }
    } catch (_) {
      // ignore storage read errors
    }
  }, []);

  useEffect(() => {
    void load();
  }, []);

  async function load() {
    setLoading(true);
    try {
      const s = await getMeSummary();
      setSummary(s);

    // 加载分享海报模板（后台可配置，多张可切换）
    try {
      const list = await listSharePosterTemplates();
      const cleaned = (list || [])
        .filter((it) => it && typeof it.image_url === 'string' && it.image_url.trim())
        .map((it) => ({ id: String(it.id || ''), title: it.title, image_url: String(it.image_url || '').trim() }))
        .filter((it) => it.id && it.image_url);
      setPosterTemplates(cleaned);
      // restore selection
      try {
        const savedId = String(Taro.getStorageSync('share_poster_template_id') || '').trim();
        if (savedId) {
          const idx = cleaned.findIndex((x) => x.id === savedId);
          if (idx >= 0) setPosterTemplateIndex(idx + 1);
        }
      } catch (_) {}
    } catch (_) {
      // ignore template load errors
    }
    } catch (e: any) {
      Taro.showToast({ title: e?.message || '加载分享数据失败', icon: 'none' });
    } finally {
      setLoading(false);
    }
  }

  const shareStats = summary?.share;

  const posterTemplateOptions = useMemo(() => {
    if (!posterTemplates.length) return [] as string[];
    const titles = posterTemplates.map((t, i) => (t.title && String(t.title).trim()) || `模板${i + 1}`);
    return ['纯色模板', ...titles];
  }, [posterTemplates]);

  async function handleGenerateShare() {
    try {
      const myId = summary?.user?.id;
      if (!myId) {
        Taro.showToast({ title: '未获取到用户信息', icon: 'none' });
        return;
      }
      const sidStr = Taro.getStorageSync('current_store_id');
      const sid = sidStr ? Number(sidStr) : undefined;
      const storeId = sid && !Number.isNaN(sid) ? sid : undefined;
      const link = buildShareLink({ referrerId: myId, storeId });

      // 生成小程序码（后端接口），并在 H5 上进行画布合成海报
      const scene = `referrer_id=${myId}` + (storeId ? `&store_id=${storeId}` : '');
      const codeDataUrl = await getWxaCode({ scene, page: 'pages/index/index', width: 240, is_hyaline: true });

      let storeInfo: { name?: string; phone?: string; address?: string } = {};
      if (storeId) {
        try {
          const s = await getStore(storeId);
          storeInfo = { name: s?.name, phone: s?.phone, address: s?.address };
        } catch (_) {}
      }

      const env = (Taro as any)?.getEnv ? (Taro as any).getEnv() : 'WEB';
      let bgDataUrl: string | undefined;
      if (posterTemplates.length > 0) {
        if (posterTemplateIndex > 0) {
          bgDataUrl = posterTemplates[posterTemplateIndex - 1]?.image_url;
        } else {
          bgDataUrl = undefined;
        }
      } else {
        bgDataUrl = useBgImage ? generateDefaultBgDataUrl() : undefined;
      }
      if (env === 'WEB' && typeof document !== 'undefined') {
        const url = await composePosterH5({
          avatar: summary?.user?.avatar,
          nickname: summary?.user?.nickname || '茶友',
          store: storeInfo,
          codeDataUrl:
            codeDataUrl ||
            (await (await import('qrcode')).default.toDataURL(link, { width: 240, margin: 1 })),
          bgDataUrl,
        });
        setPosterPath(url);
      } else {
        // 小程序环境：使用 CanvasContext 合成并导出临时文件
        const temp = await composePosterWeapp('poster-canvas', {
          avatar: summary?.user?.avatar,
          nickname: summary?.user?.nickname || '茶友',
          store: storeInfo,
          codeDataUrl:
            codeDataUrl ||
            (await (await import('qrcode')).default.toDataURL(link, { width: 240, margin: 1 })),
          bgDataUrl,
        });
        setPosterPath(temp);
      }
      Taro.showToast({ title: '已生成分享海报（点击保存）', icon: 'none' });
    } catch (e: any) {
      Taro.showToast({ title: e?.message || '生成分享失败', icon: 'none' });
    }
  }

  async function composePosterH5(params: { avatar?: string; nickname: string; store: { name?: string; phone?: string; address?: string }; codeDataUrl: string; bgDataUrl?: string }): Promise<string> {
    const canvas = document.createElement('canvas');
    const width = 750; // 750x1200 适配高清手机屏（按像素）
    const height = 1200;
    canvas.width = width;
    canvas.height = height;
    const ctx = canvas.getContext('2d')!;

    // 背景图或渐变
    if (params.bgDataUrl) {
      try {
        const bg = await loadImage(params.bgDataUrl);
        ctx.drawImage(bg, 0, 0, width, height);
      } catch {
        const grad = ctx.createLinearGradient(0, 0, 0, height);
        grad.addColorStop(0, '#fdf6e3');
        grad.addColorStop(1, '#f5deb3');
        ctx.fillStyle = grad;
        ctx.fillRect(0, 0, width, height);
      }
    } else {
      const grad = ctx.createLinearGradient(0, 0, 0, height);
      grad.addColorStop(0, '#fdf6e3');
      grad.addColorStop(1, '#f5deb3');
      ctx.fillStyle = grad;
      ctx.fillRect(0, 0, width, height);
    }

    // 标题
    ctx.fillStyle = '#333';
    ctx.font = 'bold 36px sans-serif';
    ctx.fillText('茶心阁 · 分享推广', 30, 60);

    // 头像圆形裁剪
    if (params.avatar) {
      const img = await loadImage(params.avatar);
      const r = 80;
      ctx.save();
      ctx.beginPath();
      ctx.arc(70 + r, 120, r, 0, Math.PI * 2);
      ctx.clip();
      ctx.drawImage(img, 70, 40, r * 2, r * 2);
      ctx.restore();
    }

    // 昵称
    ctx.fillStyle = '#222';
    ctx.font = '28px sans-serif';
    ctx.fillText(`${params.nickname} 的专属分享`, 30, 200);

    // 门店信息
    const storeY = 250;
    ctx.fillStyle = '#555';
    ctx.font = '24px sans-serif';
    if (params.store?.name) ctx.fillText(`门店：${params.store.name}`, 30, storeY);
    if (params.store?.phone) ctx.fillText(`电话：${params.store.phone}`, 30, storeY + 36);
    if (params.store?.address) {
      ctx.fillText(`地址：${params.store.address}`, 30, storeY + 72);
    }

    // 小程序码
    const codeImg = await loadImage(params.codeDataUrl);
    const codeSize = 380;
    const codeX = (width - codeSize) / 2;
    const codeY = 420;
    ctx.drawImage(codeImg, codeX, codeY, codeSize, codeSize);

    // 引导文案
    ctx.fillStyle = '#444';
    ctx.font = '26px sans-serif';
    ctx.fillText('长按识别小程序码进入商城/门店', (width - 480) / 2, codeY + codeSize + 50);

    return canvas.toDataURL('image/png');
  }

  function loadImage(src: string): Promise<HTMLImageElement> {
    return new Promise((resolve, reject) => {
      const img = new (window as any).Image();
      img.crossOrigin = 'anonymous';
      img.onload = () => resolve(img);
      img.onerror = reject;
      img.src = src;
    });
  }

  async function composePosterWeapp(
    canvasId: string,
    params: { avatar?: string; nickname: string; store: { name?: string; phone?: string; address?: string }; codeDataUrl: string; bgDataUrl?: string },
  ): Promise<string> {
    const width = 750;
    const height = 1200;
    const ctx = Taro.createCanvasContext(canvasId);

    // 背景图或渐变
    if (params.bgDataUrl) {
      try {
        let bgPath = '';
        if (params.bgDataUrl.startsWith('data:image')) {
          bgPath = await dataUrlToTempFile(params.bgDataUrl, `bg_${Date.now()}.png`);
        } else {
          const info = await Taro.getImageInfo({ src: params.bgDataUrl });
          bgPath = info?.path || '';
        }
        if (bgPath) {
          ctx.drawImage(bgPath, 0, 0, width, height);
        } else {
          throw new Error('empty bg path');
        }
      } catch (_) {
        const grad = ctx.createLinearGradient(0, 0, 0, height);
        grad.addColorStop(0, '#fdf6e3');
        grad.addColorStop(1, '#f5deb3');
        // @ts-ignore
        ctx.setFillStyle(grad as any);
        ctx.fillRect(0, 0, width, height);
      }
    } else {
      const grad = ctx.createLinearGradient(0, 0, 0, height);
      grad.addColorStop(0, '#fdf6e3');
      grad.addColorStop(1, '#f5deb3');
      // @ts-ignore
      ctx.setFillStyle(grad as any);
      ctx.fillRect(0, 0, width, height);
    }

    // 标题
    ctx.setFillStyle('#333');
    ctx.setFontSize(36);
    ctx.fillText('茶心阁 · 分享推广', 30, 60);

    // 头像
    if (params.avatar) {
      try {
        const info = await Taro.getImageInfo({ src: params.avatar });
        const r = 80;
        ctx.save();
        ctx.beginPath();
        ctx.arc(70 + r, 120, r, 0, Math.PI * 2);
        ctx.clip();
        ctx.drawImage(info.path, 70, 40, r * 2, r * 2);
        ctx.restore();
      } catch (_) {}
    }

    // 昵称
    ctx.setFillStyle('#222');
    ctx.setFontSize(28);
    ctx.fillText(`${params.nickname} 的专属分享`, 30, 200);

    // 门店信息
    const storeY = 250;
    ctx.setFillStyle('#555');
    ctx.setFontSize(24);
    if (params.store?.name) ctx.fillText(`门店：${params.store.name}`, 30, storeY);
    if (params.store?.phone) ctx.fillText(`电话：${params.store.phone}`, 30, storeY + 36);
    if (params.store?.address) ctx.fillText(`地址：${params.store.address}`, 30, storeY + 72);

    // 小程序码：将 base64 写入本地临时文件再 drawImage
    let codePath = '';
    try {
      codePath = await dataUrlToTempFile(params.codeDataUrl, `wxa_${Date.now()}.png`);
    } catch (_) {}
    if (codePath) {
      const codeSize = 380;
      const codeX = (width - codeSize) / 2;
      const codeY = 420;
      ctx.drawImage(codePath, codeX, codeY, codeSize, codeSize);
    }

    // 引导文案
    ctx.setFillStyle('#444');
    ctx.setFontSize(26);
    ctx.fillText('长按识别小程序码进入商城/门店', 110, 420 + 380 + 50);

    return new Promise<string>((resolve, reject) => {
      ctx.draw(false, () => {
        (Taro as any).canvasToTempFilePath(
          {
            canvasId,
            destWidth: width,
            destHeight: height,
            success: (res: any) => resolve(res.tempFilePath),
            fail: (err: any) => reject(err),
          },
          // @ts-ignore - in component scope
          this,
        );
      });
    });
  }

  async function dataUrlToTempFile(dataUrl: string, fileName: string): Promise<string> {
    const wxAny = (globalThis as any).wx;
    const base64 = dataUrl.replace(/^data:image\/[a-zA-Z]+;base64,/, '');
    const fsm = Taro.getFileSystemManager();
    const userPath = (wxAny && wxAny.env && wxAny.env.USER_DATA_PATH) || `${Date.now()}`;
    const filePath = `${userPath}/${fileName}`;
    await new Promise<void>((resolve, reject) => {
      try {
        fsm.writeFile({ filePath, data: base64, encoding: 'base64', success: () => resolve(), fail: reject });
      } catch (e) {
        reject(e);
      }
    });
    return filePath;
  }

  function generateDefaultBgDataUrl(): string {
    // 使用 SVG 生成高斯渐变背景作为“背景素材图”
    const svg = `
      <svg xmlns='http://www.w3.org/2000/svg' width='750' height='1200'>
        <defs>
          <linearGradient id='g' x1='0' y1='0' x2='0' y2='1'>
            <stop offset='0%' stop-color='#fdf6e3'/>
            <stop offset='100%' stop-color='#f5deb3'/>
          </linearGradient>
          <radialGradient id='light' cx='0.2' cy='0.1' r='0.6'>
            <stop offset='0%' stop-color='rgba(255,255,255,0.8)'/>
            <stop offset='100%' stop-color='rgba(255,255,255,0)'/>
          </radialGradient>
        </defs>
        <rect width='100%' height='100%' fill='url(#g)'/>
        <circle cx='120' cy='120' r='200' fill='url(#light)'/>
      </svg>`;
    const enc = encodeURIComponent(svg)
      .replace(/'/g, '%27')
      .replace(/\(/g, '%28')
      .replace(/\)/g, '%29');
    return `data:image/svg+xml;charset=UTF-8,${enc}`;
  }

  async function handleSavePoster() {
    if (!posterPath) {
      Taro.showToast({ title: '请先生成海报', icon: 'none' });
      return;
    }
    try {
      // H5 环境：直接打开图片即可；小程序环境尝试保存到相册
      if ((Taro as any).saveImageToPhotosAlbum) {
        const temp = posterPath;
        await (Taro as any).saveImageToPhotosAlbum({ filePath: temp });
        Taro.showToast({ title: '已保存到相册', icon: 'success' });
      } else {
        Taro.previewImage({ urls: [posterPath] });
      }
    } catch (_) {
      Taro.previewImage({ urls: [posterPath] });
    }
  }

  return (
    <View data-testid="page-share" style={{ padding: 16 }}>
      <Text style={{ fontSize: 18, fontWeight: 'bold' }}>分享推广</Text>
      {loading && <Text style={{ display: 'block', marginTop: 8 }}>加载中...</Text>}
      {!loading && (
        <View style={{ marginTop: 12 }}>
          <Text style={{ display: 'block' }}>直推人数：{shareStats?.direct_count ?? 0}</Text>
          <Text style={{ display: 'block', marginTop: 4 }}>团队人数：{shareStats?.team_count ?? 0}</Text>
          <Text style={{ display: 'block', marginTop: 4 }}>累计佣金（分）：{shareStats?.total_commission_cents ?? 0}</Text>
          <Text style={{ display: 'block', marginTop: 4 }}>可提现佣金（分）：{shareStats?.available_commission_cents ?? 0}</Text>
          <Text style={{ display: 'block', marginTop: 4 }}>冻结佣金（分）：{shareStats?.frozen_commission_cents ?? 0}</Text>

          <Button style={{ marginTop: 12 }} type="primary" size="mini" onClick={handleGenerateShare}>
            生成分享链接/海报
          </Button>
          {posterTemplates.length > 0 && posterTemplateOptions.length > 0 && (
            <Picker
              mode="selector"
              range={posterTemplateOptions}
              value={posterTemplateIndex}
              onChange={(e) => {
                const v = Number((e as any)?.detail?.value || 0);
                const next = Number.isFinite(v) ? v : 0;
                setPosterTemplateIndex(next);
                try {
                  const id = next > 0 ? posterTemplates[next - 1]?.id : '';
                  if (id) Taro.setStorageSync('share_poster_template_id', id);
                  else Taro.removeStorageSync('share_poster_template_id');
                } catch (_) {}
              }}
            >
              <Button style={{ marginTop: 12, marginLeft: 8 }} size="mini">
                模板：{posterTemplateOptions[posterTemplateIndex] || '纯色模板'}
              </Button>
            </Picker>
          )}
          {posterTemplates.length === 0 && (
            <Button
              style={{ marginTop: 12, marginLeft: 8 }}
              size="mini"
              onClick={() => {
                const next = !useBgImage;
                setUseBgImage(next);
                try {
                  Taro.setStorageSync('share_useBgImage', next ? '1' : '0');
                } catch (_) {
                  // ignore storage write errors
                }
              }}
            >
              {useBgImage ? '使用纯色模板' : '使用背景图模板'}
            </Button>
          )}
          {posterPath && (
            <View style={{ marginTop: 12 }}>
              <Text style={{ display: 'block', marginBottom: 8 }}>预览：</Text>
              <TaroImage style={{ width: '240px', height: '240px' }} src={posterPath} mode="aspectFit" />
              <Button style={{ marginTop: 8 }} size="mini" onClick={handleSavePoster}>保存海报</Button>
            </View>
          )}
          {!posterPath && (
            <Text style={{ display: 'block', marginTop: 12, color: '#999' }}>分享物料与二维码将在后续版本完善</Text>
          )}
        </View>
      )}
      {/* 隐藏画布用于小程序端海报合成 */}
      <Canvas canvasId="poster-canvas" style={{ position: 'absolute', left: -9999, top: -9999, width: '1px', height: '1px' }} />
    </View>
  );
}
