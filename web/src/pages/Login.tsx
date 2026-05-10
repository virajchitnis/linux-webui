import { useState, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/hooks/useAuth'
import { api } from '@/lib/api'

export default function Login() {
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { data: me } = useAuth()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [totpCode, setTotpCode] = useState('')
  const [totpRequired, setTotpRequired] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const totpRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (me) navigate('/dashboard', { replace: true })
  }, [me, navigate])

  useEffect(() => {
    if (totpRequired) totpRef.current?.focus()
  }, [totpRequired])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const result = await api.post<Record<string, unknown>>('/api/auth/login', {
        username,
        password,
        ...(totpRequired ? { totp_code: totpCode } : {}),
      })
      if (result && 'totp_required' in result) {
        setTotpRequired(true)
        setLoading(false)
        return
      }
      qc.setQueryData(['me'], result)
      navigate('/dashboard', { replace: true })
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-950">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-white">linux-webui</h1>
          <p className="text-gray-400 mt-2">Server Administration Panel</p>
        </div>

        <form onSubmit={handleSubmit} className="bg-gray-900 rounded-xl p-8 shadow-2xl border border-gray-800">
          {!totpRequired ? (
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">Username</label>
                <input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  autoComplete="username"
                  required
                  className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">Password</label>
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  autoComplete="current-password"
                  required
                  className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>

              {error && (
                <div className="text-red-400 text-sm bg-red-950 border border-red-800 rounded-lg px-3 py-2">
                  {error}
                </div>
              )}

              <button
                type="submit"
                disabled={loading}
                className="w-full bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white font-medium py-2 px-4 rounded-lg transition-colors"
              >
                {loading ? 'Signing in…' : 'Sign in'}
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              <div className="text-center text-sm text-gray-400 mb-2">
                <p className="text-white font-medium mb-1">Two-factor authentication</p>
                <p>Enter the 6-digit code from your authenticator app, or a recovery code.</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">Authentication code</label>
                <input
                  ref={totpRef}
                  type="text"
                  inputMode="numeric"
                  pattern="[0-9A-Z\-]{6,20}"
                  value={totpCode}
                  onChange={(e) => setTotpCode(e.target.value.trim())}
                  autoComplete="one-time-code"
                  placeholder="000000"
                  className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white text-center text-2xl font-mono tracking-widest placeholder-gray-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>

              {error && (
                <div className="text-red-400 text-sm bg-red-950 border border-red-800 rounded-lg px-3 py-2">
                  {error}
                </div>
              )}

              <button
                type="submit"
                disabled={loading || totpCode.length < 6}
                className="w-full bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white font-medium py-2 px-4 rounded-lg transition-colors"
              >
                {loading ? 'Verifying…' : 'Verify'}
              </button>
              <button
                type="button"
                onClick={() => { setTotpRequired(false); setTotpCode(''); setError('') }}
                className="w-full text-gray-500 hover:text-gray-300 text-sm transition-colors"
              >
                ← Back
              </button>
            </div>
          )}
        </form>
      </div>
    </div>
  )
}
