import { useEffect, useState } from 'react'
import { Table, Tag } from 'antd'
import { monitorApi } from '@/api/monitor'
import { formatDateTime } from '@/utils/time'

export default function MonitorAI() {
  const [data, setData] = useState<any[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)

  useEffect(() => {
    monitorApi.aiCalls({ page, page_size: 20 }).then(r => {
      setData(r.data.data?.list || [])
      setTotal(r.data.data?.total || 0)
    })
  }, [page])

  return (
    <Table
      dataSource={data} rowKey="id" size="small"
      pagination={{ current: page, total, pageSize: 20, onChange: setPage }}
      columns={[
        { title: 'Skill', dataIndex: 'skill', width: 160 },
        { title: '模型', dataIndex: 'model', ellipsis: true },
        { title: '耗时(ms)', dataIndex: 'duration_ms', width: 100 },
        { title: 'Tokens', render: (_: any, r: any) => `${r.input_tokens}+${r.output_tokens}`, width: 120 },
        { title: '状态', dataIndex: 'status', render: (v: string) => <Tag color={v === 'ok' ? 'green' : 'red'}>{v}</Tag>, width: 80 },
        { title: '时间', dataIndex: 'created_at', render: formatDateTime },
      ]}
    />
  )
}
