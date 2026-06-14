import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useMetrics } from '@/hooks/useMetrics'
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'
import { Power } from 'lucide-react'

interface SystemInfo {
  hostname: string
  kernel: string
  app_version: string
  uptime_seconds: number
  load1: number
  load5: number
  load15: number
  mem_total_kb: number
  mem_used_kb: number
  swap_total_kb: number
  swap_used_kb: number
}

function fmtUptime(seconds: number): string {
  const d = Math.floor(seconds / 86400)
  const h = Math.floor((seconds % 86400) / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  return d > 0 ? `${d}d ${h}h ${m}m` : h > 0 ? `${h}h ${m}m` : `${m}m`
}

function fmtBytes(kb: number): string {
  if (kb > 1024 * 1024) return `${(kb / 1024 / 1024).toFixed(1)} GB`
  return `${(kb / 1024).toFixed(0)} MB`
}

function useReboot() {
  const [rebooting, setRebooting] = useState(false)
  const trigger = async () => {
    if (!window.confirm('Reboot the system now? All active sessions will be terminated.')) return
    setRebooting(true)
    try {
      await api.post('/api/system/reboot', {})
    } catch {
      // Expected — server goes down immediately
    }
  }
  return { trigger, rebooting }
}

export default function Dashboard() {
  const { trigger: reboot, rebooting } = useReboot()
  const { data: info } = useQuery<SystemInfo>({
    queryKey: ['system-info'],
    queryFn: () => api.get('/api/system/info'),
    refetchInterval: 10_000,
  })
  const { latest, history } = useMetrics()

  const chartData = history.map((s) => ({
    time: new Date(s.ts).toLocaleTimeString(),
    cpu: +s.cpu_pct.toFixed(1),
    mem: +((s.mem_used_kb / s.mem_total_kb) * 100).toFixed(1),
  }))

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-white">Dashboard</h1>
        <p className="text-gray-400 text-sm mt-1">{info?.hostname ?? '…'} · {info?.kernel ?? ''}</p>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatCard label="CPU Usage" value={latest ? `${latest.cpu_pct.toFixed(1)}%` : '…'} />
        <StatCard label="Memory" value={latest ? `${fmtBytes(latest.mem_used_kb)} / ${fmtBytes(latest.mem_total_kb)}` : '…'} />
        <StatCard label="Load Avg" value={latest ? `${latest.load1.toFixed(2)}` : '…'} sub="1 min" />
        <StatCard label="Uptime" value={info ? fmtUptime(info.uptime_seconds) : '…'} />
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <ChartCard title="CPU %" color="#3b82f6" dataKey="cpu" data={chartData} />
        <ChartCard title="Memory %" color="#10b981" dataKey="mem" data={chartData} />
      </div>

      {/* System details */}
      {info && (
        <div className="bg-gray-900 rounded-xl border border-gray-800 p-4">
          <h2 className="text-sm font-semibold text-gray-400 mb-3">System Details</h2>
          <dl className="grid grid-cols-2 gap-x-8 gap-y-2 text-sm">
            <Detail label="Hostname" value={info.hostname} />
            <Detail label="Kernel" value={info.kernel} />
            <Detail label="app version" value={info.app_version} />
            <Detail label="Load (1/5/15)" value={`${info.load1.toFixed(2)} / ${info.load5.toFixed(2)} / ${info.load15.toFixed(2)}`} />
            <Detail label="Memory" value={`${fmtBytes(info.mem_used_kb)} used of ${fmtBytes(info.mem_total_kb)}`} />
            <Detail label="Swap" value={info.swap_total_kb > 0 ? `${fmtBytes(info.swap_used_kb)} used of ${fmtBytes(info.swap_total_kb)}` : 'none'} />
          </dl>
        </div>
      )}

      {/* Quick actions */}
      <div className="bg-gray-900 rounded-xl border border-gray-800 p-4">
        <h2 className="text-sm font-semibold text-gray-400 mb-3">Quick Actions</h2>
        <button
          onClick={reboot}
          disabled={rebooting}
          className="flex items-center gap-2 px-3 py-2 bg-red-900/30 hover:bg-red-900/50 border border-red-800/50 text-red-400 hover:text-red-300 text-sm rounded-lg disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          <Power className="h-4 w-4" />
          {rebooting ? 'Rebooting…' : 'Reboot system'}
        </button>
      </div>
    </div>
  )
}

function StatCard({ label, value, sub }: { label: string; value: string; sub?: string }) {
  return (
    <div className="bg-gray-900 rounded-xl border border-gray-800 p-4">
      <p className="text-xs text-gray-500 font-medium uppercase tracking-wide">{label}</p>
      <p className="text-2xl font-bold text-white mt-1">{value}</p>
      {sub && <p className="text-xs text-gray-500 mt-0.5">{sub}</p>}
    </div>
  )
}

function ChartCard({ title, color, dataKey, data }: { title: string; color: string; dataKey: string; data: Record<string, unknown>[] }) {
  return (
    <div className="bg-gray-900 rounded-xl border border-gray-800 p-4">
      <h3 className="text-sm font-semibold text-gray-300 mb-3">{title}</h3>
      <ResponsiveContainer width="100%" height={140}>
        <AreaChart data={data}>
          <defs>
            <linearGradient id={`grad-${dataKey}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor={color} stopOpacity={0.3} />
              <stop offset="95%" stopColor={color} stopOpacity={0} />
            </linearGradient>
          </defs>
          <XAxis dataKey="time" hide />
          <YAxis domain={[0, 100]} hide />
          <Tooltip
            contentStyle={{ background: '#1f2937', border: '1px solid #374151', borderRadius: 8, fontSize: 12 }}
            labelStyle={{ color: '#9ca3af' }}
          />
          <Area type="monotone" dataKey={dataKey} stroke={color} fill={`url(#grad-${dataKey})`} strokeWidth={2} dot={false} />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  )
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <>
      <dt className="text-gray-500">{label}</dt>
      <dd className="text-gray-200 font-mono text-xs">{value}</dd>
    </>
  )
}
