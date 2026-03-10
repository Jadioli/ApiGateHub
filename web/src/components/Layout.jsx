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
import './Layout.css'; // Will create this

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
    <AntLayout style={{ minHeight: '100vh', flexDirection: 'row', padding: '16px', gap: '16px' }}>
      <Sider
        breakpoint="lg"
        collapsedWidth={64}
        className="premium-sider"
        width={240}
      >
        <div className="sider-header">
          <div className="sider-logo" />
          <span className="sider-title">{t('app.title')}</span>
        </div>
        <Menu
          mode="inline"
          selectedKeys={selectedKeys}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          className="premium-menu"
        />
        <div className="sider-footer">
          <Button
            type="text"
            icon={<LogoutOutlined />}
            onClick={() => { localStorage.removeItem('token'); navigate('/login'); }}
            className="logout-btn"
            block
          >
            Logout
          </Button>
        </div>
      </Sider>

      <AntLayout className="main-layout fade-in">
        <Header className="premium-header glass-panel">
          <span className="header-title">
            {menuItems.find((item) => selectedKeys.includes(item.key))?.label || t('app.title')}
          </span>
          <div className="header-actions">
            <Dropdown menu={langMenu} placement="bottomRight">
              <Button type="text" className="lang-btn" icon={<GlobalOutlined />}>
                {lang === 'zh' ? '中文' : 'EN'}
              </Button>
            </Dropdown>
          </div>
        </Header>
        <Content className="premium-content slide-up">
          <Outlet />
        </Content>
      </AntLayout>
    </AntLayout>
  );
}

