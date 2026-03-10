import { useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Badge,
  Button,
  Checkbox,
  Collapse,
  Divider,
  Drawer,
  Empty,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Space,
  Spin,
  Switch,
  Table,
  Tag,
  Typography,
  Card,
  message,
} from 'antd';
import {
  BorderOutlined,
  CheckSquareOutlined,
  CopyOutlined,
  PlusOutlined,
  SaveOutlined,
  SettingOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { useI18n } from '../i18n';
import api from '../api';

const { Text } = Typography;

function buildSelectedState(items = []) {
  const next = {};

  items.forEach((item) => {
    next[item.provider_model_id] = {
      provider_id: item.provider_id,
      model_name: item.provider_model?.model_name || '',
      mapped_name: item.mapped_name,
      priority: item.priority || 0,
    };
  });

  return next;
}

function filterSelectedByAvailable(selectedState = {}, groups = []) {
  const allowedModelIDs = new Set();
  groups.forEach((group) => {
    (group.models || []).forEach((model) => {
      allowedModelIDs.add(model.id);
    });
  });

  return Object.fromEntries(
    Object.entries(selectedState).filter(([providerModelId]) => allowedModelIDs.has(Number(providerModelId))),
  );
}

export default function ModelConfigs() {
  const { t } = useI18n();
  const [configs, setConfigs] = useState([]);
  const [loading, setLoading] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [cloneTarget, setCloneTarget] = useState(null);
  const [editorOpen, setEditorOpen] = useState(false);
  const [editorLoading, setEditorLoading] = useState(false);
  const [editingConfig, setEditingConfig] = useState(null);
  const [availableModels, setAvailableModels] = useState([]);
  const [selected, setSelected] = useState({});
  const [saving, setSaving] = useState(false);
  const [createForm] = Form.useForm();
  const [cloneForm] = Form.useForm();
  const [editorForm] = Form.useForm();
  const [quickMapOpen, setQuickMapOpen] = useState(false);
  const [quickMapPrefix, setQuickMapPrefix] = useState('');

  const loadConfigs = async () => {
    setLoading(true);
    try {
      const [configRes, modelsRes] = await Promise.all([
        api.get('/model-configs'),
        api.get('/model-configs/available-models'),
      ]);
      setConfigs(configRes.data || []);
      setAvailableModels(modelsRes.data || []);
    } catch (error) {
      setConfigs([]);
      setAvailableModels([]);
      message.error(error.response?.data?.error || t('common.failed'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadConfigs();
  }, []);

  const openEditor = async (configId) => {
    setEditorOpen(true);
    setEditorLoading(true);
    try {
      const [configRes, modelsRes] = await Promise.all([
        api.get(`/model-configs/${configId}`),
        api.get('/model-configs/available-models'),
      ]);

      const config = configRes.data;
      const groups = modelsRes.data || [];
      const selectedState = buildSelectedState(config.items || []);
      setEditingConfig(config);
      setAvailableModels(groups);
      setSelected(filterSelectedByAvailable(selectedState, groups));
      editorForm.setFieldsValue({
        name: config.name,
        description: config.description,
      });
    } catch (error) {
      message.error(error.response?.data?.error || t('common.failed'));
      setEditorOpen(false);
    } finally {
      setEditorLoading(false);
    }
  };

  const handleCreate = async () => {
    const values = await createForm.validateFields();
    try {
      const { data } = await api.post('/model-configs', values);
      message.success(t('common.created'));
      setCreateOpen(false);
      createForm.resetFields();
      await loadConfigs();
      openEditor(data.id);
    } catch (error) {
      message.error(error.response?.data?.error || t('common.failed'));
    }
  };

  const handleClone = async () => {
    const values = await cloneForm.validateFields();
    try {
      const { data } = await api.post(`/model-configs/${cloneTarget.id}/clone`, { name: values.name });
      message.success(t('modelConfig.cloneSuccess'));
      setCloneTarget(null);
      cloneForm.resetFields();
      await loadConfigs();
      openEditor(data.id);
    } catch (error) {
      message.error(error.response?.data?.error || t('common.failed'));
    }
  };

  const handleSave = async () => {
    const values = await editorForm.validateFields();
    const items = Object.entries(selected).map(([providerModelId, value]) => ({
      provider_id: value.provider_id,
      provider_model_id: Number(providerModelId),
      mapped_name: value.mapped_name?.trim() || value.model_name,
      priority: Number(value.priority) || 0,
    }));

    setSaving(true);
    try {
      await api.put(`/model-configs/${editingConfig.id}`, values);
      await api.put(`/model-configs/${editingConfig.id}/items`, { items });
      message.success(t('common.success'));
      await loadConfigs();
      await openEditor(editingConfig.id);
    } catch (error) {
      message.error(error.response?.data?.error || t('common.failed'));
    } finally {
      setSaving(false);
    }
  };

  const toggleModel = (providerId, model) => {
    setSelected((previous) => {
      if (previous[model.id]) {
        const next = { ...previous };
        delete next[model.id];
        return next;
      }

      return {
        ...previous,
        [model.id]: {
          provider_id: providerId,
          model_name: model.model_name,
          mapped_name: model.model_name,
          priority: 0,
        },
      };
    });
  };

  const selectAll = (providerId, models) => {
    setSelected((previous) => {
      const next = { ...previous };
      models.forEach((model) => {
        if (!next[model.id]) {
          next[model.id] = {
            provider_id: providerId,
            model_name: model.model_name,
            mapped_name: model.model_name,
            priority: 0,
          };
        }
      });
      return next;
    });
  };

  const deselectAll = (models) => {
    setSelected((previous) => {
      const next = { ...previous };
      models.forEach((model) => {
        delete next[model.id];
      });
      return next;
    });
  };

  const updateMappedName = (modelId, mappedName) => {
    setSelected((previous) => ({
      ...previous,
      [modelId]: {
        ...previous[modelId],
        mapped_name: mappedName,
      },
    }));
  };

  const updatePriority = (modelId, priority) => {
    setSelected((previous) => ({
      ...previous,
      [modelId]: {
        ...previous[modelId],
        priority: Number(priority) || 0,
      },
    }));
  };

  // 一键映射：将所有模型勾选并设置 mapped_name = prefix + model_name
  const applyQuickMap = () => {
    const prefix = quickMapPrefix;
    setSelected((previous) => {
      const next = { ...previous };
      availableModels.forEach((group) => {
        const providerId = group.provider.id;
        (group.models || []).forEach((model) => {
          next[model.id] = {
            provider_id: providerId,
            model_name: model.model_name,
            mapped_name: prefix ? prefix + model.model_name : model.model_name,
            priority: next[model.id]?.priority || 0,
          };
        });
      });
      return next;
    });
    message.success(t('common.success'));
  };

  const providerCount = availableModels.length || configs.length || 0;
  const totalModels = useMemo(
    () => availableModels.reduce((total, group) => total + (group.models?.length || 0), 0),
    [availableModels],
  );

  const collapseItems = useMemo(
    () => availableModels.map((group) => {
      const provider = group.provider;
      const models = group.models || [];
      const selectedInProvider = models.filter((model) => selected[model.id]).length;
      const allChecked = models.length > 0 && selectedInProvider === models.length;
      const someChecked = selectedInProvider > 0 && !allChecked;

      return {
        key: provider.id,
        label: (
          <Space>
            <Tag color={provider.protocol === 'openai' ? 'blue' : 'orange'}>{provider.protocol}</Tag>
            <span>{provider.name}</span>
            <Badge count={selectedInProvider} style={{ backgroundColor: '#1677ff' }} />
          </Space>
        ),
        extra: models.length ? (
          <Space size={4} onClick={(event) => event.stopPropagation()}>
            <Button size="small" type="text" icon={<CheckSquareOutlined />} onClick={() => selectAll(provider.id, models)} />
            <Button size="small" type="text" icon={<BorderOutlined />} onClick={() => deselectAll(models)} />
          </Space>
        ) : null,
        children: !models.length ? (
          <Text type="secondary">{t('modelConfig.noModels')}</Text>
        ) : (
          <div>
            <div style={{ marginBottom: 12 }}>
              <Checkbox checked={allChecked} indeterminate={someChecked} onChange={() => (allChecked ? deselectAll(models) : selectAll(provider.id, models))}>
                <Text type="secondary">{t('modelConfig.selectAll')}</Text>
              </Checkbox>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              {models.map((model) => {
                const selection = selected[model.id];
                return (
                  <div key={model.id} style={{ border: '1px solid #f0f0f0', borderRadius: 8, padding: 12 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
                      <Checkbox checked={!!selection} onChange={() => toggleModel(provider.id, model)}>
                        <code style={{ fontSize: 13 }}>{model.model_name}</code>
                      </Checkbox>
                      <Space size={6} wrap>
                        {!provider.enabled && <Tag>{t('modelConfig.providerDisabled')}</Tag>}
                        {!model.enabled && <Tag>{t('modelConfig.modelDisabled')}</Tag>}
                      </Space>
                    </div>
                    {selection && (
                      <Space wrap style={{ marginTop: 12 }}>
                        <Input
                          value={selection.mapped_name}
                          onChange={(event) => updateMappedName(model.id, event.target.value)}
                          style={{ width: 260 }}
                          placeholder={t('modelConfig.mappedHint')}
                          addonBefore={t('modelConfig.mappedName')}
                        />
                        <InputNumber min={0} value={selection.priority} onChange={(value) => updatePriority(model.id, value)} addonBefore="优先级" />
                      </Space>
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        ),
      };
    }),
    [availableModels, selected, t],
  );

  const columns = [
    {
      title: t('modelConfig.name'),
      dataIndex: 'name',
      key: 'name',
      render: (value) => <Text strong>{value}</Text>,
    },
    {
      title: t('modelConfig.description'),
      dataIndex: 'description',
      key: 'description',
      render: (value) => value || <Text type="secondary">{t('common.none')}</Text>,
    },
    {
      title: t('common.enabled'),
      key: 'enabled',
      render: (_, record) => (
        <Switch
          size="small"
          checked={record.enabled}
          onChange={() => api.put(`/model-configs/${record.id}/toggle`).then(loadConfigs).catch(() => message.error(t('common.failed')))}
        />
      ),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      render: (_, record) => (
        <Space size="small">
          <Button size="small" type="primary" icon={<SettingOutlined />} onClick={() => openEditor(record.id)}>
            {t('modelConfig.open')}
          </Button>
          <Button
            size="small"
            icon={<CopyOutlined />}
            onClick={() => {
              setCloneTarget(record);
              cloneForm.setFieldsValue({ name: `${record.name} Copy` });
            }}
          >
            {t('common.clone')}
          </Button>
          <Popconfirm
            title={t('common.confirm_delete')}
            onConfirm={() => api.delete(`/model-configs/${record.id}`).then(() => {
              message.success(t('common.deleted'));
              loadConfigs();
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
          {t('menu.modelConfigs')}
        </Typography.Title>
      </div>

      <Alert
        style={{ marginBottom: 24, borderRadius: 8, border: 'none', background: '#eef2ff' }}
        type="info"
        showIcon
        message={<Text strong style={{ color: '#4f46e5' }}>{t('modelConfig.summary', { providers: providerCount, models: totalModels })}</Text>}
      />

      <Card className="premium-card">
        <div style={{ display: 'flex', justifyContent: 'flex-start', marginBottom: 16 }}>
          <Button
            type="primary"
            size="large"
            style={{ borderRadius: 8 }}
            icon={<PlusOutlined />}
            onClick={() => {
              createForm.resetFields();
              setCreateOpen(true);
            }}
          >
            {t('modelConfig.add')}
          </Button>
        </div>

        <Table
          dataSource={configs}
          columns={columns}
          rowKey="id"
          loading={loading}
          size="middle"
          pagination={{
            showSizeChanger: true,
            showTotal: (total) => `Total ${total} configs`,
          }}
          locale={{ emptyText: <Empty description={t('modelConfig.empty')} /> }}
        />
      </Card>

      <Modal
        title={t('modelConfig.add')}
        open={createOpen}
        onOk={handleCreate}
        onCancel={() => setCreateOpen(false)}
        destroyOnClose
      >
        <Form form={createForm} layout="vertical">
          <Form.Item name="name" label={t('modelConfig.name')} rules={[{ required: true }]}>
            <Input placeholder="Production Config" />
          </Form.Item>
          <Form.Item name="description" label={t('modelConfig.description')}>
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={t('modelConfig.clone')}
        open={!!cloneTarget}
        onOk={handleClone}
        onCancel={() => {
          setCloneTarget(null);
          cloneForm.resetFields();
        }}
        destroyOnClose
      >
        <Form form={cloneForm} layout="vertical">
          <Form.Item name="name" label={t('modelConfig.cloneName')} rules={[{ required: true }]}>
            <Input />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        title={editingConfig?.name || t('modelConfig.editor')}
        open={editorOpen}
        onClose={() => {
          setEditorOpen(false);
          setEditingConfig(null);
          setSelected({});
        }}
        width={880}
        extra={(
          <Button type="primary" icon={<SaveOutlined />} loading={saving} onClick={handleSave} disabled={!editingConfig}>
            {t('modelConfig.save')}
          </Button>
        )}
      >
        {editorLoading ? (
          <div style={{ textAlign: 'center', padding: 24 }}><Spin /></div>
        ) : !editingConfig ? (
          <Empty description={t('modelConfig.empty')} />
        ) : (
          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            <Form form={editorForm} layout="vertical">
              <Form.Item name="name" label={t('modelConfig.name')} rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item name="description" label={t('modelConfig.description')}>
                <Input.TextArea rows={3} />
              </Form.Item>
            </Form>

            <Alert
              type="info"
              showIcon
              message={t('modelConfig.selected', { total: Object.keys(selected).length })}
              description={t('modelConfig.selectHint')}
            />

            {/* 一键映射功能 */}
            <div style={{ border: '1px solid #d9d9d9', borderRadius: 8, padding: 16 }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: quickMapOpen ? 12 : 0 }}>
                <Button
                  type={quickMapOpen ? 'primary' : 'default'}
                  icon={<ThunderboltOutlined />}
                  onClick={() => setQuickMapOpen(!quickMapOpen)}
                  ghost={quickMapOpen}
                >
                  {t('modelConfig.quickMap')}
                </Button>
              </div>
              {quickMapOpen && (
                <div>
                  <Space direction="vertical" style={{ width: '100%' }} size={12}>
                    <Input
                      value={quickMapPrefix}
                      onChange={(e) => setQuickMapPrefix(e.target.value)}
                      placeholder={t('modelConfig.quickMap.prefixHint')}
                      addonBefore={t('modelConfig.quickMap.prefix')}
                      style={{ maxWidth: 500 }}
                    />
                    <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                      <Button type="primary" icon={<ThunderboltOutlined />} onClick={applyQuickMap}>
                        {t('modelConfig.quickMap.apply')}
                      </Button>
                      <Text type="secondary" style={{ fontSize: 12 }}>
                        {t('modelConfig.quickMap.selectAll')}
                      </Text>
                    </div>
                  </Space>
                </div>
              )}
            </div>

            <Divider style={{ margin: '8px 0' }} />

            {!availableModels.length ? (
              <Empty description={t('modelConfig.noModels')} />
            ) : (
              <Collapse items={collapseItems} />
            )}
          </Space>
        )}
      </Drawer>
    </div>
  );
}

