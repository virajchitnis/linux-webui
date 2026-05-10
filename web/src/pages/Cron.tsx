import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2, Clock } from 'lucide-react'
import cronstrue from 'cronstrue'
import { api, ApiError } from '@/lib/api'

interface CronEntry {
  schedule: string
  command: string
  raw: string
  index: number
}

function describeSchedule(schedule: string): string {
  try {
    return cronstrue.toString(schedule)
  } catch {
    return schedule
  }
}

export default function Cron() {
  const qc = useQueryClient()
  const [showAdd, setShowAdd] = useState(false)
  const [schedule, setSchedule] = useState('0 * * * *')
  const [command, setCommand] = useState('')
  const [addError, setAddError] = useState('')

  const { data: entries, isLoading } = useQuery<CronEntry[]>({
    queryKey: ['cron'],
    queryFn: () => api.get<CronEntry[]>('/api/cron'),
    refetchInterval: 30_000,
  })

  const addMutation = useMutation({
    mutationFn: () => api.post('/api/cron', { schedule, command }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['cron'] })
      setShowAdd(false)
      setSchedule('0 * * * *')
      setCommand('')
      setAddError('')
    },
    onError: (e) => setAddError(e instanceof ApiError ? e.message : String(e)),
  })

  const deleteMutation = useMutation({
    mutationFn: (index: number) => api.delete(`/api/cron/${index}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['cron'] }),
  })

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-white">Cron Jobs</h1>
          <p className="text-sm text-gray-400 mt-1">Manage your personal crontab</p>
        </div>
        <button
          onClick={() => setShowAdd(true)}
          className="flex items-center gap-2 px-3 py-1.5 bg-blue-600 hover:bg-blue-500 text-white text-sm rounded-lg transition-colors"
        >
          <Plus size={14} />
          Add Job
        </button>
      </div>

      {showAdd && (
        <div className="bg-gray-900 border border-gray-700 rounded-lg p-4 space-y-3">
          <h3 className="text-sm font-medium text-white">New Cron Job</h3>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-xs text-gray-400 mb-1 block">Schedule (5 fields)</label>
              <input
                type="text"
                value={schedule}
                onChange={(e) => setSchedule(e.target.value)}
                placeholder="* * * * *"
                className="w-full bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm text-white font-mono focus:outline-none focus:border-blue-500"
              />
              <p className="text-xs text-gray-500 mt-1">{describeSchedule(schedule)}</p>
            </div>
            <div>
              <label className="text-xs text-gray-400 mb-1 block">Command</label>
              <input
                type="text"
                value={command}
                onChange={(e) => setCommand(e.target.value)}
                placeholder="/usr/local/bin/backup.sh"
                className="w-full bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm text-white font-mono focus:outline-none focus:border-blue-500"
              />
            </div>
          </div>
          {addError && <p className="text-xs text-red-400">{addError}</p>}
          <div className="flex gap-2">
            <button
              onClick={() => addMutation.mutate()}
              disabled={addMutation.isPending || !command.trim()}
              className="px-3 py-1.5 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-sm rounded-lg transition-colors"
            >
              {addMutation.isPending ? 'Adding…' : 'Add'}
            </button>
            <button
              onClick={() => { setShowAdd(false); setAddError('') }}
              className="px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-gray-300 text-sm rounded-lg transition-colors"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      <div className="bg-gray-900 border border-gray-800 rounded-lg overflow-hidden">
        {isLoading && (
          <div className="p-8 text-center text-gray-500">Loading…</div>
        )}
        {!isLoading && (!entries || entries.length === 0) && (
          <div className="p-8 text-center">
            <Clock size={32} className="text-gray-700 mx-auto mb-3" />
            <p className="text-gray-500 text-sm">No cron jobs yet</p>
          </div>
        )}
        {entries?.map((entry) => (
          <div
            key={entry.index}
            className="flex items-center gap-3 px-4 py-3 border-b border-gray-800/50 last:border-0"
          >
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <span className="font-mono text-xs text-blue-400 shrink-0">{entry.schedule}</span>
                <span className="text-gray-600 text-xs">·</span>
                <span className="text-xs text-gray-400">{describeSchedule(entry.schedule)}</span>
              </div>
              <p className="font-mono text-sm text-gray-200 mt-0.5 truncate">{entry.command}</p>
            </div>
            <button
              onClick={() => {
                if (window.confirm(`Delete cron job: ${entry.command}?`))
                  deleteMutation.mutate(entry.index)
              }}
              className="p-1.5 text-gray-600 hover:text-red-400 transition-colors shrink-0"
              title="Delete"
            >
              <Trash2 size={14} />
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}
