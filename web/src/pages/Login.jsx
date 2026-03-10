import { useState } from 'react';
import { Form, Input, Button, Card, message } from 'antd';
import { LockOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useI18n } from '../i18n';
import axios from 'axios';

export default function Login() {
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { t } = useI18n();

  const onFinish = async (values) => {
    setLoading(true);
    try {
      const { data } = await axios.post('/admin/login', values);
      localStorage.setItem('token', data.token);
      navigate('/');
    } catch {
      message.error(t('login.failed'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #e0e7ff 0%, #ede9fe 100%)'
    }}>
      <Card
        className="premium-card"
        style={{ width: 420, padding: 24 }}
        bodyStyle={{ padding: 0 }}
      >
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <div style={{
            width: 48,
            height: 48,
            background: 'linear-gradient(135deg, var(--primary-color), #818cf8)',
            borderRadius: 12,
            margin: '0 auto 16px',
            boxShadow: '0 8px 16px rgba(99, 102, 241, 0.3)'
          }} />
          <h2 style={{ margin: 0, color: '#1e293b', fontSize: 24, fontWeight: 700 }}>
            {t('app.title')}
          </h2>
          <p style={{ margin: '8px 0 0', color: '#64748b' }}>
            {t('login.title')}
          </p>
        </div>

        <Form onFinish={onFinish} layout="vertical" size="large">
          <Form.Item name="password" rules={[{ required: true, message: t('login.password') }]}>
            <Input.Password
              prefix={<LockOutlined style={{ color: '#94a3b8' }} />}
              placeholder={t('login.password')}
              style={{ borderRadius: 8 }}
            />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0, marginTop: 24 }}>
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              block
              style={{ height: 44, fontSize: 16, fontWeight: 500 }}
            >
              {t('login.submit')}
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
