export type StoreStatus = 'open' | 'closed' | 'busy';

export type StoreSupport = {
  pickup: boolean;
  takeout: boolean;
  member: boolean;
};

export type StoreFeeRule = {
  serviceFee: number;
  couponThreshold: number;
  couponDiscount: number;
  couponLabel: string;
};

export type StoreService = {
  label: string;
  type: 'tea' | 'space' | 'facility' | 'office';
};

export type StoreItem = {
  id: string;
  name: string;
  meta: string;
  distance: string;
  actionNote: string;
  status: StoreStatus;
  supports: StoreSupport;
  feeRule: StoreFeeRule;
  businessHours: string;
  holidayNote: string;
  address: string;
  addressHint?: string;
  phone: string;
  mapTitle: string;
  mapSub: string;
  services: StoreService[];
  notices: string[];
};

export type ProductCategory = {
  id: string;
  name: string;
};

export type ProductItem = {
  id: string;
  name: string;
  desc: string;
  tag: string;
  price: number;
  categoryId: string;
  useQtyControl?: boolean;
};

export type RecommendItem = {
  id: string;
  name: string;
  price: number;
};

export type PlatformSku = {
  id: string;
  name: string;
  price: number;
  points: number;
  status: 'available' | 'soldout' | 'presale';
  stock?: number;
};

export type CartItem = {
  id: string;
  name: string;
  spec: string;
  price: number;
  quantity: number;
};

export type CartSummary = {
  storeId: string;
  storeName: string;
  storeAddress: string;
  storePhone: string;
  pickupType: 'self' | 'dinein';
  items: CartItem[];
  serviceFee: number;
  couponLabel?: string;
  couponDiscount?: number;
};

export type CartGroup = {
  storeId: string;
  storeName: string;
  items: CartItem[];
};

export type OrderStatus = 'pending' | 'making' | 'done' | 'canceled';

export type OrderItem = {
  id: string;
  name: string;
  quantity: number;
};

export type OrderSummary = {
  id: string;
  storeId: string;
  storeName: string;
  status: OrderStatus;
  time: string;
  pickupType: string;
  items: OrderItem[];
  totalLabel: string;
  totalValue?: number;
  actions: string[];
};

export type PlatformProductDetail = {
  id: string;
  categoryId: string;
  name: string;
  heroTag: string;
  heroBadge: string;
  tags: string[];
  desc: string;
  netWeight: string;
  shelfLife: string;
  shippingTips: string[];
  brewTips: string[];
  afterSales: string[];
  skus: PlatformSku[];
  defaultSkuId: string;
};

export const mockStores: StoreItem[] = [
  {
    id: 'store-1',
    name: '隽也YUYE茶馆 · 本店',
    meta: '茶心君坐店 · 今日营业 10:00-22:00',
    distance: '距离您约 1671.6km',
    actionNote: '支持自提 / 外卖',
    status: 'open',
    supports: { pickup: true, takeout: true, member: true },
    feeRule: { serviceFee: 3, couponThreshold: 50, couponDiscount: 5, couponLabel: '满 50 减 5' },
    businessHours: '周一至周日 10:00 - 22:00',
    holidayNote: '节假日营业时间以门店公告为准',
    address: '浙江省杭州市西湖区 · 某某路 88 号 · 近地铁 XX 站',
    addressHint: '周边公共停车位有限，建议绿色出行',
    phone: '400-123-4567',
    mapTitle: '浙江省 · 杭州市 · 西湖区',
    mapSub: '详细导航以实际小程序/地图为准',
    services: [
      { label: '单人茶席', type: 'tea' },
      { label: '双人/多人拼桌', type: 'tea' },
      { label: '预留包间（需预约）', type: 'space' },
      { label: '免费 Wi-Fi', type: 'facility' },
      { label: '安静阅读区', type: 'space' },
      { label: '可带电脑办公', type: 'office' }
    ],
    notices: [
      '本店一客一茶，每位客人皆有独立茶席安排。',
      '如遇高峰时段，可能适当安排拼桌，还请理解。',
      '自带茶叶/茶器，欢迎与茶心君提前沟通存放及冲泡方式。'
    ]
  },
  {
    id: 'store-2',
    name: '隽也YUYE茶馆 · 城南店',
    meta: '靠近城南商圈 · 适合朋友小聚',
    distance: '距离您约 5.2km',
    actionNote: '暂不支持平台年卡权益',
    status: 'busy',
    supports: { pickup: true, takeout: true, member: false },
    feeRule: { serviceFee: 4, couponThreshold: 60, couponDiscount: 6, couponLabel: '满 60 减 6' },
    businessHours: '周一至周日 11:00 - 22:30',
    holidayNote: '周末晚间高峰需提前预约',
    address: '杭州市滨江区 · 江南大道 1699 号',
    addressHint: '商场停车 2 小时免费',
    phone: '0571-8888-6699',
    mapTitle: '浙江省 · 杭州市 · 滨江区',
    mapSub: '高峰期请优先电话预约',
    services: [
      { label: '朋友小聚', type: 'space' },
      { label: '临窗座位', type: 'space' },
      { label: '免费 Wi-Fi', type: 'facility' },
      { label: '可外带', type: 'facility' }
    ],
    notices: [
      '晚间 19:00-21:00 预计满座，建议错峰到店。',
      '城南店暂不支持平台年卡权益。'
    ]
  },
  {
    id: 'store-3',
    name: '隽也YUYE茶馆 · 城北店',
    meta: '安静适合阅读 · 座位较紧俏',
    distance: '距离您约 8.7km',
    actionNote: '预计满座时段 19:00-21:00',
    status: 'closed',
    supports: { pickup: false, takeout: false, member: true },
    feeRule: { serviceFee: 2, couponThreshold: 40, couponDiscount: 4, couponLabel: '满 40 减 4' },
    businessHours: '周一至周日 12:00 - 20:30',
    holidayNote: '今日休息，营业时间以公告为准',
    address: '杭州市拱墅区 · 湖墅南路 88 号',
    addressHint: '步行 5 分钟可达地铁站',
    phone: '0571-6666-1122',
    mapTitle: '浙江省 · 杭州市 · 拱墅区',
    mapSub: '今日休息，建议关注公告',
    services: [
      { label: '安静阅读区', type: 'space' },
      { label: '单人茶席', type: 'tea' },
      { label: '可带电脑办公', type: 'office' }
    ],
    notices: [
      '城北店今日休息，营业时间以公告为准。',
      '支持平台年卡权益。'
    ]
  }
];

export const mockMenuCategories: ProductCategory[] = [
  { id: 'membership', name: '会员年卡' },
  { id: 'tea-wine', name: '茶·酒' },
  { id: 'snack', name: '轻食' },
  { id: 'hot-tea', name: '热泡茶' },
  { id: 'aged-tea', name: '煮老茶' },
  { id: 'triple-tea', name: '一茶三喝' }
];

export const mockMenuRecommendations: RecommendItem[] = [
  { id: 'rec-1', name: '岚山绿茶·特集', price: 98 },
  { id: 'rec-2', name: '明前西湖龙井', price: 68 },
  { id: 'rec-3', name: '当真是梨花', price: 88 }
];

export const mockMenuItems: ProductItem[] = [
  {
    id: 'menu-1',
    name: '会员年卡 · 限时特惠',
    desc: '茶也小程序-右下角-我的-年卡会员',
    tag: '会员年卡',
    price: 9999,
    categoryId: 'membership'
  },
  {
    id: 'menu-2',
    name: '小事化甜 · gelato版',
    desc: '茶·gelato版',
    tag: '茶·酒',
    price: 58,
    categoryId: 'tea-wine',
    useQtyControl: true
  },
  {
    id: 'menu-3',
    name: '挺酸',
    desc: '威士忌 · 正山小种 · 杨梅',
    tag: '茶·酒',
    price: 58,
    categoryId: 'tea-wine'
  },
  {
    id: 'menu-4',
    name: '云也小点心',
    desc: '轻食 · 下午茶搭配',
    tag: '轻食',
    price: 30,
    categoryId: 'snack'
  },
  {
    id: 'menu-5',
    name: '白桃乌龙 · 热泡',
    desc: '热泡 400ml ｜ 适合慢饮',
    tag: '热泡茶',
    price: 28,
    categoryId: 'hot-tea'
  },
  {
    id: 'menu-6',
    name: '陈年普洱 · 老茶',
    desc: '炭火慢煮 · 茶汤醇厚',
    tag: '煮老茶',
    price: 88,
    categoryId: 'aged-tea'
  },
  {
    id: 'menu-7',
    name: '一茶三喝 · 翠玉白茶',
    desc: '三段式风味体验',
    tag: '一茶三喝',
    price: 68,
    categoryId: 'triple-tea'
  }
];

export const mockProductCategories: ProductCategory[] = [
  { id: 'hot', name: '热销' },
  { id: 'drink', name: '饮品' },
  { id: 'tea', name: '茶叶' },
  { id: 'ware', name: '茶器' }
];

export const mockProductItems: ProductItem[] = [
  {
    id: 'prod-1',
    name: '桂花乌龙 · 冷萃',
    desc: '冷萃 500ml ｜ 桂花香气清甜，入口顺滑',
    tag: '热销 · 饮品',
    price: 26,
    categoryId: 'hot'
  },
  {
    id: 'prod-2',
    name: '白桃乌龙 · 热泡',
    desc: '热泡 400ml ｜ 果香饱满，适合慢饮',
    tag: '饮品',
    price: 28,
    categoryId: 'drink'
  },
  {
    id: 'prod-3',
    name: '小事化甜 · gelato版',
    desc: '茶·gelato版 ｜ 适合今日放空',
    tag: '饮品 · 特调',
    price: 58,
    categoryId: 'drink'
  },
  {
    id: 'prod-4',
    name: '岩茶 · 水仙 50g',
    desc: '岩韵显，适合回甘爱好者',
    tag: '茶叶',
    price: 88,
    categoryId: 'tea'
  },
  {
    id: 'prod-5',
    name: '桂花乌龙 50g 礼盒',
    desc: '礼盒装 ｜ 适合作为小心意送礼',
    tag: '茶叶 · 礼盒',
    price: 108,
    categoryId: 'tea'
  },
  {
    id: 'prod-6',
    name: '个人茶席茶具套装',
    desc: '盖碗 / 公道杯 / 品茗杯 4 只',
    tag: '茶器 · 套装',
    price: 268,
    categoryId: 'ware'
  },
  {
    id: 'prod-7',
    name: '云也 · 茶杯单只',
    desc: '可与现有茶席搭配使用',
    tag: '茶器',
    price: 58,
    categoryId: 'ware'
  }
];

export const mockPlatformProducts: PlatformProductDetail[] = [
  {
    id: 'platform-1',
    categoryId: 'gift',
    name: '桂花乌龙茶礼盒',
    heroTag: '精选 · 茶叶礼盒',
    heroBadge: '平台直发 ｜ 全国包邮',
    tags: ['桂花乌龙', '礼盒装', '赠礼优选'],
    desc: '以当季桂花与优质乌龙茶为主角，低温慢焙锁住香气，一盒包含 2 罐桂花乌龙茶与精致礼袋，既适合自饮，也适合赠与重要的 TA。',
    netWeight: '100g × 2 罐',
    shelfLife: '18 个月，见包装喷码',
    shippingTips: [
      '由「茶心阁官方旗舰」统一发货，全国大部分地区包邮。',
      '下单后 24 小时内安排出库，节假日可能略有延迟。',
      '支持在「订单详情」中查看物流进度。'
    ],
    brewTips: [
      '建议水温 90℃ 左右，先温杯，再投茶。',
      '茶水比例约 1:20，可根据个人口味微调浓淡。',
      '首泡洗茶 1 次，之后每泡浸泡 8-12 秒为宜。'
    ],
    afterSales: [
      '如遇破损、错发、漏发，请在签收 24 小时内通过小程序「订单售后」联系茶心君。',
      '因个人口味原因暂不支持无理由退换货，敬请理解。'
    ],
    defaultSkuId: 'sku-1',
    skus: [
      { id: 'sku-1', name: '礼盒装 · 100g × 2', price: 198, points: 198, status: 'available', stock: 26 },
      { id: 'sku-2', name: '礼盒装 · 200g × 2', price: 328, points: 328, status: 'presale', stock: 0 },
      { id: 'sku-3', name: '单罐装 · 100g', price: 118, points: 118, status: 'soldout', stock: 0 }
    ]
  },
  {
    id: 'platform-2',
    categoryId: 'ware',
    name: '云也 · 茶席周边套装',
    heroTag: '茶器周边',
    heroBadge: '平台直发 ｜ 限量发售',
    tags: ['茶器周边', '桌面美学', '套装'],
    desc: '杯垫、茶巾、小摆件一套装配，适合茶席与居家桌面布置，提升茶桌氛围感。',
    netWeight: '1 套',
    shelfLife: '不适用',
    shippingTips: [
      '平台统一发货，部分地区 3-5 天送达。',
      '易碎品请当场验收。'
    ],
    brewTips: [
      '适配日常茶席使用，可与现有茶具搭配。'
    ],
    afterSales: [
      '签收 24 小时内可申请售后。'
    ],
    defaultSkuId: 'sku-4',
    skus: [
      { id: 'sku-4', name: '周边套装 · 基础款', price: 268, points: 268, status: 'available', stock: 12 },
      { id: 'sku-5', name: '周边套装 · 礼盒款', price: 328, points: 328, status: 'available', stock: 6 }
    ]
  },
  {
    id: 'platform-3',
    categoryId: 'member',
    name: '云也会员年卡',
    heroTag: '会员权益',
    heroBadge: '平台权益 ｜ 专属折扣',
    tags: ['会员年卡', '专属折扣', '积分加成'],
    desc: '开通会员年卡后可享平台商品专属折扣、积分加成与会员限定礼遇。',
    netWeight: '权益卡 × 1',
    shelfLife: '开通后 12 个月有效',
    shippingTips: [
      '权益开通后即时生效，无需物流。'
    ],
    brewTips: [
      '会员权益可在「我的 - 会员中心」查看。'
    ],
    afterSales: [
      '虚拟商品不支持退款，敬请理解。'
    ],
    defaultSkuId: 'sku-6',
    skus: [
      { id: 'sku-6', name: '会员年卡', price: 9999, points: 9999, status: 'available', stock: 999 }
    ]
  }
];

export const mockPlatformCategories: ProductCategory[] = [
  { id: 'featured', name: '精选' },
  { id: 'gift', name: '茶叶礼盒' },
  { id: 'ware', name: '茶器周边' },
  { id: 'member', name: '会员权益' }
];

export const mockPlatformItems: ProductItem[] = [
  {
    id: 'platform-1',
    name: '桂花乌龙茶礼盒',
    desc: '内含桂花乌龙茶 2 罐 · 礼袋一只',
    tag: '精选 · 茶叶礼盒',
    price: 198,
    categoryId: 'gift'
  },
  {
    id: 'platform-2',
    name: '云也 · 茶席周边套装',
    desc: '杯垫 / 茶巾 / 小摆件套组',
    tag: '茶器周边',
    price: 268,
    categoryId: 'ware'
  },
  {
    id: 'platform-3',
    name: '个人茶席茶具套装',
    desc: '盖碗 / 公道杯 / 品茗杯 4 只',
    tag: '茶器 · 套装',
    price: 268,
    categoryId: 'ware'
  },
  {
    id: 'platform-4',
    name: '云也会员年卡',
    desc: '平台权益 ｜ 下单享积分与专属折扣',
    tag: '会员权益',
    price: 9999,
    categoryId: 'member'
  }
];

export const mockCart: CartSummary = {
  storeId: 'store-1',
  storeName: '茶心阁 · 茶心门店（茶心君坐店）',
  storeAddress: '深圳市 · 南山区 · 某某路 88 号',
  storePhone: '185****0101',
  pickupType: 'self',
  items: [
    { id: 'menu-2', name: '桂花乌龙 · 冷萃', spec: '冷萃 500ml ｜ 微糖 ｜ 去冰', price: 26, quantity: 1 },
    { id: 'menu-5', name: '白桃乌龙 · 热泡', spec: '热泡 400ml ｜ 少糖 ｜ 常温', price: 28, quantity: 1 }
  ],
  serviceFee: 3,
  couponLabel: '满 50 减 5',
  couponDiscount: 5
};

export const mockOrders: OrderSummary[] = [
  {
    id: '20250118-001',
    storeId: 'store-1',
    storeName: '茶心阁 · 茶心门店',
    status: 'making',
    time: '今日 18:20',
    pickupType: '到店自取',
    items: [
      { id: 'menu-2', name: '桂花乌龙·冷萃', quantity: 1 },
      { id: 'menu-5', name: '白桃乌龙·热泡', quantity: 1 }
    ],
    totalLabel: '共 2 件 · 含茶席服务费',
    actions: ['订单详情', '催一催']
  },
  {
    id: '20250117-008',
    storeId: 'store-1',
    storeName: '茶心阁 · 茶心门店',
    status: 'done',
    time: '昨日 20:05',
    pickupType: '店内享用',
    items: [
      { id: 'menu-4', name: '岩茶·水仙', quantity: 1 },
      { id: 'menu-4', name: '云也小点心', quantity: 1 }
    ],
    totalLabel: '实付',
    totalValue: 118,
    actions: ['再次购买', '去评价']
  },
  {
    id: '20250115-003',
    storeId: 'store-2',
    storeName: '茶心阁 · 城南店',
    status: 'pending',
    time: '01-15 14:32',
    pickupType: '到店自取',
    items: [
      { id: 'menu-5', name: '白桃乌龙·冷萃', quantity: 1 }
    ],
    totalLabel: '需支付',
    totalValue: 26,
    actions: ['取消订单', '去支付']
  },
  {
    id: '20250110-021',
    storeId: 'store-3',
    storeName: '茶心阁 · 城北店',
    status: 'canceled',
    time: '01-10 12:18',
    pickupType: '到店自取',
    items: [
      { id: 'menu-7', name: '一茶三喝·翠玉白茶', quantity: 1 }
    ],
    totalLabel: '已取消',
    totalValue: 68,
    actions: ['再次购买']
  }
];

export const mockCartGroups: CartGroup[] = [
  {
    storeId: 'store-1',
    storeName: '茶心阁 · 茶心门店',
    items: [
      { id: 'menu-2', name: '桂花乌龙 · 冷萃', spec: '冷萃 500ml ｜ 微糖 ｜ 去冰', price: 26, quantity: 1 },
      { id: 'menu-5', name: '白桃乌龙 · 热泡', spec: '热泡 400ml ｜ 少糖 ｜ 常温', price: 28, quantity: 2 }
    ]
  },
  {
    storeId: 'store-2',
    storeName: '茶心阁 · 城南店',
    items: [
      { id: 'menu-7', name: '一茶三喝 · 翠玉白茶', spec: '三段式风味体验', price: 68, quantity: 1 }
    ]
  }
];
