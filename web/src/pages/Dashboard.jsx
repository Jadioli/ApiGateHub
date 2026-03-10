import { useEffect, useState } from 'react';
import { Card, Col, Row, Typography } from 'antd';
import { CloudServerOutlined, KeyOutlined, ThunderboltOutlined } from '@ant-design/icons';
import { useI18n } from '../i18n';
import api from '../api';
import './Dashboard.css';

const { Title } = Typography;

export default function Dashboard() {
  const [stats, setStats] = useState({});
  const { t } = useI18n();

  useEffect(() => {
    api.get('/dashboard').then((r) => setStats(r.data)).catch(() => { });
  }, []);

  const items = [
    {
      title: t('dashboard.providers'),
      value: stats.provider_count,
      icon: <CloudServerOutlined />,
      gradient: 'linear-gradient(135deg, #3b82f6, #60a5fa)',
      shadowColor: 'rgba(59, 130, 246, 0.4)'
    },
    {
      title: t('dashboard.apikeys'),
      value: stats.api_key_count,
      icon: <KeyOutlined />,
      gradient: 'linear-gradient(135deg, #f59e0b, #fbbf24)',
      shadowColor: 'rgba(245, 158, 11, 0.4)'
    },
    {
      title: t('dashboard.today'),
      value: stats.today_requests,
      icon: <ThunderboltOutlined />,
      gradient: 'linear-gradient(135deg, #ef4444, #f87171)',
      shadowColor: 'rgba(239, 68, 68, 0.4)'
    },
  ];

  return (
    <div className="dashboard-container">
      <div className="dashboard-header mb-6">
        <Title level={2} style={{ marginTop: 0, marginBottom: '24px', fontWeight: 700, color: '#1e293b' }}>
          {t('menu.dashboard')}
        </Title>
      </div>
      <Row gutter={[24, 24]}>
        {items.map((item) => (
          <Col xs={24} sm={12} lg={8} key={item.title}>
            <Card className="premium-card hover-scale dashboard-stat-card">
              <div className="stat-content">
                <div
                  className="stat-icon-wrapper"
                  style={{
                    background: item.gradient,
                    boxShadow: `0 8px 16px ${item.shadowColor}`
                  }}
                >
                  {item.icon}
                </div>
                <div className="stat-info">
                  <div className="stat-title">{item.title}</div>
                  <div className="stat-value">{item.value ?? 0}</div>
                </div>
              </div>
            </Card>
          </Col>
        ))}
      </Row>
    </div>
  );
}
