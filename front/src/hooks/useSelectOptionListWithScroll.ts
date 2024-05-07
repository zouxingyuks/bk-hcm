import { Ref, reactive, ref, watch } from 'vue';
// import stores
import { useResourceStore } from '@/store';
// import types
import { QueryRuleOPEnum, RulesItem } from '@/typings';

/**
 * Select Option List - 支持滚动加载
 */
export default (
  // 加载的资源类型
  type: string,
  // 搜索条件
  rules: any,
  // 是否立即执行
  immediate = true,
  protocol: Ref<string> = ref('TCP'),
) => {
  // use stores
  const resourceStore = useResourceStore();

  // define data
  const pagination = reactive({
    start: 0,
    limit: 20,
    hasNext: true,
  });
  const isScrollLoading = ref(false); // 滚动加载loading
  const isFlashLoading = ref(false); // 刷新操作loading
  const optionList = ref([]);

  /**
   * 初始化状态: optionList, pagination
   */
  const initState = () => {
    optionList.value = [];
    Object.assign(pagination, {
      start: 0,
      limit: 20,
      hasNext: true,
    });
  };

  /**
   * 请求option list
   */
  const getOptionList = async (customRules: RulesItem[] = []) => {
    isScrollLoading.value = true;
    try {
      const [detailRes, countRes] = await Promise.all(
        [false, true].map((isCount) =>
          resourceStore.list(
            {
              filter: {
                op: QueryRuleOPEnum.AND,
                rules: [
                  ...rules,
                  ...customRules,
                  ...(['target_groups'].includes(type)
                    ? [
                        {
                          field: 'protocol',
                          op: QueryRuleOPEnum.EQ,
                          value: protocol.value,
                        },
                      ]
                    : []),
                ],
              },
              page: {
                count: isCount,
                start: isCount ? 0 : pagination.start,
                limit: isCount ? 0 : pagination.limit,
                sort: isCount ? undefined : 'created_at',
                order: isCount ? undefined : 'DESC',
              },
            },
            type,
          ),
        ),
      );

      // 将新获取的option添加至list中
      optionList.value = [...optionList.value, ...detailRes.data.details];
      if (optionList.value.length >= countRes.data.count) {
        // option列表加载完毕
        pagination.hasNext = false;
      } else {
        // option列表未加载完毕
        pagination.start += pagination.limit;
      }
    } finally {
      isScrollLoading.value = false;
    }
  };

  watch(
    () => protocol.value,
    () => {
      initState();
      getOptionList();
    },
  );

  /**
   * 滚动触底加载更多
   */
  const handleOptionListScrollEnd = () => {
    if (!pagination.hasNext) return;
    getOptionList();
  };

  /**
   * 刷新options list
   */
  const handleRefreshOptionList = async () => {
    initState();
    try {
      isFlashLoading.value = true;
      await getOptionList();
    } finally {
      isFlashLoading.value = false;
    }
  };

  // 立即执行
  immediate && getOptionList();

  return {
    isScrollLoading,
    optionList,
    initState,
    getOptionList,
    handleOptionListScrollEnd,
    isFlashLoading,
    handleRefreshOptionList,
  };
};
