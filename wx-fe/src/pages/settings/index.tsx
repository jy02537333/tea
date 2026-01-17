import React from 'react';
import { View, Text, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';

export default function SettingsPage() {
  function gotoLogin() {
    Taro.navigateTo({ url: '/pages/login/index' }).catch(() => {});
  }
  function gotoAddress() {
    Taro.navigateTo({ url: '/pages/address/index' }).catch(() => {});
  }
  function gotoAbout() {
    Taro.navigateTo({ url: '/pages/about/index' }).catch(() => {});
  }
  function gotoPrivacy() {
    Taro.navigateTo({ url: '/pages/privacy/index' }).catch(() => {});
  }
  function gotoTerms() {
    Taro.navigateTo({ url: '/pages/terms/index' }).catch(() => {});
  }
  function gotoFeedback() {
    Taro.navigateTo({ url: '/pages/feedback/index' }).catch(() => {});
  }

  function clearCache() {
    try {
      Taro.clearStorageSync();
      Taro.showToast({ title: '已清除缓存', icon: 'none' });
    } catch (_) {}
  }

  return (
    <View style={{ padding: 16 }}>
      <View style={{ background: '#fff', borderRadius: 12, padding: 16 }}>
        <Text style={{ fontSize: 16, fontWeight: 'bold' }}>账号与安全</Text>
        <Button style={{ marginTop: 12 }} size="mini" onClick={gotoLogin}>
          账号与登录
        </Button>
      </View>

      <View style={{ background: '#fff', borderRadius: 12, padding: 16, marginTop: 16 }}>
        <Text style={{ fontSize: 16, fontWeight: 'bold' }}>收货与地址</Text>
        <Button style={{ marginTop: 12 }} size="mini" onClick={gotoAddress}>
          管理收货地址
        </Button>
      </View>

      <View style={{ background: '#fff', borderRadius: 12, padding: 16, marginTop: 16 }}>
        <Text style={{ fontSize: 16, fontWeight: 'bold' }}>关于与协议</Text>
        <Button style={{ marginTop: 12 }} size="mini" onClick={gotoAbout}>
          关于我们
        </Button>
        <Button style={{ marginTop: 12 }} size="mini" onClick={gotoPrivacy}>
          隐私协议
        </Button>
        <Button style={{ marginTop: 12 }} size="mini" onClick={gotoTerms}>
          用户协议
        </Button>
      </View>

      <View style={{ background: '#fff', borderRadius: 12, padding: 16, marginTop: 16 }}>
        <Text style={{ fontSize: 16, fontWeight: 'bold' }}>其他</Text>
        <Button style={{ marginTop: 12 }} size="mini" onClick={gotoFeedback}>
          意见反馈
        </Button>
        <Button style={{ marginTop: 12 }} size="mini" onClick={clearCache}>
          清除缓存
        </Button>
      </View>
    </View>
  );
}
