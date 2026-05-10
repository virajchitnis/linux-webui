import { useEffect } from 'react'
import { Routes, Route, Navigate, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import Layout from '@/components/Layout'
import Login from '@/pages/Login'
import Setup from '@/pages/Setup'
import Dashboard from '@/pages/Dashboard'
import Services from '@/pages/Services'
import Terminal from '@/pages/Terminal'
import Logs from '@/pages/Logs'
import Packages from '@/pages/Packages'
import Processes from '@/pages/Processes'
import Firewall from '@/pages/Firewall'
import Users from '@/pages/Users'
import Files from '@/pages/Files'
import Cron from '@/pages/Cron'
import Docker from '@/pages/Docker'
import Network from '@/pages/Network'
import WireGuard from '@/pages/WireGuard'
import Tailscale from '@/pages/Tailscale'
import AI from '@/pages/AI'
import About from '@/pages/About'
import Account from '@/pages/Account'
import AuditLog from '@/pages/AuditLog'

function SetupGuard({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate()
  const { data, isLoading } = useQuery<{ complete: boolean }>({
    queryKey: ['setup-status'],
    queryFn: () => api.get('/api/setup/status'),
  })

  useEffect(() => {
    if (!isLoading && data && !data.complete) {
      navigate('/setup', { replace: true })
    }
  }, [data, isLoading, navigate])

  if (isLoading) return null
  return <>{children}</>
}

export default function App() {
  return (
    <SetupGuard>
      <Routes>
        <Route path="/setup" element={<Setup />} />
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<Layout />}>
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="services" element={<Services />} />
          <Route path="terminal" element={<Terminal />} />
          <Route path="logs" element={<Logs />} />
          <Route path="packages" element={<Packages />} />
          <Route path="processes" element={<Processes />} />
          <Route path="firewall" element={<Firewall />} />
          <Route path="users" element={<Users />} />
          <Route path="files" element={<Files />} />
          <Route path="cron" element={<Cron />} />
          <Route path="docker" element={<Docker />} />
          <Route path="network" element={<Network />} />
          <Route path="wireguard" element={<WireGuard />} />
          <Route path="tailscale" element={<Tailscale />} />
          <Route path="ai" element={<AI />} />
          <Route path="about" element={<About />} />
          <Route path="account" element={<Account />} />
          <Route path="audit" element={<AuditLog />} />
        </Route>
      </Routes>
    </SetupGuard>
  )
}
