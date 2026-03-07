import { useEffect, useState } from 'react';
import { Card, Col, Row, Statistic } from 'antd';
import { CloudServerOutlined, KeyOutlined, ThunderboltOutlined } from '@ant-design/icons';
import { useI18n } from '../i18n';
import api from '../api';

export default function Dashboard() {
  const [stats, setStats] = useState({});
  const { t } = useI18n();

  useEffect(() => {
    api.get('/dashboard').then((r) => setStats(r.data)).catch(() => {});
  }, []);

  const items = [
    { title: t('dashboard.providers'), value: stats.provider_count, icon: <CloudServerOutlined />, color: '#1677ff' },
    { title: t('dashboard.apikeys'), value: stats.api_key_count, icon: <KeyOutlined />, color: '#faad14' },
    { title: t('dashboard.today'), value: stats.today_requests, icon: <ThunderboltOutlined />, color: '#ff4d4f' },
  ];

  return (
    <Row gutter={[16, 16]}>
      {items.map((item) => (
        <Col xs={24} sm={12} lg={8} key={item.title}>
          <Card>
            <Statistic title={item.title} value={item.value ?? 0} prefix={item.icon} valueStyle={{ color: item.color }} />
          </Card>
        </Col>
      ))}
    </Row>
  );
}
