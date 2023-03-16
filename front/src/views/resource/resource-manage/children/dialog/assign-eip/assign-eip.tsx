import {
  Table,
  Loading,
  Radio,
  Message,
} from 'bkui-vue';
import {
  defineComponent,
  h,
  ref,
} from 'vue';
import {
  useI18n,
} from 'vue-i18n';
import StepDialog from '@/components/step-dialog/step-dialog';
import useQueryList  from '../../../hooks/use-query-list';
import useColumns from '../../../hooks/use-columns';
import {
  useResourceStore
} from '@/store/resource';

// 绑定eip
export default defineComponent({
  components: {
    StepDialog,
  },

  props: {
    isShow: {
      type: Boolean,
    },
    vendor: {
      type: String,
    },
    id: {
      type: String
    },
  },

  emits: ['update:isShow'],

  setup(props, { emit }) {
    const {
      t,
    } = useI18n();

    const type = ['tcloud', 'aws'].includes(props.vendor) ? 'cvms' : 'network_interfaces'
    const columnType = ['tcloud', 'aws'].includes(props.vendor) ? 'cvms' : 'networkInterface'

    const {
      datas,
      pagination,
      isLoading,
      handlePageChange,
      handlePageSizeChange,
      handleSort,
    } = useQueryList(
      {
        filter: {
          op: 'and',
          rules: [{
            field: 'vendor',
            op: 'eq',
            value: props.vendor,
          }],
        },
      },
      type
    );

    const columns = useColumns(columnType, true);
    const resourceStore = useResourceStore();
    const selection = ref<any>({});
    const isConfirmLoading = ref(false);
    const renderColumns = [
      {
        label: 'ID',
        field: 'id',
        render({ data }: any) {
          return h(
            Radio,
            {
              'model-value': selection.value.id,
              label: data.id,
              key: data.id,
              onChange() {
                selection.value = data;
              },
            }
          );
        },
      },
      ...columns
    ]

    // 方法
    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = () => {
      isConfirmLoading.value = true;
      const postData = type === 'cvms'
        ? {
          eip_id: props.id,
          cvm_id: selection.value.id
        }
        : {
          eip_id: props.id,
          network_interface_id: selection.value.id,
          cvm_id: selection.value.cvm_id
        }
      resourceStore
        .associateEip(postData)
        .then(() => {
          handleClose();
        })
        .catch((err: any) => {
          Message({
            theme: 'error',
            message: err.message || err
          })
        })
        .finally(() => {
          isConfirmLoading.value = false;
        })
    };

    return {
      datas,
      pagination,
      isLoading,
      renderColumns,
      isConfirmLoading,
      handlePageChange,
      handlePageSizeChange,
      handleSort,
      t,
      handleClose,
      handleConfirm,
    };
  },

  render() {
    const steps = [
      {
        isConfirmLoading: this.isConfirmLoading,
        component: () =>
          <Loading loading={this.isLoading}>
            <Table
              class="mt20"
              row-hover="auto"
              remote-pagination
              pagination={this.pagination}
              columns={this.renderColumns}
              data={this.datas}
              onPageLimitChange={this.handlePageSizeChange}
              onPageValueChange={this.handlePageChange}
              onColumnSort={this.handleSort}
            />
          </Loading>
      },
    ];

    return <>
      <step-dialog
        title="绑定EIP"
        isShow={this.isShow}
        steps={steps}
        onConfirm={this.handleConfirm}
        onCancel={this.handleClose}
      >
      </step-dialog>
    </>;
  },
});