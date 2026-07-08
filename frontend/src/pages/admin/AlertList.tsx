import { useEffect, useState } from 'react'
import { Table, Select, Space } from 'antd'
import { useNavigate } from 'react-router-dom'
import { alertApi } from '@/api/alert'
import { SeverityBadge, StatusBadge } from '@/components/admin/AlertBadge'
import { formatDateTime } from '@/utils/time'

export default function AlertList() {
  const [data, setData] = useState<any[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState<string | undefined>()
  const navigate = useNavigate()

  useEffect(() => {
    alertApi.list({ page, page_size: 20, status }).then(r => {
      setData(r.data.data?.list || [])
      setTotal(r.data.data?.total || 0)
    })
  }, [page, status])

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Select
          placeholder="状态筛选" allowClear style={{ width: 160 }}
          onChange={(v) => { setStatus(v); setPage(1) }}
          options={[{ value: 'firing', label: 'firing' }, { value: 'acknowledged', label: 'acknowledged' }, { value: 'resolved', label: 'resolved' }]}
        />
      </Space>
      <Table
        dataSource={data} rowKey="id" size="small"
        onRow={r => ({ onClick: () => navigate(`/admin/alerts/${r.id}`) })}
        pagination={{ current: page, total, pageSize: 20, onChange: setPage }}
        columns={[
          { title: '告警名', dataIndex: 'alert_name' },
          { title: '服务', dataIndex: 'service' },
          { title: '严重度', dataIndex: 'severity', render: (v) => <SeverityBadge severity={v} /> },
          { title: '状态', dataIndex: 'status', render: (v) => <StatusBadge status={v} /> },
          { title: '重复', dataIndex: 'repeat_count', width: 70, render: (v) => `×${v}` },
          { title: '首次', dataIndex: 'first_seen_at', render: formatDateTime },
          { title: '最近', dataIndex: 'last_seen_at', render: formatDateTime },
        ]}
      />
    </div>
  )
}
