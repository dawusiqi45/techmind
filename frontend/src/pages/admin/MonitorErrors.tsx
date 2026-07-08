import { useEffect, useState } from 'react'
import { Table, Tag } from 'antd'
import { monitorApi } from '@/api/monitor'
import { formatDateTime } from '@/utils/time'

export default function MonitorErrors() {
  const [data, setData] = useState<any[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)

  useEffect(() => {
    monitorApi.errors({ page, page_size: 20 }).then(r => {
      setData(r.data.data?.list || [])
      setTotal(r.data.data?.total || 0)
    })
  }, [page])

  return (
    <Table
      dataSource={data} rowKey="id" size="small"
      pagination={{ current: page, total, pageSize: 20, onChange: setPage }}
      columns={[
        { title: '来源', dataIndex: 'source', render: (v: string) => <Tag color="red">{v}</Tag>, width: 100 },
        { title: '路径', dataIndex: 'path', ellipsis: true },
        { title: '消息', dataIndex: 'message', ellipsis: true },
        { title: '次数', dataIndex: 'count', width: 80 },
        { title: '最近时间', dataIndex: 'updated_at', render: formatDateTime },
      ]}
    />
  )
}
