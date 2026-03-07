import { useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Alert,
  Badge,
  Button,
  Card,
  Collapse,
  Empty,
  Select,
  Space,
  Spin,
  Tag,
  Typography,
  message,
} from 'antd';
import { ArrowLeftOutlined, CopyOutlined, SaveOutlined } from '@ant-design/icons';
import { useI18n } from '../i18n';
import api from '../api';

const { Title, Text } = Typography;

function groupItemsByProvider(items) {
  const groups = new Map();

  items.forEach((item) => {
    const providerId = item.provider?.id || item.provider_id;
    if (!groups.has(providerId)) {
      groups.set(providerId, {
        key: providerId,
        provider: item.provider,
        items: [],
      });
    }
    groups.get(providerId).items.push(item);
  });

  return Array.from(groups.values());
}

export default function APIKeyDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { t } = useI18n();

  const [keyInfo, setKeyInfo] = useState(null);
  const [configs, setConfigs] = useState([]);
  const [selectedConfigId, setSelectedConfigId] = useState();
  const [previewConfig, setPreviewConfig] = useState(null);
  const [loading, setLoading] = useState(false);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  const load = async () => {
    setLoading(true);
    try {
      const [keyRes, configRes] = await Promise.all([
        api.get(`/apikeys/${id}`),
        api.get('/model-configs'),
      ]);
      setKeyInfo(keyRes.data);
      setConfigs(configRes.data || []);
      setSelectedConfigId(keyRes.data.model_config_id ?? undefined);
    } catch {
      message.error(t('common.failed'));
    } finally {
      setLoading(false);
    }
  };

  const loadPreview = async (configId) => {
    if (!configId) {
      setPreviewConfig(null);
      return;
    }

    setPreviewLoading(true);
    try {
      const { data } = await api.get(`/model-configs/${configId}`);
      setPreviewConfig(data);
    } catch {
      message.error(t('common.failed'));
    } finally {
      setPreviewLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, [id]);

  useEffect(() => {
    loadPreview(selectedConfigId);
  }, [selectedConfigId]);

  const handleSave = async () => {
    setSaving(true);
    try {
      await api.put(`/apikeys/${id}/model-config`, {
        model_config_id: selectedConfigId ?? null,
      });
      message.success(t('common.success'));
      load();
    } catch (error) {
      message.error(error.response?.data?.error || t('common.failed'));
    } finally {
      setSaving(false);
    }
  };

  const configOptions = useMemo(
    () => configs.map((config) => ({
      value: config.id,
      label: `${config.name}${config.enabled ? '' : ' (disabled)'}`,
    })),
    [configs],
  );

  const previewGroups = useMemo(
    () => groupItemsByProvider(previewConfig?.items || []).map((group) => ({
      key: group.key,
      label: (
        <Space>
          <Tag color={group.provider?.protocol === 'openai' ? 'blue' : 'orange'}>{group.provider?.protocol || 'provider'}</Tag>
          <span>{group.provider?.name || t('common.none')}</span>
          <Badge count={group.items.length} style={{ backgroundColor: '#1677ff' }} />
        </Space>
      ),
      children: (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          {group.items.map((item) => (
            <div key={item.id} style={{ border: '1px solid #f0f0f0', borderRadius: 8, padding: 12 }}>
              <Space wrap>
                <Tag>{item.provider_model?.model_name || item.provider_model_id}</Tag>
                <Text type="secondary">{item.mapped_name}</Text>
                <Tag color="geekblue">P{item.priority || 0}</Tag>
                {!item.enabled && <Tag>{t('common.pending')}</Tag>}
              </Space>
            </div>
          ))}
        </div>
      ),
    })),
    [previewConfig, t],
  );

  return (
    <Spin spinning={loading}>
      <div>
        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Space>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/apikeys')}>{t('common.back')}</Button>
            <Title level={4} style={{ margin: 0 }}>{keyInfo?.name}</Title>
          </Space>
          <Space>
            <code style={{ fontSize: 13, background: '#f5f5f5', padding: '2px 8px', borderRadius: 4 }}>{keyInfo?.key}</code>
            <Button
              size="small"
              type="text"
              icon={<CopyOutlined />}
              onClick={() => {
                navigator.clipboard.writeText(keyInfo?.key || '');
                message.success(t('apikey.copied'));
              }}
            />
          </Space>
        </div>

        <Card
          title={t('apikey.config')}
          extra={(
            <Button type="primary" icon={<SaveOutlined />} loading={saving} onClick={handleSave}>
              {t('common.save')}
            </Button>
          )}
          style={{ marginBottom: 16 }}
        >
          {!configs.length && (
            <Alert
              style={{ marginBottom: 16 }}
              type="warning"
              showIcon
              message={t('apikey.config.required')}
              action={<Button size="small" onClick={() => navigate('/model-configs')}>{t('apikey.config.openConfigs')}</Button>}
            />
          )}

          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            <div>
              <Text strong>{t('apikey.config.assigned')}</Text>
              <Select
                allowClear
                style={{ width: '100%', marginTop: 8 }}
                placeholder={t('apikey.config.placeholder')}
                options={configOptions}
                value={selectedConfigId}
                onChange={(value) => setSelectedConfigId(value)}
              />
            </div>
            <Button onClick={() => navigate('/model-configs')}>{t('apikey.config.openConfigs')}</Button>
          </Space>
        </Card>

        <Card
          title={(
            <Space>
              <span>{t('apikey.config.preview')}</span>
              <Badge count={previewConfig?.items?.length || 0} showZero style={{ backgroundColor: '#1677ff' }} />
            </Space>
          )}
        >
          {previewLoading ? (
            <div style={{ textAlign: 'center', padding: 24 }}><Spin /></div>
          ) : !selectedConfigId ? (
            <Empty description={t('apikey.config.none')} />
          ) : !previewConfig?.items?.length ? (
            <Empty description={t('apikey.config.noItems')} />
          ) : (
            <Collapse items={previewGroups} />
          )}
        </Card>
      </div>
    </Spin>
  );
}

