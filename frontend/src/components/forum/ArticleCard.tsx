import { LikeOutlined, MessageOutlined, EyeOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { fromNow } from '@/utils/time'
import styles from './ArticleCard.module.css'

interface Props {
  article: {
    id: number; title: string; summary: string; author_name: string
    like_count: number; favorite_count: number
    comment_count: number; view_count: number; created_at: string
  }
}

export default function ArticleCard({ article }: Props) {
  const nav = useNavigate()
  return (
    <div className={styles.card} onClick={() => nav(`/articles/${article.id}`)}>
      <h3 className={styles.title}>{article.title}</h3>
      <p className={styles.summary}>{article.summary}</p>
      <div className={styles.meta}>
        <span className={styles.author}>{article.author_name}</span>
        <span className={styles.time}>{fromNow(article.created_at)}</span>
        <div className={styles.stats}>
          <span><EyeOutlined /> {article.view_count}</span>
          <span><LikeOutlined /> {article.like_count}</span>
          <span><MessageOutlined /> {article.comment_count}</span>
        </div>
      </div>
    </div>
  )
}
