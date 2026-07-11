import { useEffect, useState } from 'react'
import { Button, Card, Descriptions, List, Popconfirm, Spin, Tag, message } from 'antd'
import { useParams } from 'react-router-dom'
import { incidentApi } from '@/api/incident'
import { formatDateTime } from '@/utils/time'

export default function IncidentDetail() {
  const { id } = useParams()
  const [data, setData] = useState<any>(null)

  const load = () => incidentApi.get(id!).then((res) => setData(res.data.data))
  useEffect(() => { load() }, [id])

  async function resolve() {
    try {
      await incidentApi.resolve(id!)
      message.success('故障事件已关闭，原始告警状态未被修改')
      load()
    } catch {
      message.error('关闭故障事件失败')
    }
  }

  if (!data) return <Spin style={{ display: 'block', margin: '80px auto' }} />
  const { incident, alerts = [] } = data

  return (
    <div>
      <Card
        title="故障事件"
        extra={incident.status === 'open' && (
          <Popconfirm title="确认关闭此故障事件？该操作不会修改告警状态。" onConfirm={resolve}>
            <Button type="primary">关闭事件</Button>
          </Popconfirm>
        )}
      >
        <Descriptions column={1} size="small">
          <Descriptions.Item label="标题">{incident.title}</Descriptions.Item>
          <Descriptions.Item label="严重度"><Tag color={incident.severity === 'critical' ? 'red' : 'orange'}>{incident.severity}</Tag></Descriptions.Item>
          <Descriptions.Item label="状态"><Tag color={incident.status === 'open' ? 'red' : 'green'}>{incident.status}</Tag></Descriptions.Item>
          <Descriptions.Item label="创建时间">{formatDateTime(incident.created_at)}</Descriptions.Item>
        </Descriptions>
      </Card>
      <Card title="关联告警" style={{ marginTop: 16 }}>
        <List
          dataSource={alerts}
          locale={{ emptyText: '暂无关联告警' }}
          renderItem={(alert: any) => (
            <List.Item>
              <span>{alert.alert_name} {alert.service && `· ${alert.service}`}</span>
              <span><Tag color={alert.severity === 'critical' ? 'red' : 'orange'}>{alert.severity}</Tag><Tag>{alert.status}</Tag></span>
            </List.Item>
          )}
        />
      </Card>
    </div>
  )
}
