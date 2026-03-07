import { useEffect, useState } from 'react';
import { Table, Input, Select, Space, Tag } from 'antd';
import { useI18n } from '../i18n';
import api from '../api';

export default function Logs() {
  const [data, setData] = useState({ logs: [], total: 0 });
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [filters, setFilters] = useState({});
  const { t } = useI18n();

  const load = (p = page, f = filters) => {
    setLoading(true);
    api.get('/logs', { params: { page: p, page_size: 20, ...f } })
      .then((r) => setData(r.data))
      .catch(() => {})
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    const timer = setTimeout(() => { load(); }, 0);
    return () => clearTimeout(timer);
  }, []);

  const onPageChange = (p) => { setPage(p); load(p); };
  const onFilter = (key, value) => { const f = { ...filters, [key]: value || undefined }; setFilters(f); setPage(1); load(1, f); };

  const statusColor = (c) => c >= 200 && c < 300 ? 'green' : c >= 400 && c < 500 ? 'orange' : 'red';

  const columns = [
    { title: t('log.time'), dataIndex: 'created_at', key: 'time', width: 180, render: (v) => new Date(v).toLocaleString() },
    { title: t('log.request_model'), dataIndex: 'request_model', key: 'req' },
    { title: t('log.provider_model'), dataIndex: 'provider_model', key: 'prov' },
    { title: t('log.status'), dataIndex: 'status_code', key: 'status', width: 80, render: (v) => <Tag color={statusColor(v)}>{v}</Tag> },
    { title: t('log.latency'), dataIndex: 'response_time', key: 'latency', width: 100, render: (v) => `${v}ms` },
    { title: t('log.prompt'), dataIndex: 'tokens_prompt', key: 'prompt', width: 120 },
    { title: t('log.completion'), dataIndex: 'tokens_completion', key: 'comp', width: 140 },
    { title: t('log.error'), dataIndex: 'error', key: 'error', ellipsis: true, render: (v) => v ? <Tag color="red">{v}</Tag> : '-' },
  ];

  return (
    <>
      <Space style={{ marginBottom: 16 }}>
        <Input placeholder={t('log.filter.model')} allowClear onChange={(e) => onFilter('model', e.target.value)} style={{ width: 200 }} />
        <Select placeholder={t('log.filter.status')} allowClear onChange={(v) => onFilter('status', v)} style={{ width: 140 }}
          options={[{ value: '200', label: '200' }, { value: '400', label: '400' }, { value: '401', label: '401' }, { value: '403', label: '403' }, { value: '500', label: '500' }, { value: '502', label: '502' }]}
        />
      </Space>
      <Table dataSource={data.logs} columns={columns} rowKey="id" loading={loading} size="middle"
        pagination={{ current: page, total: data.total, pageSize: 20, onChange: onPageChange, showTotal: (total) => t('log.total', { total }) }}
      />
    </>
  );
}

