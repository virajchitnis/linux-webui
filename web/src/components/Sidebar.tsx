import { NavLink } from 'react-router-dom'
import {
  LayoutDashboard, Server, Terminal, ScrollText, Package,
  Activity, Shield, Users, FolderOpen, Clock, Bot,
  Network, Container, Wifi, GitBranch, Info, LogOut,
} from 'lucide-react'
import { cn } from '@/lib/cn'
import type { Capabilities } from '@/hooks/useCapabilities'

interface NavItem {
  to: string
  label: string
  icon: React.ReactNode
  enabled: boolean
  tooltip?: string
}

interface Props {
  caps: Capabilities | undefined
  role: string
  username: string
  onLogout: () => void
}

export default function Sidebar({ caps, role, username, onLogout }: Props) {
  const isAdmin = role === 'admin'

  const items: NavItem[] = [
    { to: '/dashboard', label: 'Dashboard', icon: <LayoutDashboard size={16} />, enabled: true },
    { to: '/services', label: 'Services', icon: <Server size={16} />, enabled: true },
    { to: '/terminal', label: 'Terminal', icon: <Terminal size={16} />, enabled: isAdmin && (caps?.terminal ?? false), tooltip: !caps?.terminal ? 'bash not found' : !isAdmin ? 'Admin only' : undefined },
    { to: '/logs', label: 'Logs', icon: <ScrollText size={16} />, enabled: caps?.journalctl ?? false, tooltip: !caps?.journalctl ? 'journalctl not found' : undefined },
    { to: '/packages', label: 'Packages', icon: <Package size={16} />, enabled: (caps?.apt ?? caps?.dnf ?? caps?.pacman ?? false) },
    { to: '/processes', label: 'Processes', icon: <Activity size={16} />, enabled: true },
    { to: '/firewall', label: 'Firewall', icon: <Shield size={16} />, enabled: (caps?.firewall_ufw ?? caps?.firewall_firewalld ?? false), tooltip: 'Install ufw or firewalld to enable' },
    { to: '/users', label: 'Users', icon: <Users size={16} />, enabled: isAdmin && (caps?.useradd ?? false) },
    { to: '/files', label: 'Files', icon: <FolderOpen size={16} />, enabled: true },
    { to: '/cron', label: 'Cron Jobs', icon: <Clock size={16} />, enabled: caps?.cron ?? false, tooltip: 'Install crontab to enable' },
    { to: '/docker', label: 'Docker', icon: <Container size={16} />, enabled: caps?.docker ?? false, tooltip: 'Docker socket not readable' },
    { to: '/network', label: 'Network', icon: <Network size={16} />, enabled: true },
    { to: '/wireguard', label: 'WireGuard', icon: <Wifi size={16} />, enabled: caps?.wireguard ?? false, tooltip: 'Install wg to enable' },
    { to: '/tailscale', label: 'Tailscale', icon: <GitBranch size={16} />, enabled: caps?.tailscale ?? false, tooltip: 'Install tailscale to enable' },
    { to: '/ai', label: 'AI Assistant', icon: <Bot size={16} />, enabled: caps?.ollama ?? false, tooltip: 'Install Ollama to enable' },
    { to: '/about', label: 'About', icon: <Info size={16} />, enabled: true },
  ]

  return (
    <aside className="flex flex-col h-screen w-60 bg-gray-900 border-r border-gray-800 shrink-0">
      <div className="flex items-center gap-2 px-4 py-5 border-b border-gray-800">
        <Server size={20} className="text-blue-400" />
        <span className="font-semibold text-white text-sm">linux-admin</span>
      </div>

      <nav className="flex-1 overflow-y-auto py-2">
        {items.map((item) => (
          <NavItem key={item.to} {...item} />
        ))}
      </nav>

      <div className="border-t border-gray-800 p-3">
        <div className="text-xs text-gray-500 mb-2 px-2">
          {username} · {role}
        </div>
        <button
          onClick={onLogout}
          className="flex items-center gap-2 w-full px-3 py-2 text-sm text-gray-400 hover:text-white hover:bg-gray-800 rounded-lg transition-colors"
        >
          <LogOut size={16} />
          Sign out
        </button>
      </div>
    </aside>
  )
}

function NavItem({ to, label, icon, enabled, tooltip }: NavItem) {
  if (!enabled) {
    return (
      <div
        title={tooltip}
        className="flex items-center gap-2 px-4 py-2 text-sm text-gray-600 cursor-not-allowed"
      >
        {icon}
        <span>{label}</span>
      </div>
    )
  }
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        cn(
          'flex items-center gap-2 px-4 py-2 text-sm transition-colors',
          isActive
            ? 'text-white bg-gray-800 border-r-2 border-blue-500'
            : 'text-gray-400 hover:text-white hover:bg-gray-800/50',
        )
      }
    >
      {icon}
      <span>{label}</span>
    </NavLink>
  )
}
