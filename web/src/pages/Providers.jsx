import { useEffect, useState, useCallback } from 'react';
import { Table, Button, Modal, Form, Input, Select, Switch, Space, Tag, Drawer, message, Popconfirm } from 'antd';
import { PlusOutlined, SyncOutlined, UnorderedListOutlined, ClockCircleOutlined } from '@ant-design/icons';
import { useI18n } from '../i18n';
import api from '../api';

// 同步频率选项
const SYNC_INTERVAL_OPTIONS = [
  { value: 'none', labelKey: 'provider.syncInterval.none' },
  { value: 'hourly', labelKey: 'provider.syncInterval.hourly' },
  { value: 'daily', labelKey: 'provider.syncInterval.daily' },
  { value: 'weekly', labelKey: 'provider.syncInterval.weekly' },
];

export default function Providers() {
  const [providers, setProviders] = useState([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [models, setModels] = useState([]);
  const [drawerProvider, setDrawerProvider] = useState(null);
  const [form] = Form.useForm();
  const [syncScheduleOpen, setSyncScheduleOpen] = useState(false);
  const [syncSchedule, setSyncSchedule] = useState({});
  const [syncScheduleSaving, setSyncScheduleSaving] = useState(false);
  const [existingTags, setExistingTags] = useState([]);
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

  // 全部同步
  const handleSyncAll = async () => {
    message.loading({ content: t('provider.syncAll') + '...', key: 'syncAll' });
    try {
      await api.post('/providers/sync-all');
      message.success({ content: t('common.success'), key: 'syncAll' });
    } catch {
      message.error({ content: t('common.failed'), key: 'syncAll' });
    }
    setTimeout(load, 2000);
  };

  // 打开定时同步设置
  const openSyncSchedule = () => {
    const schedule = {};
    providers.forEach((p) => {
      schedule[p.id] = p.sync_interval || 'none';
    });
    setSyncSchedule(schedule);
    setSyncScheduleOpen(true);
  };

  // 保存定时同步设置
  const handleSaveSyncSchedule = async () => {
    setSyncScheduleSaving(true);
    try {
      const promises = providers.map((p) => {
        const newInterval = syncSchedule[p.id] || 'none';
        if (newInterval !== (p.sync_interval || 'none')) {
          return api.put(`/providers/${p.id}`, { sync_interval: newInterval });
        }
        return Promise.resolve();
      });
      await Promise.all(promises);
      message.success(t('common.success'));
      setSyncScheduleOpen(false);
      load();
    } catch (e) {
      message.error(e.response?.data?.error || t('common.failed'));
    } finally {
      setSyncScheduleSaving(false);
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

  const handleBulkToggle = async (enabled) => {
    if (!drawerProvider?.id) return;

    try {
      const { data } = await api.put(`/providers/${drawerProvider.id}/models/bulk`, { enabled });
      setModels(data);
      message.success(t('common.success'));
    } catch (error) {
      message.error(error.response?.data?.error || t('common.failed'));
    }
  };

  const columns = [
    { title: t('provider.name'), dataIndex: 'name', key: 'name' },
    { title: t('provider.protocol'), dataIndex: 'protocol', key: 'protocol', render: (v) => <Tag color={v === 'openai' ? 'blue' : 'orange'}>{v}</Tag> },
    { title: t('provider.baseurl'), dataIndex: 'base_url', key: 'base_url', ellipsis: true },
    { title: t('common.enabled'), key: 'enabled', render: (_, r) => <Switch size="small" checked={r.enabled} onChange={() => api.put(`/providers/${r.id}/toggle`).then(load)} /> },
    {
      title: t('provider.syncInterval'), key: 'sync_interval',
      render: (_, r) => {
        if (!r.sync_interval || r.sync_interval === 'none') return <Tag>{t('provider.syncInterval.none')}</Tag>;
        const opt = SYNC_INTERVAL_OPTIONS.find((o) => o.value === r.sync_interval);
        return <Tag color="cyan">{opt ? t(opt.labelKey) : r.sync_interval}</Tag>;
      },
    },
    { title: t('provider.sync.status'), dataIndex: 'sync_status', key: 'sync', render: (v) => <Tag color={v === 'success' ? 'green' : v === 'failed' ? 'red' : 'default'}>{v || t('common.pending')}</Tag> },
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

  // 定时同步设置表格列
  const syncScheduleColumns = [
    { title: t('provider.name'), dataIndex: 'name', key: 'name' },
    { title: t('provider.protocol'), dataIndex: 'protocol', key: 'protocol', render: (v) => <Tag color={v === 'openai' ? 'blue' : 'orange'}>{v}</Tag> },
    {
      title: t('provider.syncInterval'), key: 'sync_interval',
      render: (_, r) => (
        <Select
          size="small"
          value={syncSchedule[r.id] || 'none'}
          onChange={(v) => setSyncSchedule((prev) => ({ ...prev, [r.id]: v }))}
          style={{ width: 130 }}
          options={SYNC_INTERVAL_OPTIONS.map((o) => ({ value: o.value, label: t(o.labelKey) }))}
        />
      ),
    },
  ];

  return (
    <>
      <div style={{ marginBottom: 16, display: 'flex', gap: 8 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>{t('provider.add')}</Button>
        <Button icon={<SyncOutlined />} onClick={handleSyncAll}>{t('provider.syncAll')}</Button>
        <Button icon={<ClockCircleOutlined />} onClick={openSyncSchedule}>{t('provider.syncSchedule')}</Button>
      </div>
      <Table dataSource={providers} columns={columns} rowKey="id" loading={loading} size="middle" scroll={{ x: 'max-content' }} />

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
            <Button type="primary" onClick={() => handleBulkToggle(true)}>{t('provider.model.enableAll')}</Button>
            <Button onClick={() => handleBulkToggle(false)}>{t('provider.model.disableAll')}</Button>
          </Space>
        </div>
        <Table dataSource={models} rowKey="id" size="middle" pagination={false}
          columns={[
            { title: t('provider.model.name'), dataIndex: 'model_name', key: 'model_name' },
            { title: t('provider.model.enabled'), key: 'enabled', render: (_, r) => <Switch size="small" checked={r.enabled} onChange={() => toggleModel(r.id)} /> },
          ]}
        />
      </Drawer>

      {/* 定时同步设置 Modal */}
      <Modal
        title={t('provider.syncSchedule.title')}
        open={syncScheduleOpen}
        onCancel={() => setSyncScheduleOpen(false)}
        onOk={handleSaveSyncSchedule}
        confirmLoading={syncScheduleSaving}
        width={600}
        destroyOnClose
      >
        <p style={{ color: '#888', marginBottom: 16 }}>{t('provider.syncSchedule.desc')}</p>
        <Table
          dataSource={providers}
          columns={syncScheduleColumns}
          rowKey="id"
          size="small"
          pagination={false}
        />
      </Modal>
    </>
  );
}
