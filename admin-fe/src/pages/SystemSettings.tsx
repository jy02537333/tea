import { useMemo } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Alert, Button, Card, Form, Input, Space, Tabs, Typography, message } from 'antd';
import { listSystemConfigs, upsertSystemConfigs } from '../services/systemConfigs';
import { useAuthContext } from '../hooks/useAuth';

const { Title } = Typography;

const BASIC_KEYS = ['site_logo_url', 'site_phone', 'site_copyright'] as const;
const CONTENT_KEYS = ['content_about', 'content_help', 'content_privacy', 'content_terms'] as const;

type FormValues = Record<(typeof BASIC_KEYS)[number] | (typeof CONTENT_KEYS)[number], string>;

function toMap(list: { config_key: string; config_value: string }[]) {
  const map: Record<string, string> = {};
  for (const item of list) map[item.config_key] = item.config_value ?? '';
  return map;
}

export default function SystemSettingsPage() {
  const queryClient = useQueryClient();
  const { hasPermission } = useAuthContext();
  const canManage = hasPermission('system:config:manage');

  const allKeys = useMemo(() => [...BASIC_KEYS, ...CONTENT_KEYS] as string[], []);
  const [form] = Form.useForm<FormValues>();

  const query = useQuery({
    queryKey: ['system-configs', allKeys.join(',')],
    queryFn: () => listSystemConfigs({ keys: allKeys }),
  });

  const initialValues = useMemo(() => {
    const list = query.data?.list ?? [];
    const map = toMap(list);
    const v: any = {};
    for (const k of allKeys) v[k] = map[k] ?? '';
    return v as FormValues;
  }, [allKeys, query.data?.list]);

  const saveMutation = useMutation({
    mutationFn: async (values: FormValues) => {
      const items = allKeys.map((k) => ({
        config_key: k,
        config_value: String((values as any)[k] ?? ''),
        config_type: 'string',
        status: 1,
      }));
      return upsertSystemConfigs(items);
    },
    onSuccess: async () => {
      message.success('已保存');
      await queryClient.invalidateQueries({ queryKey: ['system-configs'] });
    },
    onError: (err: any) => message.error(err?.message || '保存失败'),
  });

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
        <Title level={4} style={{ margin: 0 }}>
          系统设置
        </Title>
        <Space>
          <Button onClick={() => query.refetch()} loading={query.isFetching}>
            刷新
          </Button>
          <Button type="primary" onClick={() => form.submit()} loading={saveMutation.isPending} disabled={!canManage}>
            保存
          </Button>
        </Space>
      </Space>

      {!canManage && (
        <Alert
          type="warning"
          showIcon
          message="当前账号缺少 system:config:manage 权限，已禁用保存按钮。"
        />
      )}

      {query.isError && (
        <Alert
          type="error"
          showIcon
          message="加载系统配置失败"
          description={(query.error as any)?.message || '请确认已登录且具备 system:config:view 权限'}
        />
      )}

      <Form
        form={form}
        layout="vertical"
        initialValues={initialValues}
        onFinish={(values) => saveMutation.mutate(values)}
      >
        <Tabs
          items={[
            {
              key: 'basic',
              label: '基础配置',
              children: (
                <Card>
                  <Form.Item label="Logo URL" name="site_logo_url">
                    <Input placeholder="https://..." allowClear />
                  </Form.Item>
                  <Form.Item label="客服电话" name="site_phone">
                    <Input placeholder="例如 400-xxx-xxxx" allowClear />
                  </Form.Item>
                  <Form.Item label="版权信息" name="site_copyright">
                    <Input placeholder="例如 © 茶心阁" allowClear />
                  </Form.Item>
                </Card>
              ),
            },
            {
              key: 'content',
              label: '内容管理',
              children: (
                <Card>
                  <Form.Item label="关于我们" name="content_about">
                    <Input.TextArea rows={6} placeholder="支持纯文本/Markdown（如后端渲染可再扩展）" />
                  </Form.Item>
                  <Form.Item label="帮助文档" name="content_help">
                    <Input.TextArea rows={6} />
                  </Form.Item>
                  <Form.Item label="隐私政策" name="content_privacy">
                    <Input.TextArea rows={6} />
                  </Form.Item>
                  <Form.Item label="用户协议" name="content_terms">
                    <Input.TextArea rows={6} />
                  </Form.Item>
                </Card>
              ),
            },
          ]}
        />
      </Form>
    </Space>
  );
}
