import { Button, Result } from 'antd'
import { useNavigate } from 'react-router-dom'

export default function Forbidden() {
  const navigate = useNavigate()
  return (
    <Result
      status="403"
      title="403"
      subTitle="抱歉，您没有权限访问该页面。"
      extra={
        <Button type="primary" onClick={() => navigate('/')}>
          回首页
        </Button>
      }
    />
  )
}
