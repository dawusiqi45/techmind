import { useEffect, useState } from 'react'
import { Avatar, Tabs, Spin } from 'antd'
import { authApi } from '@/api/auth'
import { articleApi } from '@/api/article'
import ArticleCard from '@/components/forum/ArticleCard'
import styles from './UserProfile.module.css'

export default function UserProfile() {
  const [user, setUser] = useState<any>(null)
  const [articles, setArticles] = useState<any[]>([])

  useEffect(() => {
    authApi.getProfile().then(r => setUser(r.data.data))
    articleApi.list({ page: 1, page_size: 20 }).then(r => setArticles(r.data.data?.list || []))
  }, [])

  if (!user) return <Spin style={{ display: 'block', margin: '80px auto' }} />

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <Avatar size={72} src={user.avatar}>{user.username[0]}</Avatar>
        <div>
          <h2>{user.username}</h2>
          <p style={{ color: '#57606a' }}>{user.email}</p>
        </div>
      </div>
      <Tabs items={[{ key: 'articles', label: '我的文章', children: articles.map(a => <ArticleCard key={a.id} article={a} />) }]} />
    </div>
  )
}
