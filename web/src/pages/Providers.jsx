import { useCallback, useEffect, useMemo, useState } from 'react';
import { Table, Button, Modal, Form, Input, Select, Switch, Space, Tag, Drawer, message, Popconfirm, Card, Typography } from 'antd';
import { FilterOutlined, PlusOutlined, SyncOutlined, UnorderedListOutlined } from '@ant-design/icons';
import { useI18n } from '../i18n';
import api from '../api';

const { Title } = Typography;

// Pre-defined tag colors for visual variety
const TAG_COLORS = ['blue', 'green', 'orange', 'purple', 'cyan', 'magenta', 'gold', 'lime', 'geekblue', 'volcano'];
function tagColor(tag) {
  let hash = 0;
  for (let i = 0; i < tag.length; i++) hash = tag.charCodeAt(i) + ((hash << 5) - hash);
  return TAG_COLORS[Math.abs(hash) % TAG_COLORS.length];
}

export default function Providers() {
  const [providers, setProviders] = useState([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [models, setModels] = useState([]);
  const [drawerProvider, setDrawerProvider] = useState(null);
  const [syncAllLoading, setSyncAllLoading] = useState(false);
  const [existingTags, setExistingTags] = useState([]);
  const [filterTags, setFilterTags] = useState([]);
  const [form] = Form.useForm();
  const { t } = useI18n();

  const loadTags = useCallback(() => {
    api.get('/providers/tags').then((r) => setExistingTags(r.data || [])).catch(() => { });
  }, []);

  const load = useCallback(() => {
    setLoading(true);
    api.get('/providers')
      .then((r) => setProviders(r.data))
      .catch((error) => {
        setProviders([]);
        message.error(error.response?.data?.error || t('common.failed'));
      })
      .finally(() => setLoading(false));
    loadTags();
  }, [t, loadTags]);

  useEffect(() => {
    const timer = setTimeout(() => { load(); }, 0);
    return () => clearTimeout(timer);
  }, [load]);

  useEffect(() => {
    if (!providers.some((provider) => provider.sync_status === 'syncing')) {
      return undefined;
    }

    const timer = setInterval(() => {
      load();
    }, 3000);

    return () => clearInterval(timer);
  }, [providers, load]);

  const openCreate = () => { setEditing(null); form.resetFields(); setModalOpen(true); };
  const openEdit = (r) => {
    setEditing(r);
    form.setFieldsValue({
      name: r.name,
      protocol: r.protocol,
      base_url: r.base_url,
      api_key: '',
      tags: r.tags ? r.tags.split(',').map((s) => s.trim()).filter(Boolean) : [],
    });
    setModalOpen(true);
  };

  const handleSubmit = async () => {
    const values = await form.validateFields();
    try {
      if (editing) {
        const payload = { name: values.name, base_url: values.base_url, tags: values.tags || [] };
        if (values.api_key) payload.api_key = values.api_key;
        await api.put(`/providers/${editing.id}`, payload);
      } else {
        await api.post('/providers', { ...values, tags: values.tags || [] });
      }
      message.success(t('common.success'));
      setModalOpen(false);
      load();
    } catch (e) {
      message.error(e.response?.data?.error || t('common.failed'));
    }
  };

  const handleSync = async (id) => {
    message.loading({ content: t('provider.sync') + '...', key: 'sync' });
    try {
      await api.post(`/providers/${id}/sync`);
      message.success({ content: t('common.success'), key: 'sync' });
      if (drawerProvider?.id === id) openModels({ id });
    } catch {
      message.error({ content: t('common.failed'), key: 'sync' });
    }
    load();
  };

  const handleSyncAll = async () => {
    setSyncAllLoading(true);
    message.loading({ content: t('provider.syncAll') + '...', key: 'sync-all' });
    try {
      await api.post('/providers/sync-all');
      message.success({ content: t('common.success'), key: 'sync-all' });
    } catch (error) {
      message.error({ content: error.response?.data?.error || t('common.failed'), key: 'sync-all' });
    } finally {
      setSyncAllLoading(false);
      load();
    }
  };

  const openModels = async (r) => {
    setDrawerProvider(r);
    const { data } = await api.get(`/providers/${r.id}/models`);
    setModels(data);
    setDrawerOpen(true);
  };

  const toggleModel = async (mid) => {
    await api.put(`/providers/${drawerProvider.id}/models/${mid}/toggle`);
    const { data } = await api.get(`/providers/${drawerProvider.id}/models`);
    setModels(data);
  };

  const handleBulkToggle = async (enable) => {
    for (const model of models) {
      if (model.enabled !== enable) {
        await api.put(`/providers/${drawerProvider.id}/models/${model.id}/toggle`);
      }
    }
    const { data } = await api.get(`/providers/${drawerProvider.id}/models`);
    setModels(data);
    message.success(t('common.success'));
  };

  // Filtered providers based on selected tags
  const filteredProviders = useMemo(() => {
    if (!filterTags.length) return providers;
    return providers.filter((p) => {
      if (!p.tags) return false;
      const providerTags = p.tags.split(',').map((s) => s.trim());
      return filterTags.some((ft) => providerTags.includes(ft));
    });
  }, [providers, filterTags]);

  const columns = [
    { title: t('provider.name'), dataIndex: 'name', key: 'name', width: '15%' },
    { title: t('provider.protocol'), dataIndex: 'protocol', key: 'protocol', render: (v) => <Tag color={v === 'openai' ? 'blue' : 'orange'}>{v}</Tag>, width: '10%' },
    {
      title: t('provider.tags'), dataIndex: 'tags', key: 'tags', width: '15%',
      render: (v) => v ? v.split(',').map((s) => s.trim()).filter(Boolean).map((tag) => (
        <Tag key={tag} color={tagColor(tag)} style={{ marginBottom: 4 }}>{tag}</Tag>
      )) : <Typography.Text type="secondary">-</Typography.Text>,
    },
    { title: t('provider.baseurl'), dataIndex: 'base_url', key: 'base_url', ellipsis: true, width: '20%' },
    { title: t('common.enabled'), key: 'enabled', render: (_, r) => <Switch size="small" checked={r.enabled} onChange={() => api.put(`/providers/${r.id}/toggle`).then(load)} />, width: '8%' },
    { title: t('provider.sync.status'), dataIndex: 'sync_status', key: 'sync', render: (v) => <Tag color={v === 'success' ? 'green' : v === 'failed' ? 'red' : 'default'}>{v || t('common.pending')}</Tag>, width: '10%' },
    {
      title: t('common.actions'), key: 'actions',
      render: (_, r) => (
        <Space size="small">
          <Button size="small" type="primary" ghost icon={<UnorderedListOutlined />} onClick={() => openModels(r)}>{t('provider.models')}</Button>
          <Button size="small" icon={<SyncOutlined />} onClick={() => handleSync(r.id)}>{t('provider.sync')}</Button>
          <Button size="small" onClick={() => openEdit(r)}>{t('common.edit')}</Button>
          <Popconfirm title={t('common.confirm_delete')} onConfirm={() => api.delete(`/providers/${r.id}`).then(() => { message.success(t('common.deleted')); load(); })}>
            <Button size="small" danger>{t('common.delete')}</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div className="dashboard-container">
      <div className="dashboard-header mb-6">
        <Title level={2} style={{ marginTop: 0, marginBottom: '24px', fontWeight: 700, color: '#1e293b' }}>
          {t('menu.providers')}
        </Title>
      </div>

      <Card className="premium-card">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16, gap: 12, flexWrap: 'wrap' }}>
          <Space wrap>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate} size="large" style={{ borderRadius: 8 }}>
              {t('provider.add')}
            </Button>
            <Select
              mode="multiple"
              allowClear
              placeholder={<span><FilterOutlined /> {t('provider.tags.filter')}</span>}
              value={filterTags}
              onChange={setFilterTags}
              style={{ minWidth: 220 }}
              options={existingTags.map((tag) => ({ value: tag, label: tag }))}
              size="large"
            />
          </Space>
          <Button icon={<SyncOutlined />} onClick={handleSyncAll} loading={syncAllLoading} size="large" style={{ borderRadius: 8 }}>
            {t('provider.syncAll')}
          </Button>
        </div>
        <Table
          dataSource={filteredProviders}
          columns={columns}
          rowKey="id"
          loading={loading}
          size="middle"
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `Total ${total} items`,
          }}
        />
      </Card>

      <Modal title={editing ? t('provider.edit') : t('provider.add')} open={modalOpen} onOk={handleSubmit} onCancel={() => setModalOpen(false)} destroyOnClose centered>
        <Form form={form} layout="vertical">
          <Form.Item name="name" label={t('provider.name')} rules={[{ required: true }]}><Input size="large" /></Form.Item>
          {!editing && <Form.Item name="protocol" label={t('provider.protocol')} rules={[{ required: true }]}><Select size="large" options={[{ value: 'openai', label: 'OpenAI' }, { value: 'anthropic', label: 'Anthropic' }]} /></Form.Item>}
          <Form.Item name="base_url" label={t('provider.baseurl')} rules={[{ required: !editing }]}><Input size="large" placeholder="https://api.openai.com" /></Form.Item>
          <Form.Item name="api_key" label={editing ? t('provider.apikey.keep') : t('provider.apikey')} rules={[{ required: !editing }]}><Input.Password size="large" /></Form.Item>
          <Form.Item name="tags" label={t('provider.tags')}>
            <Select
              mode="tags"
              size="large"
              placeholder={t('provider.tags.placeholder')}
              options={existingTags.map((tag) => ({ value: tag, label: tag }))}
            />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer title={`${t('provider.models')} - ${drawerProvider?.name || ''}`} open={drawerOpen} onClose={() => setDrawerOpen(false)} width={480}>
        <div style={{ marginBottom: 16 }}>
          <Space>
            <Button type="primary" onClick={() => handleBulkToggle(true)}>全启用</Button>
            <Button onClick={() => handleBulkToggle(false)}>全不启用</Button>
          </Space>
        </div>
        <Table dataSource={models} rowKey="id" size="middle" pagination={false}
          columns={[
            { title: t('provider.model.name'), dataIndex: 'model_name', key: 'model_name' },
            { title: t('provider.model.enabled'), key: 'enabled', render: (_, r) => <Switch size="small" checked={r.enabled} onChange={() => toggleModel(r.id)} /> },
          ]}
        />
      </Drawer>
    </div>
  );
}
