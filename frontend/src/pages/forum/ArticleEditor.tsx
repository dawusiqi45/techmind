import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Input, Select, message } from 'antd'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { articleApi } from '@/api/article'
import { useAuthStore } from '@/store/auth'
import { useLoginModal } from '@/store/loginModal'
import styles from './ArticleEditor.module.css'

export default function ArticleEditor() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { user } = useAuthStore()
  const openLoginModal = useLoginModal((s) => s.open)
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [tags, setTags] = useState<string[]>([])
  const [allTags, setAllTags] = useState<any[]>([])
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!user) {
      openLoginModal()
      return
    }
    articleApi.tags().then(r => setAllTags(r.data.data || []))
    if (id) {
      articleApi.get(id).then(r => {
        const a = r.data.data
        setTitle(a.title)
        setContent(a.content)
        setTags(a.tags || [])
      })
    }
  }, [id, user, openLoginModal])

  async function handleSubmit() {
    if (!title.trim() || !content.trim()) return message.error('标题和内容不能为空')
    setSaving(true)
    try {
      if (id) {
        await articleApi.update(id, { title, content, tags })
        message.success('更新成功')
      } else {
        const r = await articleApi.create({ title, content, tags })
        message.success('发布成功')
        navigate(`/articles/${r.data.data.id}`)
      }
    } catch {
      message.error('操作失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className={styles.page}>
      <div className={styles.toolbar}>
        <Input
          value={title}
          onChange={e => setTitle(e.target.value)}
          placeholder="文章标题..."
          className={styles.titleInput}
          variant="borderless"
        />
        <div className={styles.toolbarRight}>
          <Select
            mode="multiple"
            placeholder="添加标签"
            value={tags}
            onChange={setTags}
            options={allTags.map(t => ({ value: t.name, label: t.name }))}
            className={styles.tagSelect}
            size="small"
          />
          <button className={styles.cancelBtn} onClick={() => navigate(-1)}>取消</button>
          <button className={styles.publishBtn} onClick={handleSubmit} disabled={saving}>
            {saving ? '保存中...' : id ? '更新' : '发布'}
          </button>
        </div>
      </div>

      <div className={styles.editor}>
        <div className={styles.editorPane}>
          <textarea
            className={styles.textarea}
            value={content}
            onChange={e => setContent(e.target.value)}
            placeholder="用 Markdown 写作...&#10;&#10;支持 **粗体**、`代码`、# 标题等语法"
          />
        </div>
        <div className={styles.previewPane}>
          <div className={styles.previewLabel}>预览</div>
          <div className={styles.preview}>
            <ReactMarkdown remarkPlugins={[remarkGfm]}>
              {content || '*在左侧输入内容，这里将显示预览*'}
            </ReactMarkdown>
          </div>
        </div>
      </div>
    </div>
  )
}
