import { useEffect } from 'react'
import { Outlet, useNavigate } from 'react-router-dom'
import Sidebar from './Sidebar'
import { useAuth, useLogout } from '@/hooks/useAuth'
import { useCapabilities } from '@/hooks/useCapabilities'

export default function Layout() {
  const navigate = useNavigate()
  const { data: me, isLoading } = useAuth()
  const logout = useLogout()
  const { data: caps } = useCapabilities()

  useEffect(() => {
    if (!isLoading && !me) {
      navigate('/login', { replace: true })
    }
  }, [me, isLoading, navigate])

  if (isLoading || !me) return null

  function handleLogout() {
    logout.mutate(undefined, {
      onSettled: () => navigate('/login', { replace: true }),
    })
  }

  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar
        caps={caps}
        role={me.role}
        username={me.username}
        onLogout={handleLogout}
      />
      <main className="flex-1 overflow-y-auto bg-gray-950 p-6">
        <Outlet />
      </main>
    </div>
  )
}
