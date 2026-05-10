import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Plus, Trash2, ShieldCheck, ShieldOff } from 'lucide-react'

interface FWRule {
  num: number
  to: string
  action: string
  from: string
}

interface FWStatus {
  enabled: boolean
  rules: FWRule[]
}

const ACTION_COLORS: Record<string, string> = {
  ALLOW: 'bg-green-500/20 text-green-400 border-green-500/30',
  DENY:  'bg-red-500/20 text-red-400 border-red-500/30',
  REJECT:'bg-orange-500/20 text-orange-400 border-orange-500/30',
  LIMIT: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
}

export default function Firewall() {
  const [toast, setToast] = useState<string | null>(null)
  const [addOpen, setAddOpen] = useState(false)
  const [form, setForm] = useState({ action: 'allow', port: '', proto: 'tcp' })
  const qc = useQueryClient()

  const { data: status, isLoading, error } = useQuery<FWStatus>({
    queryKey: ['firewall'],
    queryFn: () => api.get('/api/firewall/status'),
    refetchInterval: 10_000,
  })

  const notify = (msg: string) => { setToast(msg); setTimeout(() => setToast(null), 3500) }

  const { mutate: addRule, isPending: adding } = useMutation({
    mutationFn: () => api.post('/api/firewall/rules', form),
    onSuccess: () => {
      notify(`Rule ${form.action} ${form.port}/${form.proto} added`)
      setAddOpen(false)
      setForm({ action: 'allow', port: '', proto: 'tcp' })
      qc.invalidateQueries({ queryKey: ['firewall'] })
    },
    onError: (e: Error) => notify(`Error: ${e.message}`),
  })

  const { mutate: deleteRule } = useMutation({
    mutationFn: (num: number) => api.delete(`/api/firewall/rules/${num}`),
    onSuccess: () => { notify('Rule deleted'); qc.invalidateQueries({ queryKey: ['firewall'] }) },
    onError: (e: Error) => notify(`Error: ${e.message}`),
  })

  const { mutate: toggle } = useMutation({
    mutationFn: () => status?.enabled ? api.post('/api/firewall/disable', {}) : api.post('/api/firewall/enable', {}),
    onSuccess: () => { notify(status?.enabled ? 'UFW disabled' : 'UFW enabled'); qc.invalidateQueries({ queryKey: ['firewall'] }) },
    onError: (e: Error) => notify(`Error: ${e.message}`),
  })

  const confirmDelete = (num: number) => {
    if (window.confirm(`Delete rule #${num}?`)) deleteRule(num)
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Firewall</h1>
          <p className="text-gray-400 text-sm mt-0.5">UFW (Uncomplicated Firewall)</p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setAddOpen(true)}
            className="flex items-center gap-2 px-3 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm rounded-lg transition-colors"
          >
            <Plus className="h-4 w-4" />
            Add rule
          </button>
          <button
            onClick={() => toggle()}
            className={`flex items-center gap-2 px-3 py-2 text-sm rounded-lg border transition-colors ${
              status?.enabled
                ? 'bg-red-900/30 border-red-800/50 text-red-400 hover:bg-red-900/50'
                : 'bg-green-900/30 border-green-800/50 text-green-400 hover:bg-green-900/50'
            }`}
          >
            {status?.enabled ? <ShieldOff className="h-4 w-4" /> : <ShieldCheck className="h-4 w-4" />}
            {status?.enabled ? 'Disable UFW' : 'Enable UFW'}
          </button>
        </div>
      </div>

      {/* Status banner */}
      {status && (
        <div className={`flex items-center gap-3 px-4 py-3 rounded-xl border text-sm ${
          status.enabled
            ? 'bg-green-900/20 border-green-800/50 text-green-400'
            : 'bg-gray-900 border-gray-800 text-gray-500'
        }`}>
          {status.enabled ? <ShieldCheck className="h-4 w-4" /> : <ShieldOff className="h-4 w-4" />}
          UFW is {status.enabled ? 'active' : 'inactive'} — {status.rules?.length ?? 0} rules
        </div>
      )}

      {error && (
        <div className="bg-red-900/20 border border-red-800 rounded-xl p-4 text-red-400 text-sm">
          UFW not available or insufficient permissions.
        </div>
      )}

      {/* Add rule dialog */}
      {addOpen && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-900 border border-gray-700 rounded-xl p-6 w-80 space-y-4">
            <h2 className="text-white font-semibold">Add firewall rule</h2>
            <div className="space-y-3">
              <select
                value={form.action}
                onChange={e => setForm(f => ({ ...f, action: e.target.value }))}
                className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm"
              >
                <option value="allow">Allow</option>
                <option value="deny">Deny</option>
                <option value="reject">Reject</option>
                <option value="limit">Limit (rate-limit)</option>
              </select>
              <input
                autoFocus
                placeholder="Port (e.g. 80 or 8000:9000)"
                value={form.port}
                onChange={e => setForm(f => ({ ...f, port: e.target.value }))}
                className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm focus:outline-none focus:border-blue-500"
              />
              <select
                value={form.proto}
                onChange={e => setForm(f => ({ ...f, proto: e.target.value }))}
                className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm"
              >
                <option value="tcp">TCP</option>
                <option value="udp">UDP</option>
                <option value="any">Any</option>
              </select>
            </div>
            <div className="flex justify-end gap-2">
              <button onClick={() => setAddOpen(false)} className="px-3 py-2 text-sm text-gray-400 hover:text-white">Cancel</button>
              <button
                onClick={() => addRule()}
                disabled={adding || !form.port}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm rounded-lg disabled:opacity-50"
              >
                {adding ? 'Adding…' : 'Add rule'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Rules table */}
      <div className="bg-gray-900 rounded-xl border border-gray-800 overflow-hidden">
        {isLoading && <div className="px-4 py-8 text-center text-gray-500">Loading rules…</div>}
        {!isLoading && !status?.rules?.length && !error && (
          <div className="px-4 py-8 text-center text-gray-500 text-sm">No rules configured.</div>
        )}
        {status?.rules?.length ? (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-xs text-gray-500 uppercase tracking-wide border-b border-gray-800">
                <th className="text-left px-4 py-2 font-medium w-10">#</th>
                <th className="text-left px-4 py-2 font-medium">To</th>
                <th className="text-left px-4 py-2 font-medium">Action</th>
                <th className="text-left px-4 py-2 font-medium">From</th>
                <th className="text-right px-4 py-2 font-medium">Delete</th>
              </tr>
            </thead>
            <tbody>
              {status.rules.map(rule => (
                <tr key={rule.num} className="border-b border-gray-800/50 hover:bg-gray-800/20">
                  <td className="px-4 py-2 text-gray-600">{rule.num}</td>
                  <td className="px-4 py-2 font-mono text-xs text-gray-300">{rule.to}</td>
                  <td className="px-4 py-2">
                    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${ACTION_COLORS[rule.action] ?? 'text-gray-400'}`}>
                      {rule.action}
                    </span>
                  </td>
                  <td className="px-4 py-2 text-gray-500 text-xs">{rule.from}</td>
                  <td className="px-4 py-2 text-right">
                    <button onClick={() => confirmDelete(rule.num)}
                      className="p-1.5 rounded hover:bg-red-900/40 text-gray-400 hover:text-red-400 transition-colors">
                      <Trash2 className="h-3.5 w-3.5" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : null}
      </div>

      {toast && (
        <div className="fixed bottom-4 right-4 bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-sm text-white shadow-xl">
          {toast}
        </div>
      )}
    </div>
  )
}
