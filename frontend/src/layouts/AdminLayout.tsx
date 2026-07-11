import { useEffect } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Avatar, Dropdown } from 'antd'
import type { MenuProps } from 'antd'
import {
  DashboardOutlined,
  AlertOutlined,
  ToolOutlined,
  BookOutlined,
  DeploymentUnitOutlined,
  UserOutlined,
  LogoutOutlined,
  ThunderboltOutlined,
  ClockCircleOutlined,
  WarningOutlined,
  ApiOutlined,
  RobotOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '@/store/auth'
import styles from './AdminLayout.module.css'

interface NavItem {
  path: string
  label: string
  icon: React.ReactNode
  children?: { path: string; label: string; icon: React.ReactNode }[]
}

const navItems: NavItem[] = [
  {
    path: '/admin/monitor',
    label: '监控总览',
    icon: <DashboardOutlined />,
    children: [
      { path: '/admin/monitor', label: '总览', icon: <DashboardOutlined /> },
      { path: '/admin/monitor/slow', label: '慢请求', icon: <ClockCircleOutlined /> },
      { path: '/admin/monitor/errors', label: '错误事件', icon: <WarningOutlined /> },
      { path: '/admin/monitor/queues', label: '队列状态', icon: <ApiOutlined /> },
      { path: '/admin/monitor/ai', label: 'AI 调用', icon: <RobotOutlined /> },
    ],
  },
  { path: '/admin/alerts', label: '告警中心', icon: <AlertOutlined /> },
  {
    path: '/admin/ops/reports',
    label: 'SRE 诊断',
    icon: <ToolOutlined />,
	  children: [
	    { path: '/admin/ops/reports', label: '诊断报告', icon: <ToolOutlined /> },
	    { path: '/admin/ops/diagnose', label: '手动触发', icon: <ThunderboltOutlined /> },
	    { path: '/admin/incidents', label: '故障事件', icon: <AlertOutlined /> },
	  ],
  },
  { path: '/admin/runbooks', label: 'Runbook', icon: <BookOutlined /> },
  { path: '/admin/deployments', label: '部署变更', icon: <DeploymentUnitOutlined /> },
]

export default function AdminLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)

  useEffect(() => {
    document.body.style.background = 'var(--bg)'
  }, [])

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
      danger: true,
    },
  ]

  const isActive = (path: string) => location.pathname === path
  const isGroupActive = (item: NavItem) =>
    item.children ? item.children.some((c) => isActive(c.path)) : isActive(item.path)

  return (
    <div className={styles.layout}>
      <aside className={styles.sider}>
        <div className={styles.siderHeader}>
          <span className={styles.siderLogo}>⬡ TechMind</span>
          <span className={styles.siderBadge}>Admin</span>
        </div>

        <nav className={styles.nav}>
          {navItems.map((item) => (
            <div key={item.path}>
              {item.children ? (
                <div className={`${styles.navGroup} ${isGroupActive(item) ? styles.navGroupActive : ''}`}>
                  <div className={styles.navGroupLabel}>
                    {item.icon}
                    <span>{item.label}</span>
                  </div>
                  <div className={styles.navGroupChildren}>
                    {item.children.map((child) => (
                      <button
                        key={child.path}
                        className={`${styles.navItem} ${styles.navChild} ${isActive(child.path) ? styles.navActive : ''}`}
                        onClick={() => navigate(child.path)}
                      >
                        {child.icon}
                        <span>{child.label}</span>
                      </button>
                    ))}
                  </div>
                </div>
              ) : (
                <button
                  className={`${styles.navItem} ${isActive(item.path) ? styles.navActive : ''}`}
                  onClick={() => navigate(item.path)}
                >
                  {item.icon}
                  <span>{item.label}</span>
                </button>
              )}
            </div>
          ))}
        </nav>
      </aside>

      <div className={styles.body}>
        <header className={styles.topbar}>
          <div className={styles.topbarTitle}>
            {location.pathname.split('/').pop()?.replace(/-/g, ' ')}
          </div>
          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight" trigger={['click']}>
            <button className={styles.userBtn}>
              <Avatar
                src={user?.avatar || undefined}
                icon={!user?.avatar ? <UserOutlined /> : undefined}
                size={28}
                style={{ background: 'var(--accent)' }}
              />
              <span className={styles.userName}>{user?.username}</span>
            </button>
          </Dropdown>
        </header>

        <main className={styles.content}>
          <Outlet />
        </main>
      </div>
    </div>
  )
}
