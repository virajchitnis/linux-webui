import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Server, Cpu, HardDrive, Activity, CheckCircle, XCircle } from 'lucide-react'

interface SystemInfo {
  hostname: string
  kernel: string
  os: string
  arch: string
  app_version: string
  uptime_seconds: number
  load1: number
  load5: number
  load15: number
  mem_total_kb: number
  mem_used_kb: number
}

interface Capabilities {
  distro: { id: string; family: string; version: string }
  terminal: boolean
  journalctl: boolean
  apt: boolean
  dnf: boolean
  pacman: boolean
  firewall_ufw: boolean
  firewall_firewalld: boolean
  useradd: boolean
  cron: boolean
  docker: boolean
  wireguard: boolean
  tailscale: boolean
  ollama: boolean
}

function fmtUptime(s: number): string {
  const d = Math.floor(s / 86400)
  const h = Math.floor((s % 86400) / 3600)
  const m = Math.floor((s % 3600) / 60)
  return `${d}d ${h}h ${m}m`
}

function fmtMem(kb: number): string {
  return `${(kb / 1024 / 1024).toFixed(1)} GB`
}

export default function About() {
  const { data: info } = useQuery<SystemInfo>({
    queryKey: ['system-info'],
    queryFn: () => api.get('/api/system/info'),
    refetchInterval: 30_000,
  })

  const { data: caps } = useQuery<Capabilities>({
    queryKey: ['capabilities'],
    queryFn: () => api.get('/api/capabilities'),
  })

  const capEntries: [string, boolean][] = caps
    ? [
        ['Terminal (bash + PTY)', caps.terminal],
        ['Journal viewer (journalctl)', caps.journalctl],
        ['APT packages', caps.apt],
        ['DNF/YUM packages', caps.dnf],
        ['Pacman packages', caps.pacman],
        ['Firewall (UFW)', caps.firewall_ufw],
        ['Firewall (firewalld)', caps.firewall_firewalld],
        ['User management (useradd)', caps.useradd],
        ['Cron editor', caps.cron],
        ['Docker', caps.docker],
        ['WireGuard', caps.wireguard],
        ['Tailscale', caps.tailscale],
        ['AI (Ollama)', caps.ollama],
      ]
    : []

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-2xl font-bold text-white">About</h1>
        <p className="text-gray-400 text-sm mt-0.5">System and application information</p>
      </div>

      {/* App info */}
      <section className="bg-gray-900 rounded-xl border border-gray-800 p-5 space-y-4">
        <div className="flex items-center gap-3">
          <Server className="h-6 w-6 text-blue-400" />
          <div>
            <p className="text-white font-semibold text-lg">linux-admin</p>
            <p className="text-gray-500 text-sm">Server administration panel</p>
          </div>
        </div>
        <dl className="grid grid-cols-2 gap-x-8 gap-y-2 text-sm">
          <Row label="Version" value={info?.app_version ?? '…'} />
          <Row label="Hostname" value={info?.hostname ?? '…'} />
          <Row label="Uptime" value={info ? fmtUptime(info.uptime_seconds) : '…'} />
          <Row label="Kernel" value={info?.kernel ?? '…'} />
          <Row label="OS / Arch" value={info ? `${info.os} / ${info.arch}` : '…'} />
          <Row label="Load (1/5/15)" value={info ? `${info.load1.toFixed(2)} / ${info.load5.toFixed(2)} / ${info.load15.toFixed(2)}` : '…'} />
          <Row label="Memory" value={info ? `${fmtMem(info.mem_used_kb)} used of ${fmtMem(info.mem_total_kb)}` : '…'} />
          {caps && <Row label="Distro" value={`${caps.distro.id} (${caps.distro.family}) ${caps.distro.version}`} />}
        </dl>
      </section>

      {/* Capabilities */}
      <section className="bg-gray-900 rounded-xl border border-gray-800 p-5">
        <h2 className="text-sm font-semibold text-gray-400 mb-4 flex items-center gap-2">
          <Activity className="h-4 w-4" />
          Detected capabilities
        </h2>
        <div className="grid grid-cols-2 gap-2">
          {capEntries.map(([label, enabled]) => (
            <div key={label} className="flex items-center gap-2 text-sm">
              {enabled
                ? <CheckCircle className="h-4 w-4 text-green-400 shrink-0" />
                : <XCircle className="h-4 w-4 text-gray-600 shrink-0" />}
              <span className={enabled ? 'text-gray-300' : 'text-gray-600'}>{label}</span>
            </div>
          ))}
        </div>
      </section>

      {/* System stats */}
      <section className="bg-gray-900 rounded-xl border border-gray-800 p-5">
        <h2 className="text-sm font-semibold text-gray-400 mb-4 flex items-center gap-2">
          <HardDrive className="h-4 w-4" />
          Runtime
        </h2>
        <dl className="grid grid-cols-2 gap-x-8 gap-y-2 text-sm">
          <Row label="CPU cores" value={caps ? String(navigator.hardwareConcurrency ?? '–') : '–'} />
          <Row label="Memory total" value={info ? fmtMem(info.mem_total_kb) : '–'} />
          <Row label="Memory used" value={info ? fmtMem(info.mem_used_kb) : '–'} />
        </dl>
      </section>

      <p className="text-xs text-gray-600 text-center">
        <Cpu className="h-3 w-3 inline mr-1" />
        linux-admin is open source software — MIT licensed
      </p>
    </div>
  )
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <>
      <dt className="text-gray-500">{label}</dt>
      <dd className="text-gray-200 font-mono text-xs">{value}</dd>
    </>
  )
}
