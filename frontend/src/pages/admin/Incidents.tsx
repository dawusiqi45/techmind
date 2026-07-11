import { useEffect, useState } from 'react'
import { Select, Space, Table, Tag } from 'antd'
import { useNavigate } from 'react-router-dom'
import { incidentApi } from '@/api/incident'
import { formatDateTime } from '@/utils/time'

export default function Incidents() {
  const [data, setData] = useState<any[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState<string | undefined>()
  const navigate = useNavigate()

  useEffect(() => {
    incidentApi.list({ page, page_size: 20, status }).then((res) => {
      setData(res.data.data?.list || [])
      setTotal(res.data.data?.total || 0)
    })
  }, [page, status])

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Select
          allowClear
          placeholder="状态筛选"
          style={{ width: 160 }}
          options={[
            { value: 'open', label: 'open' },
            { value: 'resolved', label: 'resolved' },
          ]}
          onChange={(value) => { setStatus(value); setPage(1) }}
        />
      </Space>
      <Table
        dataSource={data}
        rowKey="id"
        size="small"
        onRow={(record) => ({ onClick: () => navigate(`/admin/incidents/${record.id}`) })}
        pagination={{ current: page, total, pageSize: 20, onChange: setPage }}
        columns={[
          { title: '故障事件', dataIndex: 'title', ellipsis: true },
          { title: '严重度', dataIndex: 'severity', render: (value) => <Tag color={value === 'critical' ? 'red' : 'orange'}>{value}</Tag> },
          { title: '状态', dataIndex: 'status', render: (value) => <Tag color={value === 'open' ? 'red' : 'green'}>{value}</Tag> },
          { title: '创建时间', dataIndex: 'created_at', render: formatDateTime },
        ]}
      />
    </div>
  )
}
