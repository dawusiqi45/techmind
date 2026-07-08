import { useEffect, useState } from 'react'
import { Table, Button, Modal, Form, Input, Select, message, Tag } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { deploymentApi } from '@/api/deployment'
import { formatDateTime } from '@/utils/time'

export default function Deployments() {
  const [data, setData] = useState<any[]>([])
  const [open, setOpen] = useState(false)
  const [form] = Form.useForm()

  const load = () => deploymentApi.list().then(r => setData(r.data.data || []))
  useEffect(() => { load() }, [])

  async function handleCreate(values: any) {
    try {
      await deploymentApi.create(values)
      message.success('变更记录已保存')
      setOpen(false)
      form.resetFields()
      load()
    } catch {
      message.error('保存失败')
    }
  }

  return (
    <div>
      <Button type="primary" icon={<PlusOutlined />} onClick={() => setOpen(true)} style={{ marginBottom: 16 }}>记录变更</Button>
      <Table
        dataSource={data} rowKey="id" size="small"
        columns={[
          { title: '服务', dataIndex: 'service' },
          { title: '命名空间', dataIndex: 'namespace' },
          { title: '镜像', dataIndex: 'image', ellipsis: true },
          { title: '旧镜像', dataIndex: 'old_image', ellipsis: true },
          { title: '来源', dataIndex: 'source', render: (v) => <Tag>{v}</Tag>, width: 100 },
          { title: '操作人', dataIndex: 'changed_by' },
          { title: '变更时间', dataIndex: 'changed_at', render: formatDateTime },
        ]}
      />
      <Modal title="记录部署变更" open={open} onCancel={() => setOpen(false)} onOk={() => form.submit()} okText="保存">
        <Form form={form} layout="vertical" onFinish={handleCreate} initialValues={{ namespace: 'default', source: 'manual' }}>
          <Form.Item name="service" label="服务名" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="namespace" label="命名空间"><Input /></Form.Item>
          <Form.Item name="image" label="新镜像" rules={[{ required: true }]}><Input placeholder="techmind:v1.2.0" /></Form.Item>
          <Form.Item name="old_image" label="旧镜像"><Input placeholder="techmind:v1.1.0" /></Form.Item>
          <Form.Item name="changed_by" label="操作人"><Input /></Form.Item>
          <Form.Item name="source" label="来源">
            <Select options={['manual','helm','kubectl','argocd'].map(v => ({ value: v, label: v }))} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
