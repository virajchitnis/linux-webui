import { useQuery } from '@tanstack/react-query'
import { Wifi, Globe, ArrowUp, ArrowDown, Activity } from 'lucide-react'
import { api } from '@/lib/api'

interface NetInterface {
  name: string
  state: string
  type: string
  addresses: string[]
  mtu: number
  rx_bytes: number
  tx_bytes: number
  rx_packets: number
  tx_packets: number
  mac: string
}

function fmtBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  const units = ['KB', 'MB', 'GB', 'TB']
  let val = bytes / 1024
  let i = 0
  while (val >= 1024 && i < units.length - 1) { val /= 1024; i++ }
  return `${val.toFixed(1)} ${units[i]}`
}

function StateChip({ state }: { state: string }) {
  const isUp = state === 'up'
  return (
    <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${
      isUp ? 'bg-green-900/40 text-green-400' : 'bg-gray-800 text-gray-500'
    }`}>
      <span className={`w-1.5 h-1.5 rounded-full ${isUp ? 'bg-green-400' : 'bg-gray-600'}`} />
      {state}
    </span>
  )
}

function TypeIcon({ type }: { type: string }) {
  const cls = 'text-gray-500'
  if (type === 'wifi') return <Wifi size={16} className={cls} />
  if (type === 'loopback') return <Activity size={16} className={cls} />
  return <Globe size={16} className={cls} />
}

export default function Network() {
  const { data: interfaces, isLoading, error, dataUpdatedAt } = useQuery<NetInterface[]>({
    queryKey: ['network-interfaces'],
    queryFn: () => api.get<NetInterface[]>('/api/network/interfaces'),
    refetchInterval: 5_000,
  })

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-white">Network</h1>
          <p className="text-sm text-gray-400 mt-1">
            Network interfaces and traffic statistics
            {dataUpdatedAt > 0 && (
              <span className="ml-2 text-gray-600">· refreshes every 5s</span>
            )}
          </p>
        </div>
      </div>

      {isLoading && (
        <div className="p-8 text-center text-gray-500">Loading…</div>
      )}
      {error && (
        <div className="p-4 bg-red-900/20 border border-red-800 rounded-lg text-sm text-red-400">
          {String(error)}
        </div>
      )}

      <div className="grid gap-3">
        {interfaces?.map((iface) => (
          <div
            key={iface.name}
            className="bg-gray-900 border border-gray-800 rounded-lg p-4"
          >
            <div className="flex items-start justify-between gap-4">
              <div className="flex items-center gap-3">
                <TypeIcon type={iface.type} />
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-sm font-medium text-white">{iface.name}</span>
                    <StateChip state={iface.state} />
                    <span className="text-xs text-gray-600 capitalize">{iface.type}</span>
                  </div>
                  <p className="text-xs text-gray-500 font-mono mt-0.5">{iface.mac || '—'}</p>
                </div>
              </div>
              <div className="text-right text-xs text-gray-500">
                <p>MTU {iface.mtu}</p>
              </div>
            </div>

            {/* Addresses */}
            {iface.addresses && iface.addresses.length > 0 && (
              <div className="mt-2 flex flex-wrap gap-1.5">
                {iface.addresses.map((addr) => (
                  <span
                    key={addr}
                    className="px-2 py-0.5 bg-gray-800 rounded text-xs font-mono text-gray-300"
                  >
                    {addr}
                  </span>
                ))}
              </div>
            )}

            {/* Traffic stats */}
            <div className="mt-3 grid grid-cols-2 gap-3 pt-3 border-t border-gray-800">
              <div className="flex items-center gap-2">
                <ArrowDown size={14} className="text-blue-400 shrink-0" />
                <div>
                  <p className="text-xs text-gray-500">RX</p>
                  <p className="text-sm font-mono text-white">{fmtBytes(iface.rx_bytes)}</p>
                  <p className="text-xs text-gray-600">{iface.rx_packets.toLocaleString()} pkts</p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <ArrowUp size={14} className="text-green-400 shrink-0" />
                <div>
                  <p className="text-xs text-gray-500">TX</p>
                  <p className="text-sm font-mono text-white">{fmtBytes(iface.tx_bytes)}</p>
                  <p className="text-xs text-gray-600">{iface.tx_packets.toLocaleString()} pkts</p>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>

      {interfaces?.length === 0 && (
        <div className="p-8 text-center text-gray-600 text-sm">No interfaces found</div>
      )}
    </div>
  )
}
