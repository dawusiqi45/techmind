import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Descriptions, Button, Space, message, Spin, Typography } from 'antd'
import { alertApi } from '@/api/alert'
import { SeverityBadge, StatusBadge } from '@/components/admin/AlertBadge'
import { formatDateTime } from '@/utils/time'

export default function AlertDetail() {
  const { id } = useParams()
  const [alert, setAlert] = useState<any>(null)
  const navigate = useNavigate()

  useEffect(() => {
    alertApi.get(id!).then(r => setAlert(r.data.data))
  }, [id])

  async function handleAck() {
    await alertApi.ack(id!)
    message.success('已确认')
    setAlert((a: any) => ({ ...a, status: 'acknowledged' }))
  }

  async function handleDiagnose() {
    await alertApi.diagnose(id!)
    message.success('诊断任务已入队，请稍后查询报告列表')
    navigate('/admin/ops/reports')
  }

  if (!alert) return <Spin style={{ display: 'block', margin: '80px auto' }} />

  return (
    <div style={{ maxWidth: 900 }}>
      <Space style={{ marginBottom: 20 }}>
        <h2 style={{ margin: 0, color: '#e6edf3' }}>{alert.alert_name}</h2>
        <SeverityBadge severity={alert.severity} />
        <StatusBadge status={alert.status} />
      </Space>
      <Descriptions bordered column={2} size="small" style={{ marginBottom: 16 }} labelStyle={{ color: '#8b949e' }} contentStyle={{ color: '#e6edf3' }}>
        <Descriptions.Item label="服务">{alert.service}</Descriptions.Item>
        <Descriptions.Item label="端点">{alert.endpoint}</Descriptions.Item>
        <Descriptions.Item label="重复次数">{alert.repeat_count}</Descriptions.Item>
        <Descriptions.Item label="指纹">{alert.fingerprint}</Descriptions.Item>
        <Descriptions.Item label="首次触发">{formatDateTime(alert.first_seen_at)}</Descriptions.Item>
        <Descriptions.Item label="最近触发">{formatDateTime(alert.last_seen_at)}</Descriptions.Item>
        <Descriptions.Item label="Labels" span={2}><Typography.Text code>{JSON.stringify(alert.labels, null, 2)}</Typography.Text></Descriptions.Item>
        <Descriptions.Item label="Annotations" span={2}><Typography.Text code>{JSON.stringify(alert.annotations, null, 2)}</Typography.Text></Descriptions.Item>
      </Descriptions>
      <Space>
        {alert.status === 'firing' && <Button type="primary" onClick={handleAck}>确认告警</Button>}
        <Button onClick={handleDiagnose} style={{ background: '#177ddc', color: '#fff', borderColor: '#177ddc' }}>触发诊断</Button>
      </Space>
    </div>
  )
}
