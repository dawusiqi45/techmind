import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Spin } from 'antd'
import { RobotOutlined, SearchOutlined } from '@ant-design/icons'
import { searchApi } from '@/api/search'
import ArticleCard from '@/components/forum/ArticleCard'
import styles from './Search.module.css'

export default function Search() {
  const [params] = useSearchParams()
  const q = params.get('q') || ''
  const [result, setResult] = useState<any>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!q) return
    setLoading(true)
    setResult(null)
    searchApi.search(q).then(r => setResult(r.data.data)).finally(() => setLoading(false))
  }, [q])

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <SearchOutlined className={styles.headerIcon} />
        <h1 className={styles.keyword}>"{q}"</h1>
        {result && (
          <span className={styles.count}>{result.total ?? result.list?.length ?? 0} 个结果</span>
        )}
      </div>

      {loading && <div className={styles.spinWrap}><Spin /></div>}

      {!loading && result?.summary && (
        <div className={styles.aiBlock}>
          <div className={styles.aiLabel}>
            <RobotOutlined />
            AI 摘要
          </div>
          <p className={styles.aiText}>{result.summary}</p>
        </div>
      )}

      {!loading && (
        <div className={styles.list}>
          {(result?.list || []).map((a: any) => (
            <ArticleCard key={a.id} article={a} />
          ))}
          {result && result.list?.length === 0 && (
            <p className={styles.empty}>没有找到相关结果</p>
          )}
        </div>
      )}
    </div>
  )
}
