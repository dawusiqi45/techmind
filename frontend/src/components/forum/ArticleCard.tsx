import { EyeOutlined, LikeOutlined, MessageOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { fromNow } from '@/utils/time'
import styles from './ArticleCard.module.css'

interface Props {
  article: {
    id: number
    title: string
    summary: string
    author_name: string
    like_count: number
    comment_count: number
    view_count: number
    created_at: string
  }
}

export default function ArticleCard({ article }: Props) {
  const nav = useNavigate()
  return (
    <article
      className={styles.item}
      onClick={() => nav(`/articles/${article.id}`)}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => e.key === 'Enter' && nav(`/articles/${article.id}`)}
    >
      <h3 className={styles.title}>{article.title}</h3>
      {article.summary && (
        <p className={styles.summary}>{article.summary}</p>
      )}
      <div className={styles.meta}>
        <span className={styles.author}>{article.author_name}</span>
        <span className={styles.dot}>·</span>
        <span className={styles.time}>{fromNow(article.created_at)}</span>
        <div className={styles.stats}>
          <span className={styles.stat}><EyeOutlined />{article.view_count}</span>
          <span className={styles.stat}><LikeOutlined />{article.like_count}</span>
          <span className={styles.stat}><MessageOutlined />{article.comment_count}</span>
        </div>
      </div>
    </article>
  )
}
