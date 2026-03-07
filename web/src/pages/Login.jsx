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
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#f0f2f5' }}>
      <Card title={t('login.title')} style={{ width: 380 }}>
        <Form onFinish={onFinish} size="large">
          <Form.Item name="password" rules={[{ required: true }]}>
            <Input.Password prefix={<LockOutlined />} placeholder={t('login.password')} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block>{t('login.submit')}</Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
