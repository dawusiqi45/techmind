import { useState, useEffect } from 'react'
import { Tag, Spin } from 'antd'
import { articleApi } from '@/api/article'
import ArticleCard from '@/components/forum/ArticleCard'
import styles from './Home.module.css'

export default function Home() {
  const [articles, setArticles] = useState<any[]>([])
  const [hotList, setHotList] = useState<any[]>([])
  const [tags, setTags] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)

  useEffect(() => {
    articleApi.hot().then(r => setHotList(r.data.data || []))
    articleApi.tags().then(r => setTags((r.data.data || []).slice(0, 15)))
  }, [])

  useEffect(() => {
    setLoading(true)
    articleApi.list({ page, page_size: 10 }).then(r => {
      setArticles(r.data.data?.list || [])
      setTotal(r.data.data?.total || 0)
    }).finally(() => setLoading(false))
  }, [page])

  return (
    <div className={styles.container}>
      <div className={styles.feed}>
        {loading ? <Spin /> : articles.map(a => <ArticleCard key={a.id} article={a} />)}
        <div style={{ textAlign: 'center', marginTop: 16 }}>
          {total > articles.length && <a onClick={() => setPage(p => p + 1)}>加载更多</a>}
        </div>
      </div>
      <aside className={styles.sidebar}>
        <div className={styles.sideCard}>
          <h4>🔥 热榜</h4>
          {hotList.map((a, i) => (
            <div key={a.id} className={styles.hotItem}>
              <span className={styles.rank}>{i + 1}</span>
              <a href={`/articles/${a.id}`}>{a.title}</a>
            </div>
          ))}
        </div>
        <div className={styles.sideCard}>
          <h4>🏷 热门标签</h4>
          <div className={styles.tagCloud}>
            {tags.map(t => <Tag key={t.id} color="blue" style={{ marginBottom: 6 }}>{t.name}</Tag>)}
          </div>
        </div>
      </aside>
    </div>
  )
}
