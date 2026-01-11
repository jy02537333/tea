import { createComponent } from 'solid-js/web';
import { View, ScrollView } from '@tarojs/components';
import * as React from 'react';

// 定义常量
const PICKER_LINE_HEIGHT = 34; // px
const PICKER_VISIBLE_ITEMS = 7; // 可见行数
const PICKER_BLANK_ITEMS = 3; // 空白行数
function PickerGroupBasic(props) {
  const {
    range = [],
    rangeKey,
    columnId,
    updateIndex,
    onColumnChange,
    selectedIndex = 0,
    // 使用selectedIndex参数，默认为0
    colors = {}
  } = props;
  const [targetScrollTop, setTargetScrollTop] = React.useState(0);
  const scrollViewRef = React.useRef(null);
  const itemRefs = React.useRef([]);
  // 使用selectedIndex初始化当前索引
  const [currentIndex, setCurrentIndex] = React.useState(selectedIndex);
  // 触摸状态用于优化用户体验
  const [isTouching, setIsTouching] = React.useState(false);
  const itemHeightRef = React.useRef(PICKER_LINE_HEIGHT);
  React.useEffect(() => {
    if (scrollViewRef.current) {
      itemHeightRef.current = scrollViewRef.current.scrollHeight / scrollViewRef.current.childNodes.length;
    }
  }, [range.length]); // 只在range长度变化时重新计算
  // 获取选中的索引
  const getSelectedIndex = scrollTop => {
    return Math.round(scrollTop / itemHeightRef.current);
  };
  // 当selectedIndex变化时，调整滚动位置
  React.useEffect(() => {
    if (scrollViewRef.current && range.length > 0 && !isTouching) {
      setTargetScrollTop(selectedIndex * itemHeightRef.current);
      setCurrentIndex(selectedIndex);
    }
  }, [selectedIndex, range]);
  // 是否处于归中状态
  const isCenterTimerId = React.useRef(null);
  // 简化为直接在滚动结束时通知父组件
  const handleScrollEnd = () => {
    if (!scrollViewRef.current) return;
    if (isCenterTimerId.current) {
      clearTimeout(isCenterTimerId.current);
      isCenterTimerId.current = null;
    }
    // 做一个0.1s延时  0.1s之内没有新的滑动 则把选项归到中间 然后更新选中项
    isCenterTimerId.current = setTimeout(() => {
      if (!scrollViewRef.current) return;
      const scrollTop = scrollViewRef.current.scrollTop;
      const newIndex = getSelectedIndex(scrollTop);
      setIsTouching(false);
      setTargetScrollTop(newIndex * itemHeightRef.current + Math.random() * 0.001); // 随机数为了在一个项内滚动时强制刷新
      updateIndex(newIndex, columnId);
      onColumnChange === null || onColumnChange === void 0 ? void 0 : onColumnChange({
        columnId,
        index: newIndex
      });
    }, 100);
  };
  // 滚动处理 - 在滚动时计算索引然后更新选中项样式
  const handleScroll = () => {
    if (!scrollViewRef.current) return;
    if (isCenterTimerId.current) {
      clearTimeout(isCenterTimerId.current);
      isCenterTimerId.current = null;
    }
    const scrollTop = scrollViewRef.current.scrollTop;
    const newIndex = getSelectedIndex(scrollTop);
    if (newIndex !== currentIndex) {
      setCurrentIndex(newIndex);
    }
  };
  // 渲染选项
  const pickerItem = range.map((item, index) => {
    const content = rangeKey && item && typeof item === 'object' ? item[rangeKey] : item;
    return createComponent(View, {
      id: `picker-item-${columnId}-${index}`,
      key: index,
      ref: el => itemRefs.current[index] = el,
      className: `taro-picker__item${index === currentIndex ? ' taro-picker__item--selected' : ''}`,
      get style() {
        return {
          height: PICKER_LINE_HEIGHT,
          color: index === currentIndex ? colors.itemSelectedColor || undefined : colors.itemDefaultColor || undefined
        };
      },
      children: content
    });
  });
  const realPickerItems = [...new Array(PICKER_BLANK_ITEMS).fill(null).map((_, idx) => createComponent(View, {
    key: `blank-top-${idx}`,
    className: "taro-picker__item taro-picker__item--blank",
    style: {
      height: PICKER_LINE_HEIGHT
    }
  })), ...pickerItem, ...new Array(PICKER_BLANK_ITEMS).fill(null).map((_, idx) => createComponent(View, {
    key: `blank-bottom-${idx}`,
    className: "taro-picker__item taro-picker__item--blank",
    style: {
      height: PICKER_LINE_HEIGHT
    }
  }))];
  return createComponent(View, {
    className: "taro-picker__group",
    get children() {
      return [createComponent(View, {
        className: "taro-picker__mask"
      }), createComponent(View, {
        className: "taro-picker__indicator"
      }), createComponent(ScrollView, {
        ref: scrollViewRef,
        scrollY: true,
        showScrollbar: false,
        className: "taro-picker__content",
        style: {
          height: PICKER_LINE_HEIGHT * PICKER_VISIBLE_ITEMS
        },
        scrollTop: targetScrollTop,
        onScroll: handleScroll,
        onTouchStart: () => setIsTouching(true),
        onScrollEnd: handleScrollEnd,
        scrollWithAnimation: true,
        children: realPickerItems
      })];
    }
  });
}
// 时间选择器实现
function PickerGroupTime(props) {
  const {
    range = [],
    rangeKey,
    columnId,
    updateIndex,
    selectedIndex = 0,
    colors = {}
  } = props;
  const [targetScrollTop, setTargetScrollTop] = React.useState(0);
  const scrollViewRef = React.useRef(null);
  const itemRefs = React.useRef([]);
  const [currentIndex, setCurrentIndex] = React.useState(selectedIndex);
  const [isTouching, setIsTouching] = React.useState(false);
  const itemHeightRef = React.useRef(PICKER_LINE_HEIGHT);
  React.useEffect(() => {
    if (scrollViewRef.current) {
      itemHeightRef.current = scrollViewRef.current.scrollHeight / scrollViewRef.current.childNodes.length;
    }
  }, [range.length]); // 只在range长度变化时重新计算
  const getSelectedIndex = scrollTop => {
    return Math.round(scrollTop / itemHeightRef.current);
  };
  // 当selectedIndex变化时，调整滚动位置
  React.useEffect(() => {
    if (scrollViewRef.current && range.length > 0 && !isTouching) {
      setTargetScrollTop(selectedIndex * itemHeightRef.current);
      setCurrentIndex(selectedIndex);
    }
  }, [selectedIndex, range]);
  // 是否处于归中状态
  const isCenterTimerId = React.useRef(null);
  // 简化为直接在滚动结束时通知父组件
  const handleScrollEnd = () => {
    if (!scrollViewRef.current) return;
    if (isCenterTimerId.current) {
      clearTimeout(isCenterTimerId.current);
      isCenterTimerId.current = null;
    }
    // 做一个0.1s延时  0.1s之内没有新的滑动 则把选项归到中间 然后更新选中项
    isCenterTimerId.current = setTimeout(() => {
      if (!scrollViewRef.current) return;
      const scrollTop = scrollViewRef.current.scrollTop;
      const newIndex = getSelectedIndex(scrollTop);
      setIsTouching(false);
      // 调用updateIndex执行限位逻辑，获取是否触发了限位
      const isLimited = Boolean(updateIndex(newIndex, columnId, true));
      // 如果没有触发限位，才执行归中逻辑
      if (!isLimited) {
        setTargetScrollTop(newIndex * itemHeightRef.current + Math.random() * 0.001);
      }
    }, 100);
  };
  // 滚动处理
  const handleScroll = () => {
    if (!scrollViewRef.current) return;
    if (isCenterTimerId.current) {
      clearTimeout(isCenterTimerId.current);
      isCenterTimerId.current = null;
    }
    const scrollTop = scrollViewRef.current.scrollTop;
    const newIndex = getSelectedIndex(scrollTop);
    if (newIndex !== currentIndex) {
      setCurrentIndex(newIndex);
    }
  };
  // 渲染选项
  const pickerItem = range.map((item, index) => {
    const content = rangeKey && item && typeof item === 'object' ? item[rangeKey] : item;
    return createComponent(View, {
      id: `picker-item-${columnId}-${index}`,
      key: index,
      ref: el => itemRefs.current[index] = el,
      className: `taro-picker__item${index === currentIndex ? ' taro-picker__item--selected' : ''}`,
      get style() {
        return {
          height: PICKER_LINE_HEIGHT,
          color: index === currentIndex ? colors.itemSelectedColor || undefined : colors.itemDefaultColor || undefined
        };
      },
      children: content
    });
  });
  const realPickerItems = [...new Array(PICKER_BLANK_ITEMS).fill(null).map((_, idx) => createComponent(View, {
    key: `blank-top-${idx}`,
    className: "taro-picker__item taro-picker__item--blank",
    style: {
      height: PICKER_LINE_HEIGHT
    }
  })), ...pickerItem, ...new Array(PICKER_BLANK_ITEMS).fill(null).map((_, idx) => createComponent(View, {
    key: `blank-bottom-${idx}`,
    className: "taro-picker__item taro-picker__item--blank",
    style: {
      height: PICKER_LINE_HEIGHT
    }
  }))];
  return createComponent(View, {
    className: "taro-picker__group",
    get children() {
      return [createComponent(View, {
        className: "taro-picker__mask"
      }), createComponent(View, {
        className: "taro-picker__indicator"
      }), createComponent(ScrollView, {
        ref: scrollViewRef,
        scrollY: true,
        showScrollbar: false,
        className: "taro-picker__content",
        style: {
          height: PICKER_LINE_HEIGHT * PICKER_VISIBLE_ITEMS
        },
        scrollTop: targetScrollTop,
        onScroll: handleScroll,
        onTouchStart: () => setIsTouching(true),
        onScrollEnd: handleScrollEnd,
        scrollWithAnimation: true,
        children: realPickerItems
      })];
    }
  });
}
// 日期选择器实现
function PickerGroupDate(props) {
  const {
    range = [],
    columnId,
    updateDay,
    selectedIndex = 0,
    colors = {}
  } = props;
  const [targetScrollTop, setTargetScrollTop] = React.useState(0);
  const scrollViewRef = React.useRef(null);
  const [currentIndex, setCurrentIndex] = React.useState(selectedIndex);
  const [isTouching, setIsTouching] = React.useState(false);
  const itemHeightRef = React.useRef(PICKER_LINE_HEIGHT);
  React.useEffect(() => {
    if (scrollViewRef.current) {
      itemHeightRef.current = scrollViewRef.current.scrollHeight / scrollViewRef.current.childNodes.length;
    }
  }, [range.length]); // 只在range长度变化时重新计算
  const getSelectedIndex = scrollTop => {
    return Math.round(scrollTop / itemHeightRef.current);
  };
  // 当selectedIndex变化时，调整滚动位置
  React.useEffect(() => {
    if (scrollViewRef.current && range.length > 0 && !isTouching) {
      setTargetScrollTop(selectedIndex * itemHeightRef.current);
      setCurrentIndex(selectedIndex);
    }
  }, [selectedIndex, range]);
  // 是否处于归中状态
  const isCenterTimerId = React.useRef(null);
  // 简化为直接在滚动结束时通知父组件
  const handleScrollEnd = () => {
    if (!scrollViewRef.current) return;
    if (isCenterTimerId.current) {
      clearTimeout(isCenterTimerId.current);
      isCenterTimerId.current = null;
    }
    // 做一个0.1s延时  0.1s之内没有新的滑动 则把选项归到中间 然后更新选中项
    isCenterTimerId.current = setTimeout(() => {
      if (!scrollViewRef.current) return;
      const scrollTop = scrollViewRef.current.scrollTop;
      const newIndex = getSelectedIndex(scrollTop);
      setIsTouching(false);
      setTargetScrollTop(newIndex * itemHeightRef.current + Math.random() * 0.001); // 随机数为了在一个项内滚动时强制刷新
      // 更新日期值
      if (updateDay) {
        // 解析文本中的数字（移除年、月、日等后缀）
        const valueText = range[newIndex] || '';
        const numericValue = parseInt(valueText.replace(/[^0-9]/g, ''));
        updateDay(isNaN(numericValue) ? 0 : numericValue, parseInt(columnId));
      }
    }, 100);
  };
  // 滚动处理
  const handleScroll = () => {
    if (!scrollViewRef.current) return;
    if (isCenterTimerId.current) {
      clearTimeout(isCenterTimerId.current);
      isCenterTimerId.current = null;
    }
    const scrollTop = scrollViewRef.current.scrollTop;
    const newIndex = getSelectedIndex(scrollTop);
    if (newIndex !== currentIndex) {
      setCurrentIndex(newIndex);
    }
  };
  // 渲染选项
  const pickerItem = range.map((item, index) => {
    return createComponent(View, {
      id: `picker-item-${columnId}-${index}`,
      key: index,
      className: `taro-picker__item${index === currentIndex ? ' taro-picker__item--selected' : ''}`,
      get style() {
        return {
          height: PICKER_LINE_HEIGHT,
          color: index === currentIndex ? colors.itemSelectedColor || undefined : colors.itemDefaultColor || undefined
        };
      },
      children: item
    });
  });
  const realPickerItems = [...new Array(PICKER_BLANK_ITEMS).fill(null).map((_, idx) => createComponent(View, {
    key: `blank-top-${idx}`,
    className: "taro-picker__item taro-picker__item--blank",
    style: {
      height: PICKER_LINE_HEIGHT
    }
  })), ...pickerItem, ...new Array(PICKER_BLANK_ITEMS).fill(null).map((_, idx) => createComponent(View, {
    key: `blank-bottom-${idx}`,
    className: "taro-picker__item taro-picker__item--blank",
    style: {
      height: PICKER_LINE_HEIGHT
    }
  }))];
  return createComponent(View, {
    className: "taro-picker__group",
    get children() {
      return [createComponent(View, {
        className: "taro-picker__mask"
      }), createComponent(View, {
        className: "taro-picker__indicator"
      }), createComponent(ScrollView, {
        ref: scrollViewRef,
        scrollY: true,
        showScrollbar: false,
        className: "taro-picker__content",
        style: {
          height: PICKER_LINE_HEIGHT * PICKER_VISIBLE_ITEMS
        },
        scrollTop: targetScrollTop,
        onScroll: handleScroll,
        onTouchStart: () => setIsTouching(true),
        onScrollEnd: handleScrollEnd,
        scrollWithAnimation: true,
        children: realPickerItems
      })];
    }
  });
}
// 地区选择器实现
function PickerGroupRegion(props) {
  const {
    range = [],
    rangeKey,
    columnId,
    updateIndex,
    selectedIndex = 0,
    // 使用selectedIndex参数，默认为0
    colors = {}
  } = props;
  const scrollViewRef = React.useRef(null);
  const [targetScrollTop, setTargetScrollTop] = React.useState(0);
  const [currentIndex, setCurrentIndex] = React.useState(selectedIndex);
  const [isTouching, setIsTouching] = React.useState(false);
  const itemHeightRef = React.useRef(PICKER_LINE_HEIGHT);
  const isUserBeginScrollRef = React.useRef(false);
  React.useEffect(() => {
    if (scrollViewRef.current) {
      itemHeightRef.current = scrollViewRef.current.scrollHeight / scrollViewRef.current.childNodes.length;
    }
  }, [range.length]); // 只在range长度变化时重新计算
  const getSelectedIndex = scrollTop => {
    return Math.round(scrollTop / itemHeightRef.current);
  };
  // 当selectedIndex变化时，调整滚动位置
  React.useEffect(() => {
    if (scrollViewRef.current && range.length > 0 && !isTouching) {
      setTargetScrollTop(selectedIndex * itemHeightRef.current);
      setCurrentIndex(selectedIndex);
    }
  }, [selectedIndex, range]);
  // 滚动结束处理
  const isCenterTimerId = React.useRef(null);
  const handleScrollEnd = () => {
    if (!scrollViewRef.current) return;
    if (isCenterTimerId.current) {
      clearTimeout(isCenterTimerId.current);
      isCenterTimerId.current = null;
    }
    // 做一个0.1s延时  0.1s之内没有新的滑动 则把选项归到中间 然后更新选中项
    isCenterTimerId.current = setTimeout(() => {
      if (!scrollViewRef.current) return;
      const scrollTop = scrollViewRef.current.scrollTop;
      const newIndex = getSelectedIndex(scrollTop);
      setIsTouching(false);
      setTargetScrollTop(newIndex * itemHeightRef.current + Math.random() * 0.001); // 随机数为了在一个项内滚动时强制刷新
      updateIndex(newIndex, columnId, false, isUserBeginScrollRef.current);
    }, 100);
  };
  // 滚动处理 - 在滚动时计算索引
  const handleScroll = () => {
    if (!scrollViewRef.current) return;
    if (isCenterTimerId.current) {
      clearTimeout(isCenterTimerId.current);
      isCenterTimerId.current = null;
    }
    const scrollTop = scrollViewRef.current.scrollTop;
    const newIndex = getSelectedIndex(scrollTop);
    if (newIndex !== currentIndex) {
      setCurrentIndex(newIndex);
    }
  };
  // 渲染选项
  const pickerItem = range.map((item, index) => {
    const content = rangeKey && item && typeof item === 'object' ? item[rangeKey] : item;
    return createComponent(View, {
      id: `picker-item-${columnId}-${index}`,
      key: index,
      className: `taro-picker__item${index === currentIndex ? ' taro-picker__item--selected' : ''}`,
      get style() {
        return {
          height: PICKER_LINE_HEIGHT,
          color: index === currentIndex ? colors.itemSelectedColor || undefined : colors.itemDefaultColor || undefined
        };
      },
      children: content
    });
  });
  const realPickerItems = [...new Array(PICKER_BLANK_ITEMS).fill(null).map((_, idx) => createComponent(View, {
    key: `blank-top-${idx}`,
    className: "taro-picker__item taro-picker__item--blank",
    style: {
      height: PICKER_LINE_HEIGHT
    }
  })), ...pickerItem, ...new Array(PICKER_BLANK_ITEMS).fill(null).map((_, idx) => createComponent(View, {
    key: `blank-bottom-${idx}`,
    className: "taro-picker__item taro-picker__item--blank",
    style: {
      height: PICKER_LINE_HEIGHT
    }
  }))];
  return createComponent(View, {
    className: "taro-picker__group",
    get children() {
      return [createComponent(View, {
        className: "taro-picker__mask"
      }), createComponent(View, {
        className: "taro-picker__indicator"
      }), createComponent(ScrollView, {
        ref: scrollViewRef,
        scrollY: true,
        showScrollbar: false,
        className: "taro-picker__content",
        style: {
          height: PICKER_LINE_HEIGHT * PICKER_VISIBLE_ITEMS
        },
        scrollTop: targetScrollTop,
        onScroll: handleScroll,
        onTouchStart: () => {
          setIsTouching(true);
          isUserBeginScrollRef.current = true;
        },
        onScrollEnd: handleScrollEnd,
        scrollWithAnimation: true,
        children: realPickerItems
      })];
    }
  });
}
// 默认导出，根据 mode 自动分发
function PickerGroup(props) {
  switch (props.mode) {
    case 'time':
      return createComponent(PickerGroupTime, props);
    case 'date':
      return createComponent(PickerGroupDate, props);
    case 'region':
      return createComponent(PickerGroupRegion, props);
    default:
      return createComponent(PickerGroupBasic, props);
  }
}

export { PickerGroup, PickerGroupBasic, PickerGroupDate, PickerGroupRegion, PickerGroupTime };
//# sourceMappingURL=picker-group.js.map
