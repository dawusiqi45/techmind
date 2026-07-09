import { useState } from 'react'
import { Form, Input, Tabs, message } from 'antd'
import { useNavigate } from 'react-router-dom'
import { authApi } from '@/api/auth'
import { useAuthStore } from '@/store/auth'
import styles from './Login.module.css'

export default function Login() {
  const navigate = useNavigate()
  const login = useAuthStore((s) => s.login)
  const [activeTab, setActiveTab] = useState('login')
  const [loginLoading, setLoginLoading] = useState(false)
  const [registerLoading, setRegisterLoading] = useState(false)

  const handleLogin = async (values: { username: string; password: string }) => {
    setLoginLoading(true)
    try {
      const res = await authApi.login(values)
      const { access_token, refresh_token } = res.data.data as { access_token: string; refresh_token: string }
      const profileRes = await authApi.getProfile()
      const user = profileRes.data.data as { id: number; username: string; role: number; avatar: string }
      login(access_token, refresh_token, user)
      navigate('/')
    } catch (err: unknown) {
      const error = err as { response?: { data?: { msg?: string } } }
      message.error(error?.response?.data?.msg ?? '登录失败，请重试')
    } finally {
      setLoginLoading(false)
    }
  }

  const handleRegister = async (values: { username: string; email: string; password: string }) => {
    setRegisterLoading(true)
    try {
      await authApi.register(values)
      message.success('注册成功，请登录')
      setActiveTab('login')
    } catch (err: unknown) {
      const error = err as { response?: { data?: { msg?: string } } }
      message.error(error?.response?.data?.msg ?? '注册失败，请重试')
    } finally {
      setRegisterLoading(false)
    }
  }

  const loginForm = (
    <Form layout="vertical" onFinish={handleLogin} autoComplete="off" className={styles.form}>
      <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
        <Input placeholder="用户名" size="large" className={styles.input} />
      </Form.Item>
      <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
        <Input.Password placeholder="密码" size="large" className={styles.input} />
      </Form.Item>
      <button
        type="submit"
        className={styles.submitBtn}
        disabled={loginLoading}
      >
        {loginLoading ? '登录中...' : '登录'}
      </button>
    </Form>
  )

  const registerForm = (
    <Form layout="vertical" onFinish={handleRegister} autoComplete="off" className={styles.form}>
      <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
        <Input placeholder="用户名" size="large" className={styles.input} />
      </Form.Item>
      <Form.Item name="email" rules={[{ required: true }, { type: 'email', message: '邮箱格式不正确' }]}>
        <Input placeholder="邮箱" size="large" className={styles.input} />
      </Form.Item>
      <Form.Item name="password" rules={[{ required: true }, { min: 6, message: '密码至少 6 位' }]}>
        <Input.Password placeholder="密码（至少 6 位）" size="large" className={styles.input} />
      </Form.Item>
      <button
        type="submit"
        className={styles.submitBtn}
        disabled={registerLoading}
      >
        {registerLoading ? '注册中...' : '注册'}
      </button>
    </Form>
  )

  return (
    <div className={styles.page}>
      <div className={styles.card}>
        <div className={styles.brand}>
          <span className={styles.brandIcon}>⬡</span>
          <span className={styles.brandName}>TechMind</span>
        </div>
        <p className={styles.tagline}>开发者的技术交流社区</p>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          centered
          items={[
            { key: 'login', label: '登录', children: loginForm },
            { key: 'register', label: '注册', children: registerForm },
          ]}
        />
      </div>
    </div>
  )
}
