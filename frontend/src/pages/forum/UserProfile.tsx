import { useEffect, useState } from 'react'
import { Tabs, Input, Spin, message, Upload } from 'antd'
import { MailOutlined, CalendarOutlined, CameraOutlined, SaveOutlined } from '@ant-design/icons'
import { authApi } from '@/api/auth'
import { articleApi } from '@/api/article'
import { useAuthStore } from '@/store/auth'
import { useLoginModal } from '@/store/loginModal'
import { useNavigate, useSearchParams } from 'react-router-dom'
import ArticleCard from '@/components/forum/ArticleCard'
import { formatDateTime } from '@/utils/time'
import styles from './UserProfile.module.css'

export default function UserProfile() {
  const [searchParams] = useSearchParams()
  const tabFromUrl = searchParams.get('tab') || 'profile'
  const [user, setUser] = useState<any>(null)
  const [articles, setArticles] = useState<any[]>([])
  const [favorites, setFavorites] = useState<any[]>([])
  const [likes, setLikes] = useState<any[]>([])
  const [activeTab, setActiveTab] = useState(tabFromUrl)
  const { user: currentUser, setUser: setCurrentUser } = useAuthStore()
  const openLoginModal = useLoginModal((s) => s.open)
  const navigate = useNavigate()

  const [editUsername, setEditUsername] = useState('')
  const [editEmail, setEditEmail] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!currentUser) {
      openLoginModal()
      return
    }
    authApi.getProfile().then(r => {
      const u = r.data.data
      setUser(u)
      setEditUsername(u.username)
      setEditEmail(u.email || '')
    })
  }, [currentUser, openLoginModal])

  useEffect(() => {
    if (!user) return
    if (activeTab === 'articles') {
      articleApi.list({ page: 1, page_size: 50 }).then(r => {
        const list = r.data.data?.list || []
        setArticles(list.filter((a: any) => a.author_name === user.username))
      })
    } else if (activeTab === 'favorites') {
      authApi.getFavorites({ page: 1, page_size: 50 }).then(r => {
        setFavorites(r.data.data?.list || [])
      })
    } else if (activeTab === 'likes') {
      authApi.getLikes({ page: 1, page_size: 50 }).then(r => {
        setLikes(r.data.data?.list || [])
      })
    }
  }, [user, activeTab])

  const handleTabChange = (key: string) => {
    setActiveTab(key)
    navigate(`/user/profile${key !== 'profile' ? `?tab=${key}` : ''}`, { replace: true })
  }

  const handleSaveProfile = async () => {
    setSaving(true)
    try {
      const data: { username?: string; email?: string } = {}
      if (editUsername !== user.username) data.username = editUsername
      if (editEmail !== (user.email || '')) data.email = editEmail
      if (Object.keys(data).length === 0) {
        message.info('没有修改')
        setSaving(false)
        return
      }
      const res = await authApi.updateProfile(data)
      const updated = res.data.data
      setUser(updated)
      setCurrentUser({ ...currentUser, username: updated.username, email: updated.email, avatar: updated.avatar })
      message.success('保存成功')
    } catch {
      message.error('保存失败')
    } finally {
      setSaving(false)
    }
  }

  const handleAvatarUpload = async (file: File) => {
    try {
      const res = await authApi.uploadAvatar(file)
      const avatar = res.data.data.avatar
      setUser((u: any) => ({ ...u, avatar }))
      setCurrentUser({ ...currentUser, avatar })
      message.success('头像更新成功')
    } catch {
      message.error('头像上传失败')
    }
    return false
  }

  if (!user) return <div className={styles.spinWrap}><Spin /></div>

  const profileTab = (
    <div className={styles.profileTab}>
      <div className={styles.avatarSection}>
        <div className={styles.avatarWrap}>
          <div className={styles.avatar}>{user.username?.[0]?.toUpperCase()}</div>
          <Upload
            beforeUpload={(f) => { handleAvatarUpload(f); return false }}
            showUploadList={false}
            accept=".jpg,.jpeg,.png,.webp"
          >
            <button className={styles.avatarUploadBtn}>
              <CameraOutlined /> 更换
            </button>
          </Upload>
        </div>
        <h1 className={styles.username}>{user.username}</h1>
        <div className={styles.metaRow}>
          <span className={styles.metaItem}><MailOutlined /> {user.email || '未设置邮箱'}</span>
          <span className={styles.metaItem}><CalendarOutlined /> 加入于 {formatDateTime(user.created_at)}</span>
        </div>
      </div>

      <div className={styles.editSection}>
        <h3 className={styles.sectionTitle}>编辑资料</h3>
        <div className={styles.editField}>
          <label className={styles.label}>用户名</label>
          <Input
            value={editUsername}
            onChange={e => setEditUsername(e.target.value)}
            className={styles.editInput}
          />
        </div>
        <div className={styles.editField}>
          <label className={styles.label}>邮箱</label>
          <Input
            value={editEmail}
            onChange={e => setEditEmail(e.target.value)}
            className={styles.editInput}
          />
        </div>
        <button className={styles.saveBtn} onClick={handleSaveProfile} disabled={saving}>
          <SaveOutlined /> {saving ? '保存中...' : '保存修改'}
        </button>
      </div>
    </div>
  )

  const articlesTab = (
    <div className={styles.listTab}>
      {articles.length === 0 ? (
        <p className={styles.emptyList}>还没有发布文章</p>
      ) : (
        articles.map(a => <ArticleCard key={a.id} article={a} />)
      )}
    </div>
  )

  const favoritesTab = (
    <div className={styles.listTab}>
      {favorites.length === 0 ? (
        <p className={styles.emptyList}>还没有收藏文章</p>
      ) : (
        favorites.map(a => <ArticleCard key={a.id} article={a} />)
      )}
    </div>
  )

  const likesTab = (
    <div className={styles.listTab}>
      {likes.length === 0 ? (
        <p className={styles.emptyList}>还没有点赞文章</p>
      ) : (
        likes.map(a => <ArticleCard key={a.id} article={a} />)
      )}
    </div>
  )

  return (
    <div className={styles.page}>
      <Tabs
        activeKey={activeTab}
        onChange={handleTabChange}
        centered
        items={[
          { key: 'profile', label: '资料', children: profileTab },
          { key: 'articles', label: `我的文章 (${articles.length})`, children: articlesTab },
          { key: 'favorites', label: `我的收藏 (${favorites.length})`, children: favoritesTab },
          { key: 'likes', label: `我的点赞 (${likes.length})`, children: likesTab },
        ]}
      />
    </div>
  )
}
