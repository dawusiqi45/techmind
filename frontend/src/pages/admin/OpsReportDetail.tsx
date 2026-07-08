import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { Card, Collapse, List, Spin, Tag } from 'antd'
import { BulbOutlined, SearchOutlined, WarningOutlined, InfoCircleOutlined } from '@ant-design/icons'
import { opsApi } from '@/api/ops'
import { formatDateTime } from '@/utils/time'
import styles from './OpsReportDetail.module.css'

export default function OpsReportDetail() {
  const { id } = useParams()
  const [report, setReport] = useState<any>(null)

  useEffect(() => {
    opsApi.getReport(id!).then(r => setReport(r.data.data))
  }, [id])

  if (!report) return <Spin style={{ display: 'block', margin: '80px auto' }} />

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <Tag color={report.trigger_type === 'alert' ? 'red' : 'blue'}>{report.trigger_type}</Tag>
        <span className={styles.time}>{formatDateTime(report.created_at)}</span>
      </div>
      <Card className={styles.block} title={<><InfoCircleOutlined /> 摘要</>}>
        <p>{report.summary}</p>
      </Card>
      <Card className={styles.block} title={<><SearchOutlined /> 证据</>}>
        <List dataSource={report.evidence || []} renderItem={(item: string) => <List.Item>{item}</List.Item>} size="small" />
      </Card>
      <Card className={styles.block} title={<><WarningOutlined /> 根因</>}>
        <p>{report.root_cause}</p>
      </Card>
      {report.impact && (
        <Card className={styles.block} title="影响范围"><p>{report.impact}</p></Card>
      )}
      <Card className={styles.block} title={<><BulbOutlined /> 建议</>}>
        <List dataSource={report.suggestions || []} renderItem={(item: string, i) => <List.Item><Tag color="green">{i + 1}</Tag>{item}</List.Item>} size="small" />
      </Card>
      {report.related_changes?.length > 0 && (
        <Card className={styles.block} title="关联变更">
          {report.related_changes.map((c: any, i: number) => (
            <p key={i} style={{ marginBottom: 8 }}>{c.service} {c.old_image} → {c.image} ({formatDateTime(c.changed_at)})</p>
          ))}
        </Card>
      )}
      <Collapse ghost items={[{
        key: '1', label: '工具调用记录',
        children: <pre style={{ fontSize: 12, color: '#8b949e' }}>{JSON.stringify(report.tool_calls, null, 2)}</pre>
      }]} />
    </div>
  )
}
