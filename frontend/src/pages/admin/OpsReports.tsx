import { useEffect, useState } from 'react'
import { Table, Tag } from 'antd'
import { useNavigate } from 'react-router-dom'
import { opsApi } from '@/api/ops'
import { formatDateTime } from '@/utils/time'

export default function OpsReports() {
  const [data, setData] = useState<any[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const navigate = useNavigate()

  useEffect(() => {
    opsApi.listReports({ page, page_size: 20 }).then(r => {
      setData(r.data.data?.list || [])
      setTotal(r.data.data?.total || 0)
    })
  }, [page])

  return (
    <Table
      dataSource={data} rowKey="id" size="small"
      onRow={r => ({ onClick: () => navigate(`/admin/ops/reports/${r.id}`) })}
      pagination={{ current: page, total, pageSize: 20, onChange: setPage }}
        columns={[
          { title: '摘要', dataIndex: 'summary', ellipsis: true },
          { title: '事件', dataIndex: 'incident_id', render: (v) => v && v !== '0' ? <Tag color="red">已关联</Tag> : <Tag>手动诊断</Tag>, width: 100 },
          { title: '触发方式', dataIndex: 'trigger_type', render: (v) => <Tag color={v === 'alert' ? 'red' : 'blue'}>{v}</Tag>, width: 100 },
        { title: '状态', dataIndex: 'status', width: 80 },
        { title: '创建时间', dataIndex: 'created_at', render: formatDateTime },
      ]}
    />
  )
}
