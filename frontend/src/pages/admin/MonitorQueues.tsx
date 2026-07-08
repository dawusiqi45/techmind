import { useEffect, useState } from 'react'
import { Card, Row, Col, Statistic } from 'antd'
import { monitorApi } from '@/api/monitor'

export default function MonitorQueues() {
  const [data, setData] = useState<any>(null)

  useEffect(() => {
    monitorApi.queues().then(r => setData(r.data.data))
    const t = setInterval(() => monitorApi.queues().then(r => setData(r.data.data)), 10000)
    return () => clearInterval(t)
  }, [])

  if (!data) return null

  return (
    <Row gutter={[16, 16]}>
      {Object.entries(data).map(([key, val]: any) => (
        <Col span={8} key={key}>
          <Card style={{ background: '#161b22', border: '1px solid #30363d' }}>
            <Statistic title={<span style={{ color: '#8b949e' }}>{key}</span>} value={val.pending ?? val} valueStyle={{ color: val.pending > 50 ? '#ff4d4f' : '#52c41a' }} suffix="pending" />
          </Card>
        </Col>
      ))}
    </Row>
  )
}
