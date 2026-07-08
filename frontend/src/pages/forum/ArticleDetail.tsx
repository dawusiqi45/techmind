import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button, Input, message, Spin, Avatar } from 'antd'
import { LikeOutlined, StarOutlined, LikeFilled, StarFilled } from '@ant-design/icons'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { articleApi } from '@/api/article'
import { useAuthStore } from '@/store/auth'
import { fromNow } from '@/utils/time'
import styles from './ArticleDetail.module.css'

export default function ArticleDetail() {
  const { id } = useParams<{ id: string }>()
  const [article, setArticle] = useState<any>(null)
  const [comments, setComments] = useState<any[]>([])
  const [commentText, setCommentText] = useState('')
  const [liked, setLiked] = useState(false)
  const [favorited, setFavorited] = useState(false)
  const { user } = useAuthStore()
  const navigate = useNavigate()

  useEffect(() => {
    if (!id) return
    articleApi.get(id).then(r => setArticle(r.data.data))
    articleApi.listComments(id).then(r => setComments(r.data.data || []))
  }, [id])

  async function handleLike() {
    if (!user) return navigate('/login')
    await articleApi.like(id!)
    setLiked(true)
    setArticle((a: any) => ({ ...a, like_count: a.like_count + 1 }))
  }

  async function handleFavorite() {
    if (!user) return navigate('/login')
    await articleApi.favorite(id!)
    setFavorited(true)
    setArticle((a: any) => ({ ...a, favorite_count: a.favorite_count + 1 }))
  }

  async function handleComment() {
    if (!user) return navigate('/login')
    if (!commentText.trim()) return
    await articleApi.createComment(id!, { content: commentText })
    setCommentText('')
    const r = await articleApi.listComments(id!)
    setComments(r.data.data || [])
    message.success('评论成功')
  }

  if (!article) return <Spin style={{ display: 'block', margin: '80px auto' }} />

  return (
    <div className={styles.container}>
      <article className={styles.article}>
        <h1 className={styles.title}>{article.title}</h1>
        <div className={styles.meta}>
          <Avatar size="small" src={article.author_avatar}>{article.author_name?.[0]}</Avatar>
          <span>{article.author_name}</span>
          <span>{fromNow(article.created_at)}</span>
          <span>{article.view_count} 阅读</span>
        </div>
        <div className={styles.content}>
          <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            components={{
              code({ className, children }) {
                const lang = /language-(\w+)/.exec(className || '')?.[1]
                return lang ? (
                  <SyntaxHighlighter style={oneLight} language={lang}>{String(children)}</SyntaxHighlighter>
                ) : <code className={className}>{children}</code>
              }
            }}
          >{article.content}</ReactMarkdown>
        </div>
        <div className={styles.actions}>
          <Button icon={liked ? <LikeFilled /> : <LikeOutlined />} onClick={handleLike} type={liked ? 'primary' : 'default'}>
            {article.like_count} 点赞
          </Button>
          <Button icon={favorited ? <StarFilled /> : <StarOutlined />} onClick={handleFavorite} type={favorited ? 'primary' : 'default'}>
            {article.favorite_count} 收藏
          </Button>
        </div>
      </article>
      <section className={styles.comments}>
        <h3>评论 ({comments.length})</h3>
        <div className={styles.commentInput}>
          <Input.TextArea rows={3} value={commentText} onChange={e => setCommentText(e.target.value)} placeholder="写下你的评论..." />
          <Button type="primary" onClick={handleComment} style={{ marginTop: 8 }}>发布评论</Button>
        </div>
        {comments.map((c: any) => (
          <div key={c.id} className={styles.comment}>
            <Avatar size="small" src={c.author_avatar}>{c.author_name?.[0]}</Avatar>
            <div>
              <span className={styles.commentAuthor}>{c.author_name}</span>
              <span className={styles.commentTime}>{fromNow(c.created_at)}</span>
              <p>{c.content}</p>
              {c.replies?.map((r: any) => (
                <div key={r.id} className={styles.comment} style={{ marginLeft: 24, border: 'none', paddingTop: 0 }}>
                  <Avatar size="small" src={r.author_avatar}>{r.author_name?.[0]}</Avatar>
                  <div>
                    <span className={styles.commentAuthor}>{r.author_name}</span>
                    <span className={styles.commentTime}>{fromNow(r.created_at)}</span>
                    <p>{r.content}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </section>
    </div>
  )
}
