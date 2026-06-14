import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ChevronRight, Folder, FileText, File, Link, ArrowLeft } from 'lucide-react'
import { api } from '@/lib/api'

interface Entry {
  name: string
  path: string
  is_dir: boolean
  size: number
  mode: string
  mod_time: string
  symlink?: string
}

function fmtSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  const units = ['KB', 'MB', 'GB']
  let val = bytes / 1024
  let i = 0
  while (val >= 1024 && i < units.length - 1) { val /= 1024; i++ }
  return `${val.toFixed(1)} ${units[i]}`
}

export default function Files() {
  const [path, setPath] = useState('/')
  const [viewFile, setViewFile] = useState<string | null>(null)

  const { data: roots } = useQuery<string[]>({
    queryKey: ['files-roots'],
    queryFn: () => api.get<string[]>('/api/files/roots'),
  })

  const { data: entries, isLoading, error } = useQuery<Entry[]>({
    queryKey: ['files-list', path],
    queryFn: () => api.get<Entry[]>(`/api/files/list?path=${encodeURIComponent(path)}`),
    enabled: viewFile === null,
  })

  const { data: fileContent, isLoading: fileLoading } = useQuery<string>({
    queryKey: ['files-read', viewFile],
    queryFn: async () => {
      const resp = await fetch(`/api/files/read?path=${encodeURIComponent(viewFile!)}`, {
        credentials: 'same-origin',
      })
      if (!resp.ok) throw new Error(await resp.text())
      return resp.text()
    },
    enabled: viewFile !== null,
  })

  const pathParts = path.split('/').filter(Boolean)

  function navigate(newPath: string) {
    setViewFile(null)
    setPath(newPath || '/')
  }

  function goUp() {
    const parts = path.split('/').filter(Boolean)
    parts.pop()
    setViewFile(null)
    setPath('/' + parts.join('/') || '/')
  }

  return (
    <div className="p-6 space-y-4">
      <div>
        <h1 className="text-xl font-semibold text-white">File Browser</h1>
        <p className="text-sm text-gray-400 mt-1">
          Read-only access to {roots?.join(', ')}
        </p>
      </div>

      {/* Breadcrumb */}
      <div className="flex items-center gap-1 text-sm text-gray-400 flex-wrap">
        <button onClick={() => navigate('/')} className="hover:text-white transition-colors">/</button>
        {pathParts.map((part, i) => {
          const to = '/' + pathParts.slice(0, i + 1).join('/')
          return (
            <span key={i} className="flex items-center gap-1">
              <ChevronRight size={14} className="text-gray-600" />
              <button onClick={() => navigate(to)} className="hover:text-white transition-colors">{part}</button>
            </span>
          )
        })}
        {viewFile && (
          <span className="flex items-center gap-1">
            <ChevronRight size={14} className="text-gray-600" />
            <span className="text-white">{viewFile.split('/').pop()}</span>
          </span>
        )}
      </div>

      {viewFile ? (
        <div className="bg-gray-900 border border-gray-800 rounded-lg overflow-hidden">
          <div className="flex items-center gap-2 px-4 py-3 border-b border-gray-800">
            <button
              onClick={() => setViewFile(null)}
              className="flex items-center gap-1 text-sm text-gray-400 hover:text-white transition-colors"
            >
              <ArrowLeft size={14} />
              Back
            </button>
            <span className="text-sm text-gray-300 ml-2 font-mono">{viewFile}</span>
          </div>
          {fileLoading ? (
            <div className="p-8 text-center text-gray-500">Loading…</div>
          ) : (
            <pre className="p-4 text-xs text-gray-300 overflow-auto max-h-[600px] font-mono whitespace-pre-wrap break-all">
              {fileContent}
            </pre>
          )}
        </div>
      ) : (
        <div className="bg-gray-900 border border-gray-800 rounded-lg overflow-hidden">
          {path !== '/' && (
            <button
              onClick={goUp}
              className="flex items-center gap-2 w-full px-4 py-2.5 text-sm text-gray-400 hover:bg-gray-800 border-b border-gray-800/50 transition-colors"
            >
              <ArrowLeft size={14} />
              <span>..</span>
            </button>
          )}
          {isLoading && (
            <div className="p-8 text-center text-gray-500">Loading…</div>
          )}
          {error && (
            <div className="p-4 text-sm text-red-400">{String(error)}</div>
          )}
          {entries?.map((entry) => (
            <button
              key={entry.path}
              onClick={() => {
                if (entry.is_dir) navigate(entry.path)
                else setViewFile(entry.path)
              }}
              className="flex items-center gap-3 w-full px-4 py-2.5 text-sm text-gray-300 hover:bg-gray-800 border-b border-gray-800/30 transition-colors text-left"
            >
              <span className="shrink-0">
                {entry.is_dir ? (
                  <Folder size={16} className="text-blue-400" />
                ) : entry.symlink ? (
                  <Link size={16} className="text-yellow-400" />
                ) : entry.name.includes('.') ? (
                  <FileText size={16} className="text-gray-500" />
                ) : (
                  <File size={16} className="text-gray-500" />
                )}
              </span>
              <span className="flex-1 truncate font-mono">
                {entry.name}
                {entry.symlink && <span className="text-gray-600 ml-1 not-italic">→ {entry.symlink}</span>}
              </span>
              <span className="text-gray-600 text-xs font-mono shrink-0">{entry.mode}</span>
              {!entry.is_dir && (
                <span className="text-gray-600 text-xs shrink-0">{fmtSize(entry.size)}</span>
              )}
            </button>
          ))}
          {entries?.length === 0 && (
            <div className="p-8 text-center text-gray-600 text-sm">Empty directory</div>
          )}
        </div>
      )}
    </div>
  )
}
