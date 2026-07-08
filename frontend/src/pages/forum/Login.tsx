import { useState } from 'react'
import { Form, Input, Button, Tabs, message } from 'antd'
import { useNavigate } from 'react-router-dom'
import { authApi } from '@/api/auth'
import { useAuthStore } from '@/store/auth'
import styles from './Login.module.css'

interface LoginFormValues {
  username: string
  password: string
}

interface RegisterFormValues {
  username: string
  email: string
  password: string
}

export default function Login() {
  const navigate = useNavigate()
  const login = useAuthStore((s) => s.login)
  const [activeTab, setActiveTab] = useState('login')
  const [loginLoading, setLoginLoading] = useState(false)
  const [registerLoading, setRegisterLoading] = useState(false)

  const handleLogin = async (values: LoginFormValues) => {
    setLoginLoading(true)
    try {
      const res = await authApi.login(values)
      const { access_token, refresh_token } = (res as { data: { access_token: string; refresh_token: string } }).data
      const profileRes = await authApi.getProfile()
      const user = (profileRes as { data: { id: number; username: string; role: number; avatar: string } }).data
      login(access_token, refresh_token, user)
      navigate('/')
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } }
      message.error(error?.response?.data?.message ?? '登录失败，请重试')
    } finally {
      setLoginLoading(false)
    }
  }

  const handleRegister = async (values: RegisterFormValues) => {
    setRegisterLoading(true)
    try {
      await authApi.register(values)
      message.success('注册成功，请登录')
      setActiveTab('login')
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } }
      message.error(error?.response?.data?.message ?? '注册失败，请重试')
    } finally {
      setRegisterLoading(false)
    }
  }

  const loginForm = (
    <Form layout="vertical" onFinish={handleLogin} autoComplete="off">
      <Form.Item
        label="用户名"
        name="username"
        rules={[{ required: true, message: '请输入用户名' }]}
      >
        <Input placeholder="请输入用户名" />
      </Form.Item>
      <Form.Item
        label="密码"
        name="password"
        rules={[{ required: true, message: '请输入密码' }]}
      >
        <Input.Password placeholder="请输入密码" />
      </Form.Item>
      <Form.Item>
        <Button type="primary" htmlType="submit" loading={loginLoading} block>
          登录
        </Button>
      </Form.Item>
    </Form>
  )

  const registerForm = (
    <Form layout="vertical" onFinish={handleRegister} autoComplete="off">
      <Form.Item
        label="用户名"
        name="username"
        rules={[{ required: true, message: '请输入用户名' }]}
      >
        <Input placeholder="请输入用户名" />
      </Form.Item>
      <Form.Item
        label="邮箱"
        name="email"
        rules={[
          { required: true, message: '请输入邮箱' },
          { type: 'email', message: '邮箱格式不正确' },
        ]}
      >
        <Input placeholder="请输入邮箱" />
      </Form.Item>
      <Form.Item
        label="密码"
        name="password"
        rules={[
          { required: true, message: '请输入密码' },
          { min: 6, message: '密码至少 6 位' },
        ]}
      >
        <Input.Password placeholder="请输入密码（至少 6 位）" />
      </Form.Item>
      <Form.Item>
        <Button type="primary" htmlType="submit" loading={registerLoading} block>
          注册
        </Button>
      </Form.Item>
    </Form>
  )

  const items = [
    { key: 'login', label: '登录', children: loginForm },
    { key: 'register', label: '注册', children: registerForm },
  ]

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={items}
          centered
        />
      </div>
    </div>
  )
}
