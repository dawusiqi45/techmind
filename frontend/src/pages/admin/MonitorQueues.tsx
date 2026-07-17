import { useEffect, useState } from 'react'
import { Card, Row, Col, Statistic } from 'antd'
import { monitorApi } from '@/api/monitor'

interface QueueStats {
  ai_task_stream: string
  ai_task_pending: number
  ai_task_length: number
  ai_dead_letter_len: number
}

export default function MonitorQueues() {
  const [data, setData] = useState<QueueStats | null>(null)

  useEffect(() => {
    monitorApi.queues().then(r => setData(r.data.data))
    const t = setInterval(() => monitorApi.queues().then(r => setData(r.data.data)), 10000)
    return () => clearInterval(t)
  }, [])

  if (!data) return null

  const stats = [
    { title: 'AI Stream', value: data.ai_task_stream, suffix: undefined, danger: false },
    { title: '待确认任务', value: data.ai_task_pending, suffix: 'pending', danger: data.ai_task_pending > 50 },
    { title: '队列长度', value: data.ai_task_length, suffix: 'messages', danger: data.ai_task_length > 100 },
    { title: '死信数量', value: data.ai_dead_letter_len, suffix: 'messages', danger: data.ai_dead_letter_len > 0 },
  ]

  return (
    <Row gutter={[16, 16]}>
      {stats.map((item) => (
        <Col span={6} key={item.title}>
          <Card style={{ background: '#161b22', border: '1px solid #30363d' }}>
            <Statistic
              title={<span style={{ color: '#8b949e' }}>{item.title}</span>}
              value={item.value}
              valueStyle={{ color: item.danger ? '#ff4d4f' : '#52c41a', fontSize: typeof item.value === 'string' ? 16 : undefined }}
              suffix={item.suffix}
            />
          </Card>
        </Col>
      ))}
    </Row>
  )
}
