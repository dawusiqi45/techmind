import { useEffect, useState } from 'react'
import { Table } from 'antd'
import { monitorApi } from '@/api/monitor'
import { formatDateTime } from '@/utils/time'

export default function MonitorSlow() {
  const [data, setData] = useState<any[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)

  useEffect(() => {
    monitorApi.slowRequests({ page, page_size: 20 }).then(r => {
      setData(r.data.data?.list || [])
      setTotal(r.data.data?.total || 0)
    })
  }, [page])

  return (
    <Table
      dataSource={data} rowKey="id" size="small"
      pagination={{ current: page, total, pageSize: 20, onChange: setPage }}
      columns={[
        { title: 'Request ID', dataIndex: 'request_id', ellipsis: true, width: 220 },
        { title: '方法', dataIndex: 'method', width: 80 },
        { title: '路径', dataIndex: 'path', ellipsis: true },
        { title: '状态码', dataIndex: 'status_code', width: 80 },
        { title: '耗时(ms)', dataIndex: 'duration_ms', width: 100, sorter: (a: any, b: any) => a.duration_ms - b.duration_ms },
        { title: '时间', dataIndex: 'created_at', render: formatDateTime },
      ]}
    />
  )
}
