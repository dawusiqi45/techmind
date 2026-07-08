import { useEffect } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { ConfigProvider, Layout, Menu, Avatar, Dropdown, theme } from 'antd'
import type { MenuProps } from 'antd'
import {
  DashboardOutlined,
  AlertOutlined,
  ToolOutlined,
  BookOutlined,
  DeploymentUnitOutlined,
  UserOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '@/store/auth'
import { useThemeStore } from '@/store/theme'
import styles from './AdminLayout.module.css'

const { Sider, Header, Content } = Layout

export default function AdminLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)
  const setTheme = useThemeStore((s) => s.setTheme)

  useEffect(() => {
    setTheme('admin')
    document.body.setAttribute('data-theme', 'admin')
  }, [setTheme])

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'logout',
      label: '退出登录',
      onClick: handleLogout,
    },
  ]

  const menuItems: MenuProps['items'] = [
    {
      key: 'monitor',
      icon: <DashboardOutlined />,
      label: '监控',
      children: [
        { key: '/admin/monitor', label: '总览', onClick: () => navigate('/admin/monitor') },
        { key: '/admin/monitor/slow', label: '慢请求', onClick: () => navigate('/admin/monitor/slow') },
        { key: '/admin/monitor/errors', label: '错误事件', onClick: () => navigate('/admin/monitor/errors') },
        { key: '/admin/monitor/queues', label: '队列状态', onClick: () => navigate('/admin/monitor/queues') },
        { key: '/admin/monitor/ai', label: 'AI调用', onClick: () => navigate('/admin/monitor/ai') },
      ],
    },
    {
      key: '/admin/alerts',
      icon: <AlertOutlined />,
      label: '告警中心',
      onClick: () => navigate('/admin/alerts'),
    },
    {
      key: 'ops',
      icon: <ToolOutlined />,
      label: 'SRE 诊断',
      children: [
        { key: '/admin/ops/reports', label: '诊断报告', onClick: () => navigate('/admin/ops/reports') },
        { key: '/admin/ops/diagnose', label: '手动触发', onClick: () => navigate('/admin/ops/diagnose') },
      ],
    },
    {
      key: '/admin/runbooks',
      icon: <BookOutlined />,
      label: 'Runbook',
      onClick: () => navigate('/admin/runbooks'),
    },
    {
      key: '/admin/deployments',
      icon: <DeploymentUnitOutlined />,
      label: '部署变更',
      onClick: () => navigate('/admin/deployments'),
    },
  ]

  return (
    <ConfigProvider
      theme={{
        algorithm: theme.darkAlgorithm,
        token: {
          colorPrimary: '#177ddc',
          colorBgContainer: '#161b22',
          colorBgLayout: '#0f1117',
        },
      }}
    >
      <Layout style={{ minHeight: '100vh' }}>
        <Sider
          width={220}
          style={{
            background: '#161b22',
            borderRight: '1px solid #30363d',
            position: 'fixed',
            height: '100vh',
            left: 0,
            top: 0,
            zIndex: 100,
            overflow: 'auto',
          }}
        >
          <div className={styles.siderLogo}>TechMind Admin</div>
          <Menu
            mode="inline"
            theme="dark"
            selectedKeys={[location.pathname]}
            defaultOpenKeys={['monitor', 'ops']}
            items={menuItems}
            style={{ background: '#161b22', borderRight: 'none' }}
          />
        </Sider>
        <Layout style={{ marginLeft: 220, background: '#0f1117' }}>
          <Header
            style={{
              background: '#161b22',
              borderBottom: '1px solid #30363d',
              height: 56,
              padding: '0 24px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'flex-end',
              position: 'sticky',
              top: 0,
              zIndex: 99,
            }}
          >
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, cursor: 'pointer', color: '#e6edf3' }}>
                <Avatar
                  src={user?.avatar || undefined}
                  icon={!user?.avatar ? <UserOutlined /> : undefined}
                  size="small"
                />
                <span>{user?.username}</span>
              </div>
            </Dropdown>
          </Header>
          <Content style={{ padding: 24, background: '#0f1117', minHeight: 'calc(100vh - 56px)' }}>
            <Outlet />
          </Content>
        </Layout>
      </Layout>
    </ConfigProvider>
  )
}
