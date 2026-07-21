import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useEffect } from 'react'
import type { ReactNode } from 'react'
import ForumLayout from '@/layouts/ForumLayout'
import AdminLayout from '@/layouts/AdminLayout'
import Login from '@/pages/forum/Login'
import Home from '@/pages/forum/Home'
import Search from '@/pages/forum/Search'
import ArticleDetail from '@/pages/forum/ArticleDetail'
import ArticleEditor from '@/pages/forum/ArticleEditor'
import UserProfile from '@/pages/forum/UserProfile'
import NotFound from '@/pages/NotFound'
import Forbidden from '@/pages/Forbidden'
import Monitor from '@/pages/admin/Monitor'
import MonitorSlow from '@/pages/admin/MonitorSlow'
import MonitorErrors from '@/pages/admin/MonitorErrors'
import MonitorQueues from '@/pages/admin/MonitorQueues'
import MonitorAI from '@/pages/admin/MonitorAI'
import AlertList from '@/pages/admin/AlertList'
import AlertDetail from '@/pages/admin/AlertDetail'
import OpsReports from '@/pages/admin/OpsReports'
import OpsReportDetail from '@/pages/admin/OpsReportDetail'
import OpsDiagnose from '@/pages/admin/OpsDiagnose'
import Runbooks from '@/pages/admin/Runbooks'
import Deployments from '@/pages/admin/Deployments'
import Incidents from '@/pages/admin/Incidents'
import IncidentDetail from '@/pages/admin/IncidentDetail'
import { useAuthStore } from '@/store/auth'
import LoginModal from '@/components/forum/LoginModal'

function AdminGuard({ children }: { children: ReactNode }) {
  const user = useAuthStore((s) => s.user)
  const initialized = useAuthStore((s) => s.initialized)
  if (!initialized) return null
  if (!user) return <Navigate to="/login" replace />
  if (user.role !== 1) return <Navigate to="/403" replace />
  return <>{children}</>
}

export default function App() {
  const init = useAuthStore((s) => s.init)
  useEffect(() => { init() }, [init])
  return (
    <BrowserRouter>
      <Routes>
        {/* Forum routes */}
        <Route element={<ForumLayout />}>
          <Route path="/" element={<Home />} />
          <Route path="/login" element={<Login />} />
          <Route path="/search" element={<Search />} />
          <Route path="/articles/new" element={<ArticleEditor />} />
          <Route path="/articles/:id" element={<ArticleDetail />} />
          <Route path="/articles/:id/edit" element={<ArticleEditor />} />
          <Route path="/user/profile" element={<UserProfile />} />
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
          <Route path="/admin/monitor" element={<Monitor />} />
          <Route path="/admin/monitor/slow" element={<MonitorSlow />} />
          <Route path="/admin/monitor/errors" element={<MonitorErrors />} />
          <Route path="/admin/monitor/queues" element={<MonitorQueues />} />
          <Route path="/admin/monitor/ai" element={<MonitorAI />} />
          <Route path="/admin/alerts" element={<AlertList />} />
          <Route path="/admin/alerts/:id" element={<AlertDetail />} />
          <Route path="/admin/ops/reports" element={<OpsReports />} />
          <Route path="/admin/ops/reports/:id" element={<OpsReportDetail />} />
          <Route path="/admin/ops/diagnose" element={<OpsDiagnose />} />
          <Route path="/admin/incidents" element={<Incidents />} />
          <Route path="/admin/incidents/:id" element={<IncidentDetail />} />
          <Route path="/admin/runbooks" element={<Runbooks />} />
          <Route path="/admin/deployments" element={<Deployments />} />
        </Route>

        {/* Error pages */}
        <Route path="/403" element={<Forbidden />} />
        <Route path="*" element={<NotFound />} />
      </Routes>
      <LoginModal />
    </BrowserRouter>
  )
}
