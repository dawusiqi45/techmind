import { useState } from 'react'
import { Form, Input, Button, message, Card } from 'antd'
import { useNavigate } from 'react-router-dom'
import { opsApi } from '@/api/ops'

export default function OpsDiagnose() {
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  async function handleSubmit(values: { service?: string; alert_name?: string }) {
    setLoading(true)
    try {
      await opsApi.diagnose({ service: values.service, alert_name: values.alert_name })
      message.success('诊断任务已入队，请稍后查询报告列表')
      navigate('/admin/ops/reports')
    } catch {
      message.error('触发失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Card title="手动触发诊断" style={{ maxWidth: 600, background: '#161b22', border: '1px solid #30363d' }} headStyle={{ color: '#e6edf3', borderBottom: '1px solid #30363d' }}>
      <Form layout="vertical" onFinish={handleSubmit}>
        <Form.Item name="alert_name" label={<span style={{ color: '#8b949e' }}>告警名称</span>}>
          <Input placeholder="例如: APIHighErrorRate" style={{ background: '#0f1117', color: '#e6edf3', borderColor: '#30363d' }} />
        </Form.Item>
        <Form.Item name="service" label={<span style={{ color: '#8b949e' }}>服务名</span>}>
          <Input placeholder="例如: techmind-server" style={{ background: '#0f1117', color: '#e6edf3', borderColor: '#30363d' }} />
        </Form.Item>
        <Button type="primary" htmlType="submit" loading={loading}>触发诊断</Button>
      </Form>
    </Card>
  )
}
