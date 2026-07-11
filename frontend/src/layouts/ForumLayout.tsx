import { useEffect, useRef, useState } from 'react'
import { Outlet, useNavigate, Link } from 'react-router-dom'
import { Input, Avatar, Dropdown } from 'antd'
import type { MenuProps } from 'antd'
import { SearchOutlined, EditOutlined, UserOutlined, LogoutOutlined, SettingOutlined, HeartOutlined, StarOutlined, HomeOutlined } from '@ant-design/icons'
import { useAuthStore } from '@/store/auth'
import { useLoginModal } from '@/store/loginModal'
import styles from './ForumLayout.module.css'

export default function ForumLayout() {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)
  const openLoginModal = useLoginModal((s) => s.open)
  const [scrolled, setScrolled] = useState(false)
  const [searchExpanded, setSearchExpanded] = useState(false)
  const searchRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 8)
    window.addEventListener('scroll', handleScroll, { passive: true })
    return () => window.removeEventListener('scroll', handleScroll)
  }, [])

  const handleSearch = (value: string) => {
    if (value.trim()) {
      navigate(`/search?q=${encodeURIComponent(value.trim())}`)
      setSearchExpanded(false)
    }
  }

  const handleLogout = () => {
    logout()
    navigate('/')
  }

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <SettingOutlined />,
      label: '个人主页',
      onClick: () => navigate('/user/profile'),
    },
    {
      key: 'favorites',
      icon: <StarOutlined />,
      label: '我的收藏',
      onClick: () => navigate('/user/profile?tab=favorites'),
    },
    {
      key: 'likes',
      icon: <HeartOutlined />,
      label: '我的点赞',
      onClick: () => navigate('/user/profile?tab=likes'),
    },
    { type: 'divider' },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
      danger: true,
    },
  ]

  return (
    <div className={styles.forumShell}>
      <header className={`${styles.header} ${scrolled ? styles.scrolled : ''}`}>
        <div className={styles.inner}>
          <Link to="/" className={styles.logo}>
            <span className={styles.logoIcon}>⬡</span>
            TechMind
          </Link>

          <Link to="/" className={styles.homeLink}>
            <HomeOutlined />
            <span>首页</span>
          </Link>

          <div
            ref={searchRef}
            className={`${styles.searchWrap} ${searchExpanded ? styles.searchExpanded : ''}`}
          >
            <Input
              prefix={<SearchOutlined style={{ color: 'var(--text-3)' }} />}
              placeholder="搜索文章、话题..."
              variant="filled"
              onPressEnter={(e) => handleSearch((e.target as HTMLInputElement).value)}
              onFocus={() => setSearchExpanded(true)}
              onBlur={() => setSearchExpanded(false)}
              className={styles.searchInput}
            />
          </div>

          <div className={styles.actions}>
            {user ? (
              <>
                <button
                  className={styles.writeBtn}
                  onClick={() => navigate('/articles/new')}
                >
                  <EditOutlined />
                  <span>写文章</span>
                </button>
                <Dropdown
                  menu={{ items: userMenuItems }}
                  placement="bottomRight"
                  trigger={['click']}
                >
                  <button className={styles.avatarBtn}>
                    <Avatar
                      src={user.avatar || undefined}
                      icon={!user.avatar ? <UserOutlined /> : undefined}
                      size={32}
                      style={{ background: 'var(--accent)' }}
                    />
                    <span className={styles.username}>{user.username}</span>
                  </button>
                </Dropdown>
              </>
            ) : (
              <button
                className={styles.loginBtn}
                onClick={openLoginModal}
              >
                登录 / 注册
              </button>
            )}
          </div>
        </div>
      </header>

      <main className={styles.main}>
        <Outlet />
      </main>
    </div>
  )
}
