import { useEffect, useState } from 'react'
import { Table, Button, Modal, Form, Input, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { runbookApi } from '@/api/runbook'
import { formatDateTime } from '@/utils/time'

export default function Runbooks() {
  const [data, setData] = useState<any[]>([])
  const [open, setOpen] = useState(false)
  const [form] = Form.useForm()

	const load = () => runbookApi.list().then(r => setData(r.data.data?.list || []))
  useEffect(() => { load() }, [])

  async function handleCreate(values: any) {
    try {
      await runbookApi.create(values)
      message.success('创建成功，向量索引异步生成中')
      setOpen(false)
      form.resetFields()
      load()
    } catch {
      message.error('创建失败')
    }
  }

  return (
    <div>
      <Button type="primary" icon={<PlusOutlined />} onClick={() => setOpen(true)} style={{ marginBottom: 16 }}>新增 Runbook</Button>
      <Table
        dataSource={data} rowKey="id" size="small"
        columns={[
          { title: '标题', dataIndex: 'title' },
          { title: '关联告警', dataIndex: 'alert_name' },
          { title: '服务', dataIndex: 'service' },
          { title: '创建时间', dataIndex: 'created_at', render: formatDateTime },
        ]}
      />
      <Modal title="新增 Runbook" open={open} onCancel={() => setOpen(false)} onOk={() => form.submit()} okText="创建">
        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Form.Item name="title" label="标题" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="alert_name" label="关联告警名"><Input placeholder="例如: APIHighErrorRate" /></Form.Item>
          <Form.Item name="service" label="服务"><Input placeholder="例如: techmind-server" /></Form.Item>
          <Form.Item name="content" label="内容（Markdown）" rules={[{ required: true }]}>
            <Input.TextArea rows={8} placeholder={"# 告警说明\n...\n## 排查步骤\n..."} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
