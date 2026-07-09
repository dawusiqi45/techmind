import { useEffect, useState } from 'react'
import { Tabs, Spin } from 'antd'
import { CalendarOutlined, MailOutlined } from '@ant-design/icons'
import { authApi } from '@/api/auth'
import { articleApi } from '@/api/article'
import { useAuthStore } from '@/store/auth'
import { useNavigate } from 'react-router-dom'
import ArticleCard from '@/components/forum/ArticleCard'
import { formatDateTime } from '@/utils/time'
import styles from './UserProfile.module.css'

export default function UserProfile() {
  const [user, setUser] = useState<any>(null)
  const [articles, setArticles] = useState<any[]>([])
  const { user: currentUser } = useAuthStore()
  const navigate = useNavigate()

  useEffect(() => {
    if (!currentUser) {
      navigate('/login')
      return
    }
    authApi.getProfile().then(r => setUser(r.data.data))
    articleApi.list({ page: 1, page_size: 20 }).then(r => setArticles(r.data.data?.list || []))
  }, [currentUser, navigate])

  if (!user) return <div className={styles.spinWrap}><Spin /></div>

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <div className={styles.avatar}>{user.username?.[0]?.toUpperCase()}</div>
        <div className={styles.info}>
          <h1 className={styles.username}>{user.username}</h1>
          <div className={styles.meta}>
            <span className={styles.metaItem}><MailOutlined /> {user.email}</span>
            <span className={styles.metaItem}><CalendarOutlined /> 加入于 {formatDateTime(user.created_at)}</span>
          </div>
        </div>
      </div>

      <div className={styles.body}>
        <Tabs
          items={[{
            key: 'articles',
            label: `我的文章 (${articles.length})`,
            children: (
              <div className={styles.list}>
                {articles.length === 0 ? (
                  <p className={styles.empty}>还没有发布文章</p>
                ) : (
                  articles.map(a => <ArticleCard key={a.id} article={a} />)
                )}
              </div>
            ),
          }]}
        />
      </div>
    </div>
  )
}
