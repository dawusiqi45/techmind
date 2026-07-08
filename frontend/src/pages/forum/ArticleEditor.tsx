import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Input, Button, Select, message } from 'antd'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { articleApi } from '@/api/article'
import styles from './ArticleEditor.module.css'

export default function ArticleEditor() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [tagIds, setTagIds] = useState<number[]>([])
  const [allTags, setAllTags] = useState<any[]>([])

  useEffect(() => {
    articleApi.tags().then(r => setAllTags(r.data.data || []))
    if (id) {
      articleApi.get(id).then(r => {
        const a = r.data.data
        setTitle(a.title)
        setContent(a.content)
        setTagIds(a.tags?.map((t: any) => t.id) || [])
      })
    }
  }, [id])

  async function handleSubmit() {
    if (!title.trim() || !content.trim()) return message.error('标题和内容不能为空')
    try {
      if (id) {
        await articleApi.update(id, { title, content, tag_ids: tagIds })
        message.success('更新成功')
      } else {
        const r = await articleApi.create({ title, content, tag_ids: tagIds })
        message.success('发布成功')
        navigate(`/articles/${r.data.data.id}`)
      }
    } catch {
      message.error('操作失败')
    }
  }

  return (
    <div className={styles.container}>
      <Input
        className={styles.titleInput}
        placeholder="文章标题..."
        value={title}
        onChange={e => setTitle(e.target.value)}
      />
      <Select
        mode="multiple"
        placeholder="选择标签"
        value={tagIds}
        onChange={setTagIds}
        options={allTags.map(t => ({ value: t.id, label: t.name }))}
        className={styles.tagSelect}
      />
      <div className={styles.editor}>
        <textarea
          className={styles.textarea}
          value={content}
          onChange={e => setContent(e.target.value)}
          placeholder="用 Markdown 写作..."
        />
        <div className={styles.preview}>
          <ReactMarkdown remarkPlugins={[remarkGfm]}>{content || '*预览区域*'}</ReactMarkdown>
        </div>
      </div>
      <div className={styles.footer}>
        <Button onClick={() => navigate(-1)}>取消</Button>
        <Button type="primary" onClick={handleSubmit}>{id ? '更新' : '发布'}</Button>
      </div>
    </div>
  )
}
