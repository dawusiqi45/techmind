import { Tag } from 'antd'
const severityColor: Record<string, string> = { critical: 'red', warning: 'gold', info: 'blue' }
const statusColor: Record<string, string> = { firing: 'red', acknowledged: 'orange', resolved: 'green' }
export function SeverityBadge({ severity }: { severity: string }) {
  return <Tag color={severityColor[severity] ?? 'default'}>{severity}</Tag>
}
export function StatusBadge({ status }: { status: string }) {
  return <Tag color={statusColor[status] ?? 'default'}>{status}</Tag>
}
