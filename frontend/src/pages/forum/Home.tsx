import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Spin, Tag } from 'antd'
import { FireOutlined, TagsOutlined } from '@ant-design/icons'
import { articleApi } from '@/api/article'
import { useAuthStore } from '@/store/auth'
import { useLoginModal } from '@/store/loginModal'
import ArticleCard from '@/components/forum/ArticleCard'
import styles from './Home.module.css'

export default function Home() {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const openLoginModal = useLoginModal((s) => s.open)
  const [articles, setArticles] = useState<any[]>([])
  const [hotList, setHotList] = useState<any[]>([])
  const [tags, setTags] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loadingMore, setLoadingMore] = useState(false)

  useEffect(() => {
    articleApi.hot().then(r => setHotList(r.data.data || []))
    articleApi.tags().then(r => setTags((r.data.data || []).slice(0, 18)))
  }, [])

  useEffect(() => {
    if (page === 1) setLoading(true)
    else setLoadingMore(true)
    articleApi.list({ page, page_size: 15 }).then(r => {
      const list = r.data.data?.list || []
      setArticles(prev => page === 1 ? list : [...prev, ...list])
      setTotal(r.data.data?.total || 0)
    }).finally(() => {
      setLoading(false)
      setLoadingMore(false)
    })
  }, [page])

  return (
    <div className={styles.page}>
      <div className={styles.layout}>
        {/* 文章列表 */}
        <section className={styles.feed}>
          {loading ? (
            <div className={styles.spinWrap}><Spin /></div>
          ) : (
            <>
              {articles.map(a => <ArticleCard key={a.id} article={a} />)}
              {total > articles.length && (
                <button
                  className={styles.loadMore}
                  onClick={() => setPage(p => p + 1)}
                  disabled={loadingMore}
                >
                  {loadingMore ? <Spin size="small" /> : '加载更多'}
                </button>
              )}
              {articles.length === 0 && !loading && (
                <div className={styles.hero}>
                  <h2 className={styles.heroTitle}>分享你的技术见解</h2>
                  <p className={styles.heroDesc}>加入 TechMind，与开发者一起交流技术、探讨架构、分享经验</p>
                  <div className={styles.heroBtns}>
                    {user ? (
                      <button className={styles.heroPrimary} onClick={() => navigate('/articles/new')}>开始写文章</button>
                    ) : (
                      <>
                        <button className={styles.heroPrimary} onClick={openLoginModal}>注册 / 登录</button>
                        <button className={styles.heroSecondary} onClick={() => navigate('/search?q=Go')}>浏览文章</button>
                      </>
                    )}
                  </div>
                </div>
              )}
            </>
          )}
        </section>

        {/* 侧边栏 */}
        <aside className={styles.sidebar}>
          {hotList.length > 0 && (
            <div className={styles.sideCard}>
              <div className={styles.sideCardTitle}>
                <FireOutlined style={{ color: '#f59e0b' }} />
                热榜
              </div>
              <div className={styles.hotList}>
                {hotList.slice(0, 8).map((a, i) => (
                  <a
                    key={a.id}
                    href={`/articles/${a.id}`}
                    className={styles.hotItem}
                  >
                    <span className={`${styles.rank} ${i < 3 ? styles.rankTop : ''}`}>
                      {i + 1}
                    </span>
                    <span className={styles.hotTitle}>{a.title}</span>
                  </a>
                ))}
              </div>
            </div>
          )}

          {tags.length > 0 && (
            <div className={styles.sideCard}>
              <div className={styles.sideCardTitle}>
                <TagsOutlined style={{ color: 'var(--accent)' }} />
                热门标签
              </div>
              <div className={styles.tagCloud}>
                {tags.map(t => (
                  <Tag
                    key={t.id}
                    className={styles.tag}
                    onClick={() => navigate(`/search?q=${encodeURIComponent(t.name)}`)}
                  >
                    {t.name}
                  </Tag>
                ))}
              </div>
            </div>
          )}
        </aside>
      </div>
    </div>
  )
}
