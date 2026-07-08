import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Spin } from 'antd'
import { RobotOutlined } from '@ant-design/icons'
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
    searchApi.search(q).then(r => setResult(r.data.data)).finally(() => setLoading(false))
  }, [q])

  if (loading) return <Spin style={{ display: 'block', margin: '80px auto' }} />

  return (
    <div className={styles.container}>
      <p className={styles.keyword}>"{q}" 的搜索结果</p>
      {result?.summary && (
        <div className={styles.aiSummary}>
          <div className={styles.aiHeader}><RobotOutlined /> AI 搜索总结</div>
          <p>{result.summary}</p>
        </div>
      )}
      <div>{(result?.list || []).map((a: any) => <ArticleCard key={a.id} article={a} />)}</div>
    </div>
  )
}
