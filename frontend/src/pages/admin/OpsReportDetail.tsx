import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Alert, Card, Collapse, List, Spin, Tag, Timeline, Typography } from 'antd'
import { BulbOutlined, CodeOutlined, SearchOutlined, WarningOutlined, InfoCircleOutlined } from '@ant-design/icons'
import { opsApi } from '@/api/ops'
import { formatDateTime } from '@/utils/time'
import styles from './OpsReportDetail.module.css'

interface CommandGuidance {
  purpose: string
  command: string
  expected?: string
  risk?: 'low' | 'medium' | 'high'
  approval_required?: boolean
}

interface ChangeGuidance {
  target: string
  instruction: string
  command_or_patch?: string
  risk?: 'low' | 'medium' | 'high'
  preconditions?: string[]
  validation?: string
  rollback?: string
  approval_required?: boolean
}

const riskColor = (risk?: string) => risk === 'high' ? 'red' : risk === 'medium' ? 'orange' : 'green'

function CommandList({ items }: { items?: CommandGuidance[] }) {
  if (!items?.length) return null
  return <List dataSource={items} renderItem={(item, index) => (
    <List.Item className={styles.guidanceItem}>
      <div className={styles.guidanceHeader}>
        <strong>{index + 1}. {item.purpose}</strong>
        <span>
          <Tag color={riskColor(item.risk)}>{item.risk || 'low'}</Tag>
          {item.approval_required && <Tag color="gold">需要人工审批</Tag>}
        </span>
      </div>
      <Typography.Text className={styles.command} code copyable={{ text: item.command }}>{item.command}</Typography.Text>
      {item.expected && <p className={styles.expected}>判断标准：{item.expected}</p>}
    </List.Item>
  )} />
}

export default function OpsReportDetail() {
  const { id } = useParams()
  const [report, setReport] = useState<any>(null)
  const [timeline, setTimeline] = useState<any[]>([])

  useEffect(() => {
    opsApi.getReport(id!).then(r => setReport(r.data.data))
    opsApi.getTimeline(id!).then(r => setTimeline(r.data.data || []))
  }, [id])

  if (!report) return <Spin style={{ display: 'block', margin: '80px auto' }} />

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <Tag color={report.trigger_type === 'alert' ? 'red' : 'blue'}>{report.trigger_type}</Tag>
        <Tag color={report.status === 'done' ? 'green' : report.status === 'failed' ? 'red' : 'processing'}>{report.status}</Tag>
        {report.incident && <Link to={`/admin/incidents/${report.incident.id}`}>关联故障事件：{report.incident.title}</Link>}
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
      {(report.verification_commands?.length > 0 || report.change_plan?.length > 0 || report.validation_commands?.length > 0 || report.rollback_commands?.length > 0) && (
        <Alert className={styles.approvalWarning} type="warning" showIcon message="以下内容由 Agent 生成，仅供人工审核后执行" description="排查和验证命令经过只读白名单过滤；修改与回滚步骤不会由 TechMind 自动执行。执行前请确认命名空间、资源名、当前版本和备份。" />
      )}
      {report.verification_commands?.length > 0 && (
        <Card className={styles.block} title={<><SearchOutlined /> 可复制排查命令</>}>
          <CommandList items={report.verification_commands} />
        </Card>
      )}
      {report.change_plan?.length > 0 && (
        <Card className={styles.block} title={<><CodeOutlined /> 建议修改方案（需审批）</>}>
          <List dataSource={report.change_plan as ChangeGuidance[]} renderItem={(item, index) => (
            <List.Item className={styles.guidanceItem}>
              <div className={styles.guidanceHeader}>
                <strong>{index + 1}. {item.target}</strong>
                <span><Tag color={riskColor(item.risk)}>{item.risk || 'medium'}</Tag><Tag color="gold">需要人工审批</Tag></span>
              </div>
              <p>{item.instruction}</p>
              {!!item.preconditions?.length && <p className={styles.expected}>前置条件：{item.preconditions.join('；')}</p>}
              {item.command_or_patch && <Typography.Text className={styles.command} code copyable={{ text: item.command_or_patch }}>{item.command_or_patch}</Typography.Text>}
              {item.validation && <p className={styles.expected}>验证：{item.validation}</p>}
              {item.rollback && <p className={styles.expected}>回滚：{item.rollback}</p>}
            </List.Item>
          )} />
        </Card>
      )}
      {report.validation_commands?.length > 0 && (
        <Card className={styles.block} title="修改后验证命令"><CommandList items={report.validation_commands} /></Card>
      )}
      {report.rollback_commands?.length > 0 && (
        <Card className={styles.block} title="回滚命令（需审批）"><CommandList items={report.rollback_commands} /></Card>
      )}
      {report.related_changes?.length > 0 && (
        <Card className={styles.block} title="关联变更">
          {report.related_changes.map((change: string, i: number) => (
            <p key={i} style={{ marginBottom: 8 }}>{change}</p>
          ))}
        </Card>
      )}
      <Card className={styles.block} title="诊断证据链">
        <Timeline items={timeline.map((call) => ({
          color: 'blue',
          children: <div className={styles.timelineItem}>
            <strong>{call.tool_name}</strong>
            <span>{call.duration_ms}ms</span>
            <p>{call.input?.reason || '基础取证'}</p>
            <pre>{JSON.stringify(call.output, null, 2)}</pre>
          </div>,
        }))} />
      </Card>
      <Collapse ghost items={[{
        key: '1', label: '工具调用记录',
        children: <pre style={{ fontSize: 12, color: '#8b949e' }}>{JSON.stringify(report.tool_calls, null, 2)}</pre>
      }]} />
    </div>
  )
}
