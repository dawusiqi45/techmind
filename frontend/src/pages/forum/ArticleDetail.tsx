import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Input, Spin } from 'antd'
import { LikeOutlined, StarOutlined, LikeFilled, StarFilled, ArrowLeftOutlined } from '@ant-design/icons'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { articleApi } from '@/api/article'
import { useAuthStore } from '@/store/auth'
import { useLoginModal } from '@/store/loginModal'
import { fromNow } from '@/utils/time'
import styles from './ArticleDetail.module.css'

export default function ArticleDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [article, setArticle] = useState<any>(null)
  const [comments, setComments] = useState<any[]>([])
  const [commentText, setCommentText] = useState('')
  const [liked, setLiked] = useState(false)
  const [favorited, setFavorited] = useState(false)
  const { user } = useAuthStore()
  const openLoginModal = useLoginModal((s) => s.open)

  useEffect(() => {
    if (!id) return
    articleApi.get(id).then(r => setArticle(r.data.data))
    articleApi.listComments(id).then(r => setComments(r.data.data || []))
  }, [id])

  const requireLogin = (action: () => void) => {
    if (!user) {
      openLoginModal()
      return
    }
    action()
  }

	async function handleLike() {
		requireLogin(async () => {
			const res = await articleApi.like(id!)
			const nextLiked = Boolean(res.data.data?.liked)
			setLiked(nextLiked)
			setArticle((a: any) => ({ ...a, like_count: Math.max(0, a.like_count + (nextLiked ? 1 : -1)) }))
		})
	}

  async function handleFavorite() {
    requireLogin(async () => {
			const res = await articleApi.favorite(id!)
			const nextFavorited = Boolean(res.data.data?.favorited)
			setFavorited(nextFavorited)
			setArticle((a: any) => ({ ...a, favorite_count: Math.max(0, a.favorite_count + (nextFavorited ? 1 : -1)) }))
		})
	}

  async function handleComment() {
    requireLogin(async () => {
      if (!commentText.trim()) return
      await articleApi.createComment(id!, { content: commentText })
      setCommentText('')
      const r = await articleApi.listComments(id!)
      setComments(r.data.data || [])
    })
  }

  if (!article) return (
    <div className={styles.spinWrap}><Spin /></div>
  )

  return (
    <div className={styles.page}>
      <div className={styles.container}>
        <button className={styles.back} onClick={() => navigate(-1)}>
          <ArrowLeftOutlined /> 返回
        </button>

        <article className={styles.article}>
          <h1 className={styles.title}>{article.title}</h1>
          <div className={styles.meta}>
            <div className={styles.avatar}>{article.author_name?.[0]?.toUpperCase()}</div>
            <span className={styles.author}>{article.author_name}</span>
            <span className={styles.dot}>·</span>
            <span className={styles.time}>{fromNow(article.created_at)}</span>
            <span className={styles.dot}>·</span>
            <span className={styles.views}>{article.view_count} 阅读</span>
          </div>

          <div className={styles.content}>
            <ReactMarkdown
              remarkPlugins={[remarkGfm]}
              components={{
                code({ className, children }) {
                  const lang = /language-(\w+)/.exec(className || '')?.[1]
                  return lang ? (
                    <SyntaxHighlighter
                      style={oneDark}
                      language={lang}
                      customStyle={{ borderRadius: 8, fontSize: 13 }}
                    >
                      {String(children).replace(/\n$/, '')}
                    </SyntaxHighlighter>
                  ) : (
                    <code className={className}>{children}</code>
                  )
                },
              }}
            >
              {article.content}
            </ReactMarkdown>
          </div>

          <div className={styles.actions}>
            <button
              className={`${styles.actionBtn} ${liked ? styles.actionActive : ''}`}
              onClick={handleLike}
            >
              {liked ? <LikeFilled /> : <LikeOutlined />}
              {article.like_count}
            </button>
            <button
              className={`${styles.actionBtn} ${favorited ? styles.actionActive : ''}`}
              onClick={handleFavorite}
            >
              {favorited ? <StarFilled /> : <StarOutlined />}
              {article.favorite_count}
            </button>
          </div>
        </article>

        <section className={styles.comments}>
          <h2 className={styles.commentsTitle}>评论 <span>{comments.length}</span></h2>

          <div className={styles.commentInput}>
            {user ? (
              <>
                <Input.TextArea
                  rows={3}
                  value={commentText}
                  onChange={e => setCommentText(e.target.value)}
                  placeholder="写下你的想法..."
                  className={styles.textarea}
                />
                <button className={styles.submitBtn} onClick={handleComment}>
                  发布评论
                </button>
              </>
            ) : (
              <div className={styles.loginPrompt}>
                <span>登录后参与讨论</span>
                <button className={styles.promptBtn} onClick={() => openLoginModal()}>
                  去登录 →
                </button>
              </div>
            )}
          </div>

          <div className={styles.commentList}>
            {comments.map((c: any) => (
              <div key={c.id} className={styles.comment}>
                <div className={styles.commentAvatar}>{c.author_name?.[0]?.toUpperCase()}</div>
                <div className={styles.commentBody}>
                  <div className={styles.commentHeader}>
                    <span className={styles.commentAuthor}>{c.author_name}</span>
                    <span className={styles.commentTime}>{fromNow(c.created_at)}</span>
                  </div>
                  <p className={styles.commentText}>{c.content}</p>
                </div>
              </div>
            ))}
          </div>
        </section>
      </div>
    </div>
  )
}
