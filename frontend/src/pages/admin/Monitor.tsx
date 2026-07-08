import { useEffect, useState } from 'react'
import { Row, Col, Table, Spin } from 'antd'
import { monitorApi } from '@/api/monitor'
import StatCard from '@/components/admin/StatCard'
import { formatDateTime } from '@/utils/time'
import styles from './Monitor.module.css'

export default function Monitor() {
  const [overview, setOverview] = useState<any>(null)
  const [slowReqs, setSlowReqs] = useState<any[]>([])
  const [errors, setErrors] = useState<any[]>([])

  async function fetchData() {
    const [ov, sl, er] = await Promise.all([
      monitorApi.overview(),
      monitorApi.slowRequests({ page: 1, page_size: 5 }),
      monitorApi.errors({ page: 1, page_size: 5 }),
    ])
    setOverview(ov.data.data)
    setSlowReqs(sl.data.data?.list || [])
    setErrors(er.data.data?.list || [])
  }

  useEffect(() => {
    fetchData()
    const timer = setInterval(fetchData, 30000)
    return () => clearInterval(timer)
  }, [])

  if (!overview) return <Spin style={{ display: 'block', margin: '80px auto' }} />

  const slowTotal = overview.slow_request_total ?? 0
  const errTotal = overview.error_event_total ?? 0
  const aiPending = overview.ai_task_pending ?? 0
  const streamLen = overview.ai_stream_length ?? 0

  return (
    <div>
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col span={6}><StatCard title="慢请求总数" value={slowTotal} status={slowTotal > 10 ? 'warning' : 'normal'} /></Col>
        <Col span={6}><StatCard title="错误事件总数" value={errTotal} status={errTotal > 0 ? 'error' : 'normal'} /></Col>
        <Col span={6}><StatCard title="AI 待处理任务" value={aiPending} status={aiPending > 20 ? 'warning' : 'normal'} /></Col>
        <Col span={6}><StatCard title="Stream 队列长度" value={streamLen} status={streamLen > 50 ? 'warning' : 'normal'} /></Col>
      </Row>
      <Row gutter={[16, 16]}>
        <Col span={12}>
          <div className={styles.tableCard}>
            <h4 className={styles.tableTitle}>最近慢请求</h4>
            <Table
              size="small"
              dataSource={slowReqs}
              rowKey="id"
              pagination={false}
              columns={[
                { title: 'Request ID', dataIndex: 'request_id', ellipsis: true, width: 200 },
                { title: '方法', dataIndex: 'method', width: 80 },
                { title: '路径', dataIndex: 'path', ellipsis: true },
                { title: '状态码', dataIndex: 'status_code', width: 80 },
                { title: '耗时(ms)', dataIndex: 'duration_ms', width: 100 },
                { title: '时间', dataIndex: 'created_at', render: formatDateTime },
              ]}
            />
          </div>
        </Col>
        <Col span={12}>
          <div className={styles.tableCard}>
            <h4 className={styles.tableTitle}>最近错误事件</h4>
            <Table
              size="small"
              dataSource={errors}
              rowKey="id"
              pagination={false}
              columns={[
                { title: '来源', dataIndex: 'source' },
                { title: '消息', dataIndex: 'message', ellipsis: true },
                { title: '次数', dataIndex: 'count' },
              ]}
            />
          </div>
        </Col>
      </Row>
    </div>
  )
}
