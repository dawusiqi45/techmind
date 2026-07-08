import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import type { ReactNode } from 'react'
import ForumLayout from '@/layouts/ForumLayout'
import AdminLayout from '@/layouts/AdminLayout'
import Login from '@/pages/forum/Login'
import NotFound from '@/pages/NotFound'
import Forbidden from '@/pages/Forbidden'
import { useAuthStore } from '@/store/auth'

// 占位，后续 Task 会替换
const Placeholder = ({ name }: { name: string }) => (
  <div style={{ padding: 40, color: '#888' }}>{name} (待实现)</div>
)

function AdminGuard({ children }: { children: ReactNode }) {
  const user = useAuthStore((s) => s.user)
  if (!user) return <Navigate to="/login" replace />
  if (user.role !== 1) return <Navigate to="/403" replace />
  return <>{children}</>
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Forum routes */}
        <Route element={<ForumLayout />}>
          <Route path="/" element={<Placeholder name="Home" />} />
          <Route path="/login" element={<Login />} />
          <Route path="/search" element={<Placeholder name="Search" />} />
          <Route path="/articles/new" element={<Placeholder name="ArticleEditor" />} />
          <Route path="/articles/:id" element={<Placeholder name="ArticleDetail" />} />
          <Route path="/articles/:id/edit" element={<Placeholder name="ArticleEditor (edit)" />} />
          <Route path="/user/profile" element={<Placeholder name="UserProfile" />} />
        </Route>

        {/* Admin routes */}
        <Route
          element={
            <AdminGuard>
              <AdminLayout />
            </AdminGuard>
          }
        >
          <Route path="/admin" element={<Navigate to="/admin/monitor" replace />} />
          <Route path="/admin/monitor" element={<Placeholder name="Monitor" />} />
          <Route path="/admin/monitor/slow" element={<Placeholder name="MonitorSlow" />} />
          <Route path="/admin/monitor/errors" element={<Placeholder name="MonitorErrors" />} />
          <Route path="/admin/monitor/queues" element={<Placeholder name="MonitorQueues" />} />
          <Route path="/admin/monitor/ai" element={<Placeholder name="MonitorAI" />} />
          <Route path="/admin/alerts" element={<Placeholder name="AlertList" />} />
          <Route path="/admin/alerts/:id" element={<Placeholder name="AlertDetail" />} />
          <Route path="/admin/ops/reports" element={<Placeholder name="OpsReports" />} />
          <Route path="/admin/ops/reports/:id" element={<Placeholder name="OpsReportDetail" />} />
          <Route path="/admin/ops/diagnose" element={<Placeholder name="OpsDiagnose" />} />
          <Route path="/admin/runbooks" element={<Placeholder name="Runbooks" />} />
          <Route path="/admin/deployments" element={<Placeholder name="Deployments" />} />
        </Route>

        {/* Error pages */}
        <Route path="/403" element={<Forbidden />} />
        <Route path="*" element={<NotFound />} />
      </Routes>
    </BrowserRouter>
  )
}
