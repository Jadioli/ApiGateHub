import { Layout as AntLayout, Menu, Dropdown, Button } from 'antd';
import {
  ApartmentOutlined,
  CloudServerOutlined,
  DashboardOutlined,
  FileTextOutlined,
  GlobalOutlined,
  KeyOutlined,
  LogoutOutlined,
} from '@ant-design/icons';
import { useNavigate, useLocation, Outlet } from 'react-router-dom';
import { useI18n } from '../i18n';

const { Sider, Content, Header } = AntLayout;

export default function Layout() {
  const navigate = useNavigate();
  const location = useLocation();
  const { t, lang, switchLang } = useI18n();

  const menuItems = [
    { key: '/', icon: <DashboardOutlined />, label: t('menu.dashboard') },
    { key: '/providers', icon: <CloudServerOutlined />, label: t('menu.providers') },
    { key: '/model-configs', icon: <ApartmentOutlined />, label: t('menu.modelConfigs') },
    { key: '/apikeys', icon: <KeyOutlined />, label: t('menu.apikeys') },
    { key: '/logs', icon: <FileTextOutlined />, label: t('menu.logs') },
  ];

  const selected = menuItems
    .filter((item) => item.key !== '/' && location.pathname.startsWith(item.key))
    .map((item) => item.key);
  const selectedKeys = selected.length ? selected : ['/'];

  const langMenu = {
    items: [
      { key: 'zh', label: '中文' },
      { key: 'en', label: 'English' },
    ],
    onClick: ({ key }) => switchLang(key),
  };

  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Sider breakpoint="lg" collapsedWidth={64}>
        <div style={{ color: '#fff', textAlign: 'center', padding: '16px 0', fontSize: 20, fontWeight: 700 }}>
          {t('app.title')}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={selectedKeys}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
        <div style={{ position: 'absolute', bottom: 16, width: '100%', textAlign: 'center' }}>
          <LogoutOutlined
            style={{ color: '#fff', fontSize: 18, cursor: 'pointer' }}
            onClick={() => { localStorage.removeItem('token'); navigate('/login'); }}
            title="Logout"
          />
        </div>
      </Sider>
      <AntLayout>
        <Header style={{ background: '#fff', padding: '0 24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span style={{ fontSize: 16, fontWeight: 600 }}>
            {menuItems.find((item) => selectedKeys.includes(item.key))?.label || t('app.title')}
          </span>
          <Dropdown menu={langMenu}>
            <Button type="text" icon={<GlobalOutlined />}>{lang === 'zh' ? '中文' : 'EN'}</Button>
          </Dropdown>
        </Header>
        <Content style={{ margin: 24 }}>
          <Outlet />
        </Content>
      </AntLayout>
    </AntLayout>
  );
}

