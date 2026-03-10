import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, DatePicker, Form, Input, Modal, Popconfirm, Select, Space, Switch, Table, Tag, Typography, Card, message } from 'antd';
import { PlusOutlined, SettingOutlined, CopyOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useI18n } from '../i18n';
import api from '../api';

const { Text } = Typography;

export default function APIKeys() {
  const [keys, setKeys] = useState([]);
  const [configs, setConfigs] = useState([]);
  const [loading, setLoading] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const { t } = useI18n();

  const enabledConfigs = useMemo(
    () => configs.filter((config) => config.enabled),
    [configs],
  );

  const load = async () => {
    setLoading(true);
    try {
      const [keyRes, configRes] = await Promise.all([
        api.get('/apikeys'),
        api.get('/model-configs'),
      ]);
      setKeys(keyRes.data || []);
      setConfigs(configRes.data || []);
    } catch {
      message.error(t('common.failed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const handleCreate = async () => {
    const values = await form.validateFields();
    const payload = {
      name: values.name,
      model_config_id: values.model_config_id,
    };

    if (values.expires_at) {
      payload.expires_at = values.expires_at.toISOString();
    }

    try {
      await api.post('/apikeys', payload);
      message.success(t('common.created'));
      setCreateOpen(false);
      form.resetFields();
      load();
    } catch (error) {
      message.error(error.response?.data?.error || t('common.failed'));
    }
  };

  const openCreate = () => {
    if (!enabledConfigs.length) {
      message.warning(t('apikey.config.required'));
      return;
    }
    form.resetFields();
    setCreateOpen(true);
  };

  const columns = [
    { title: t('apikey.name'), dataIndex: 'name', key: 'name' },
    {
      title: t('apikey.key'),
      dataIndex: 'key',
      key: 'key',
      render: (value) => (
        <Space>
          <Tag>{value}</Tag>
          <Button
            size="small"
            type="text"
            icon={<CopyOutlined />}
            onClick={() => {
              navigator.clipboard.writeText(value);
              message.success(t('apikey.copied'));
            }}
          />
        </Space>
      ),
    },
    {
      title: t('apikey.config.schemeColumn'),
      key: 'model_config',
      render: (_, record) => record.model_config ? (
        <Tag color={record.model_config.enabled ? 'blue' : 'default'}>{record.model_config.name}</Tag>
      ) : (
        <Text type="secondary">{t('apikey.config.none')}</Text>
      ),
    },
    {
      title: t('common.enabled'),
      key: 'enabled',
      render: (_, record) => (
        <Switch
          size="small"
          checked={record.enabled}
          onChange={() => api.put(`/apikeys/${record.id}/toggle`).then(load).catch(() => message.error(t('common.failed')))}
        />
      ),
    },
    {
      title: t('apikey.expires'),
      dataIndex: 'expires_at',
      key: 'expires',
      render: (value) => value ? new Date(value).toLocaleDateString() : t('apikey.expires.never'),
    },
    {
      title: t('apikey.lastused'),
      dataIndex: 'last_used_at',
      key: 'last_used',
      render: (value) => value ? new Date(value).toLocaleString() : t('apikey.lastused.never'),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      render: (_, record) => (
        <Space size="small">
          <Button size="small" type="primary" icon={<SettingOutlined />} onClick={() => navigate(`/apikeys/${record.id}`)}>
            {t('apikey.config.manage')}
          </Button>
          <Popconfirm
            title={t('common.confirm_delete')}
            onConfirm={() => api.delete(`/apikeys/${record.id}`).then(() => {
              message.success(t('common.deleted'));
              load();
            }).catch(() => message.error(t('common.failed')))}
          >
            <Button size="small" danger>{t('common.delete')}</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div className="dashboard-container">
      <div className="dashboard-header mb-6">
        <Typography.Title level={2} style={{ marginTop: 0, marginBottom: '24px', fontWeight: 700, color: '#1e293b' }}>
          {t('menu.apikeys')}
        </Typography.Title>
      </div>

      {!enabledConfigs.length && (
        <Alert
          style={{ marginBottom: 24, borderRadius: 8, border: 'none', background: '#fffbeb' }}
          type="warning"
          showIcon
          message={<Text strong style={{ color: '#d97706' }}>{t('apikey.config.required')}</Text>}
          action={<Button size="small" type="primary" onClick={() => navigate('/model-configs')}>{t('apikey.config.openConfigs')}</Button>}
        />
      )}

      <Card className="premium-card">
        <div style={{ display: 'flex', justifyContent: 'flex-start', marginBottom: 16 }}>
          <Button
            type="primary"
            size="large"
            style={{ borderRadius: 8 }}
            icon={<PlusOutlined />}
            onClick={openCreate}
            disabled={!enabledConfigs.length}
          >
            {t('apikey.add')}
          </Button>
        </div>

        <Table dataSource={keys} columns={columns} rowKey="id" loading={loading} size="middle" scroll={{ x: 'max-content' }} />

        <Modal
          title={t('apikey.add')}
          open={createOpen}
          onOk={handleCreate}
          onCancel={() => setCreateOpen(false)}
          destroyOnClose
        >
          <Form form={form} layout="vertical">
            <Form.Item name="name" label={t('apikey.name')} rules={[{ required: true }]}>
              <Input placeholder="My Key" />
            </Form.Item>
            <Form.Item name="model_config_id" label={t('apikey.config')} rules={[{ required: true }]}>
              <Select
                placeholder={t('apikey.config.placeholder')}
                options={enabledConfigs.map((config) => ({ value: config.id, label: config.name }))}
              />
            </Form.Item>
            <Form.Item name="expires_at" label={t('apikey.expires')}>
              <DatePicker showTime style={{ width: '100%' }} />
            </Form.Item>
          </Form>
        </Modal>
      </Card>
    </div>
  );
}

