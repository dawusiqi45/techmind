import { useEffect } from 'react'
import { Outlet, useNavigate, Link } from 'react-router-dom'
import { ConfigProvider, Input, Button, Avatar, Dropdown, theme } from 'antd'
import type { MenuProps } from 'antd'
import { UserOutlined } from '@ant-design/icons'
import { useAuthStore } from '@/store/auth'
import { useThemeStore } from '@/store/theme'
import styles from './ForumLayout.module.css'

export default function ForumLayout() {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)
  const setTheme = useThemeStore((s) => s.setTheme)

  useEffect(() => {
    setTheme('forum')
    document.body.setAttribute('data-theme', 'forum')
  }, [setTheme])

  const handleSearch = (value: string) => {
    if (value.trim()) {
      navigate(`/search?q=${encodeURIComponent(value.trim())}`)
    }
  }

  const handleLogout = () => {
    logout()
    navigate('/')
  }

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      label: '个人主页',
      onClick: () => navigate('/user/profile'),
    },
    {
      key: 'logout',
      label: '退出登录',
      onClick: handleLogout,
    },
  ]

  return (
    <ConfigProvider
      theme={{
        algorithm: theme.defaultAlgorithm,
        token: { colorPrimary: '#1677ff' },
      }}
    >
      <div className={styles.header}>
        <div className={styles.headerInner}>
          <Link to="/" className={styles.logo}>TechMind</Link>
          <div className={styles.search}>
            <Input.Search
              placeholder="搜索文章..."
              onSearch={handleSearch}
              style={{ maxWidth: 480 }}
            />
          </div>
          <div className={styles.actions}>
            {user ? (
              <>
                <Button type="primary" onClick={() => navigate('/articles/new')}>
                  写文章
                </Button>
                <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
                  <Avatar
                    src={user.avatar || undefined}
                    icon={!user.avatar ? <UserOutlined /> : undefined}
                    style={{ cursor: 'pointer', marginLeft: 8 }}
                  />
                </Dropdown>
              </>
            ) : (
              <Button type="primary" onClick={() => navigate('/login')}>
                登录
              </Button>
            )}
          </div>
        </div>
      </div>
      <main className={styles.main}>
        <Outlet />
      </main>
    </ConfigProvider>
  )
}
