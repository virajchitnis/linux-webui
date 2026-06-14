import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { QRCodeSVG } from 'qrcode.react'
import { api, ApiError } from '@/lib/api'
import { useAuth } from '@/hooks/useAuth'
import { Key, Trash2, Monitor, ShieldCheck, ShieldOff, Copy, Check } from 'lucide-react'

interface SessionInfo {
  id: string
  ip: string
  user_agent: string
  created_at: string
  last_seen: string
}

interface TOTPEnroll {
  secret: string
  otpauth_url: string
}

export default function Account() {
  const { data: me } = useAuth()
  const qc = useQueryClient()
  const [toast, setToast] = useState<string | null>(null)
  const notify = (msg: string) => { setToast(msg); setTimeout(() => setToast(null), 3500) }

  // ── Password change ──────────────────────────────────────────────────────
  const [pwd, setPwd] = useState({ current: '', next: '', confirm: '' })
  const { mutate: changePwd, isPending: changingPwd } = useMutation({
    mutationFn: () => api.post('/api/account/password', { current_password: pwd.current, new_password: pwd.next }),
    onSuccess: () => { notify('Password changed'); setPwd({ current: '', next: '', confirm: '' }) },
    onError: (e: Error) => notify(`Error: ${e.message}`),
  })
  const pwdValid = pwd.next.length >= 8 && pwd.next === pwd.confirm && pwd.current.length > 0

  // ── Own sessions ─────────────────────────────────────────────────────────
  const { data: sessions } = useQuery<SessionInfo[]>({
    queryKey: ['own-sessions'],
    queryFn: () => api.get('/api/account/sessions'),
  })
  const { mutate: revokeSession } = useMutation({
    mutationFn: (id: string) => api.delete(`/api/account/sessions/${id}`),
    onSuccess: () => { notify('Session revoked'); qc.invalidateQueries({ queryKey: ['own-sessions'] }) },
    onError: (e: Error) => notify(`Error: ${e.message}`),
  })

  // ── TOTP ─────────────────────────────────────────────────────────────────
  const { data: totpStatus, refetch: refetchTOTP } = useQuery<{ enabled: boolean }>({
    queryKey: ['totp-status'],
    queryFn: () => api.get('/api/account/totp/status'),
  })

  const [enrollData, setEnrollData] = useState<TOTPEnroll | null>(null)
  const [confirmCode, setConfirmCode] = useState('')
  const [confirmError, setConfirmError] = useState('')
  const [recoveryCodes, setRecoveryCodes] = useState<string[]>([])
  const [copiedIdx, setCopiedIdx] = useState<number | null>(null)

  const { mutate: startEnroll, isPending: enrolling } = useMutation({
    mutationFn: () => api.post<TOTPEnroll>('/api/account/totp/enroll'),
    onSuccess: (d) => { setEnrollData(d); setConfirmCode('') },
    onError: (e: Error) => notify(`Error: ${e.message}`),
  })

  const { mutate: confirmTOTP, isPending: confirming } = useMutation({
    mutationFn: () =>
      api.post<{ recovery_codes: string[] }>('/api/account/totp/confirm', {
        secret: enrollData!.secret,
        code: confirmCode,
      }),
    onSuccess: (d) => {
      setRecoveryCodes(d.recovery_codes)
      setEnrollData(null)
      setConfirmCode('')
      setConfirmError('')
      refetchTOTP()
    },
    onError: (e) => setConfirmError(e instanceof ApiError ? e.message : String(e)),
  })

  const { mutate: revokeTOTP, isPending: revokingTOTP } = useMutation({
    mutationFn: () => api.delete('/api/account/totp'),
    onSuccess: () => { notify('2FA disabled'); refetchTOTP() },
    onError: (e: Error) => notify(`Error: ${e.message}`),
  })

  function copyCode(code: string, i: number) {
    navigator.clipboard.writeText(code)
    setCopiedIdx(i)
    setTimeout(() => setCopiedIdx(null), 2000)
  }

  return (
    <div className="space-y-6 max-w-xl p-6">
      <div>
        <h1 className="text-2xl font-bold text-white">Account</h1>
        <p className="text-gray-400 text-sm mt-0.5">{me?.username} · {me?.role}</p>
      </div>

      {/* Change password */}
      <section className="bg-gray-900 rounded-xl border border-gray-800 p-5 space-y-4">
        <h2 className="text-sm font-semibold text-gray-300 flex items-center gap-2">
          <Key className="h-4 w-4" />
          Change password
        </h2>
        <div className="space-y-3">
          <input
            type="password"
            placeholder="Current password"
            value={pwd.current}
            onChange={e => setPwd(p => ({ ...p, current: e.target.value }))}
            className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm focus:outline-none focus:border-blue-500"
          />
          <input
            type="password"
            placeholder="New password (min 8 chars)"
            value={pwd.next}
            onChange={e => setPwd(p => ({ ...p, next: e.target.value }))}
            className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm focus:outline-none focus:border-blue-500"
          />
          <input
            type="password"
            placeholder="Confirm new password"
            value={pwd.confirm}
            onChange={e => setPwd(p => ({ ...p, confirm: e.target.value }))}
            className={`w-full px-3 py-2 bg-gray-800 border rounded-lg text-white text-sm focus:outline-none ${
              pwd.confirm && pwd.confirm !== pwd.next ? 'border-red-500' : 'border-gray-700 focus:border-blue-500'
            }`}
          />
        </div>
        {pwd.confirm && pwd.confirm !== pwd.next && (
          <p className="text-red-400 text-xs">Passwords do not match</p>
        )}
        <button
          onClick={() => changePwd()}
          disabled={changingPwd || !pwdValid}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm rounded-lg disabled:opacity-50 transition-colors"
        >
          {changingPwd ? 'Changing…' : 'Change password'}
        </button>
      </section>

      {/* Two-factor authentication */}
      <section className="bg-gray-900 rounded-xl border border-gray-800 p-5 space-y-4">
        <h2 className="text-sm font-semibold text-gray-300 flex items-center gap-2">
          <ShieldCheck className="h-4 w-4" />
          Two-factor authentication
        </h2>

        {totpStatus?.enabled ? (
          <>
            <p className="text-sm text-green-400">2FA is enabled on your account.</p>
            <button
              onClick={() => { if (window.confirm('Disable two-factor authentication?')) revokeTOTP() }}
              disabled={revokingTOTP}
              className="flex items-center gap-2 px-3 py-2 bg-red-900/30 hover:bg-red-900/50 border border-red-800 text-red-400 text-sm rounded-lg disabled:opacity-50 transition-colors"
            >
              <ShieldOff className="h-4 w-4" />
              {revokingTOTP ? 'Disabling…' : 'Disable 2FA'}
            </button>
          </>
        ) : recoveryCodes.length > 0 ? (
          <div className="space-y-3">
            <p className="text-sm text-green-400 font-medium">2FA enabled! Save your recovery codes:</p>
            <p className="text-xs text-gray-400">Each code can only be used once. Store them securely — they will not be shown again.</p>
            <div className="grid grid-cols-2 gap-1.5">
              {recoveryCodes.map((code, i) => (
                <button
                  key={i}
                  onClick={() => copyCode(code, i)}
                  className="flex items-center justify-between px-3 py-2 bg-gray-800 hover:bg-gray-700 border border-gray-700 rounded-lg font-mono text-sm text-white transition-colors"
                >
                  {code}
                  {copiedIdx === i ? <Check size={12} className="text-green-400" /> : <Copy size={12} className="text-gray-600" />}
                </button>
              ))}
            </div>
            <button
              onClick={() => setRecoveryCodes([])}
              className="text-xs text-gray-500 hover:text-gray-300 transition-colors"
            >
              I've saved these codes — dismiss
            </button>
          </div>
        ) : enrollData ? (
          <div className="space-y-4">
            <p className="text-sm text-gray-300">Scan this QR code with your authenticator app:</p>
            <div className="flex justify-center">
              <div className="bg-white p-3 rounded-lg">
                <QRCodeSVG value={enrollData.otpauth_url} size={180} />
              </div>
            </div>
            <p className="text-xs text-gray-500 text-center font-mono break-all">{enrollData.secret}</p>
            <div>
              <label className="text-xs text-gray-400 mb-1 block">Enter the 6-digit code to confirm</label>
              <input
                type="text"
                inputMode="numeric"
                pattern="[0-9]{6}"
                maxLength={6}
                value={confirmCode}
                onChange={e => { setConfirmCode(e.target.value.trim()); setConfirmError('') }}
                placeholder="000000"
                className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white text-center text-xl font-mono tracking-widest placeholder-gray-600 focus:outline-none focus:border-blue-500"
              />
            </div>
            {confirmError && <p className="text-xs text-red-400">{confirmError}</p>}
            <div className="flex gap-2">
              <button
                onClick={() => confirmTOTP()}
                disabled={confirming || confirmCode.length !== 6}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm rounded-lg disabled:opacity-50 transition-colors"
              >
                {confirming ? 'Verifying…' : 'Activate 2FA'}
              </button>
              <button
                onClick={() => { setEnrollData(null); setConfirmCode('') }}
                className="px-4 py-2 bg-gray-800 hover:bg-gray-700 text-gray-300 text-sm rounded-lg transition-colors"
              >
                Cancel
              </button>
            </div>
          </div>
        ) : (
          <>
            <p className="text-sm text-gray-400">
              Add an extra layer of security with a time-based authenticator app (Google Authenticator, Authy, etc.).
            </p>
            <button
              onClick={() => startEnroll()}
              disabled={enrolling}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm rounded-lg disabled:opacity-50 transition-colors"
            >
              <ShieldCheck className="h-4 w-4" />
              {enrolling ? 'Generating…' : 'Enable 2FA'}
            </button>
          </>
        )}
      </section>

      {/* Active sessions */}
      <section className="bg-gray-900 rounded-xl border border-gray-800 overflow-hidden">
        <div className="px-5 py-4 border-b border-gray-800">
          <h2 className="text-sm font-semibold text-gray-300 flex items-center gap-2">
            <Monitor className="h-4 w-4" />
            Active sessions
          </h2>
        </div>
        {!sessions?.length && (
          <div className="px-5 py-4 text-gray-500 text-sm">No sessions found.</div>
        )}
        {sessions?.map(s => (
          <div key={s.id} className="flex items-start justify-between px-5 py-3 border-b border-gray-800/50 last:border-0">
            <div className="space-y-0.5 text-xs">
              <p className="text-gray-300 font-mono">{s.ip}</p>
              <p className="text-gray-500 truncate max-w-xs">{s.user_agent}</p>
              <p className="text-gray-600">Last seen: {new Date(s.last_seen).toLocaleString()}</p>
            </div>
            <button
              onClick={() => revokeSession(s.id)}
              title="Revoke session"
              className="p-1.5 rounded hover:bg-red-900/40 text-gray-500 hover:text-red-400 transition-colors shrink-0"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </button>
          </div>
        ))}
      </section>

      {toast && (
        <div className="fixed bottom-4 right-4 bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-sm text-white shadow-xl z-50">
          {toast}
        </div>
      )}
    </div>
  )
}
