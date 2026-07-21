import { useState } from 'react'
import { Alert, Form, Input, Button, message, Card } from 'antd'
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
    <Card title="SRE Agent 手动诊断" style={{ maxWidth: 600, background: '#161b22', border: '1px solid #30363d' }} headStyle={{ color: '#e6edf3', borderBottom: '1px solid #30363d' }}>
      <div style={{ marginBottom: 24, color: '#8b949e', lineHeight: 1.7 }}>
        <p>Agent 会只读采集慢请求、错误事件、Redis Stream、Prometheus 趋势、Pod/Event/Deployment、受限 Pod 日志、近期变更和 Runbook；每次查询都会写入报告的证据链，它不会修改集群或业务数据。</p>
        <ol style={{ margin: '10px 0 0', paddingLeft: 20 }}>
          <li>填写当前异常对应的告警名称或服务名。</li>
          <li>点击“触发诊断”，任务将交给 Worker 异步处理。</li>
          <li>在“诊断报告”中查看状态与最终结论。</li>
        </ol>
      </div>
      <Alert
        type="info"
        showIcon
        style={{ marginBottom: 20 }}
        message="生成最终报告需要部署时配置 LLM API Key"
        description="未配置 Key 时，工具取证仍会执行，但报告会标记为 failed；可在 Helm 的 secrets.llmApiKey 中注入，不会在页面或接口中暴露 Key。"
      />
      <Form layout="vertical" onFinish={handleSubmit}>
        <Form.Item name="alert_name" label={<span style={{ color: '#8b949e' }}>告警名称</span>}>
          <Input placeholder="例如: SearchLatencyHigh（留空则执行通用诊断）" style={{ background: '#0f1117', color: '#e6edf3', borderColor: '#30363d' }} />
        </Form.Item>
        <Form.Item name="service" label={<span style={{ color: '#8b949e' }}>服务名</span>}>
          <Input placeholder="例如: techmind-server（可选）" style={{ background: '#0f1117', color: '#e6edf3', borderColor: '#30363d' }} />
        </Form.Item>
        <Button type="primary" htmlType="submit" loading={loading}>触发诊断</Button>
      </Form>
    </Card>
  )
}
