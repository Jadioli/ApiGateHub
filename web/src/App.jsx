import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import { I18nProvider } from './i18n';
import Layout from './components/Layout';
import ProtectedRoute from './components/ProtectedRoute';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Providers from './pages/Providers';
import ModelConfigs from './pages/ModelConfigs';
import APIKeys from './pages/APIKeys';
import APIKeyDetail from './pages/APIKeyDetail';
import Logs from './pages/Logs';

export default function App() {
  return (
    <I18nProvider>
      <ConfigProvider
        theme={{
          token: {
            colorPrimary: '#6366f1',
            borderRadius: 8,
            fontFamily: "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
            colorBgContainer: '#ffffff',
            wireframe: false,
          },
          components: {
            Card: {
              borderRadiusLG: 16,
              boxShadowTertiary: '0 10px 25px -5px rgba(0, 0, 0, 0.05), 0 8px 10px -6px rgba(0, 0, 0, 0.01)',
            },
            Table: {
              borderRadiusLG: 12,
              headerBg: '#f8fafc',
            },
            Layout: {
              headerBg: 'rgba(255, 255, 255, 0.8)',
            }
          }
        }}
      >
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route
              element={(
                <ProtectedRoute>
                  <Layout />
                </ProtectedRoute>
              )}
            >
              <Route path="/" element={<Dashboard />} />
              <Route path="/providers" element={<Providers />} />
              <Route path="/model-configs" element={<ModelConfigs />} />
              <Route path="/apikeys" element={<APIKeys />} />
              <Route path="/apikeys/:id" element={<APIKeyDetail />} />
              <Route path="/logs" element={<Logs />} />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </ConfigProvider>
    </I18nProvider>
  );
}

