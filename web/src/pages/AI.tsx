import { useState, useRef, useEffect, useCallback } from 'react'
import { Send, Bot, User, Loader2, Copy, Check } from 'lucide-react'

interface Message {
  role: 'user' | 'assistant'
  content: string
  streaming?: boolean
}

function CodeBlock({ code }: { code: string }) {
  const [copied, setCopied] = useState(false)
  const copy = () => {
    navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }
  return (
    <div className="relative group my-2">
      <pre className="bg-gray-950 border border-gray-700 rounded-lg p-3 text-xs font-mono text-gray-200 overflow-x-auto whitespace-pre-wrap">
        {code}
      </pre>
      <button
        onClick={copy}
        className="absolute top-2 right-2 p-1 text-gray-600 hover:text-gray-300 transition-colors opacity-0 group-hover:opacity-100"
        title="Copy"
      >
        {copied ? <Check size={12} className="text-green-400" /> : <Copy size={12} />}
      </button>
    </div>
  )
}

function MessageContent({ content }: { content: string }) {
  const parts = content.split(/(```[\s\S]*?```)/g)
  return (
    <div>
      {parts.map((part, i) => {
        if (part.startsWith('```') && part.endsWith('```')) {
          const code = part.replace(/^```[^\n]*\n?/, '').replace(/\n?```$/, '')
          return <CodeBlock key={i} code={code} />
        }
        return <span key={i} className="whitespace-pre-wrap">{part}</span>
      })}
    </div>
  )
}

export default function AI() {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [connected, setConnected] = useState(false)
  const [connecting, setConnecting] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)
  const bottomRef = useRef<HTMLDivElement>(null)
  const isStreaming = useRef(false)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const connect = useCallback(() => {
    if (wsRef.current) return
    setConnecting(true)
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const ws = new WebSocket(`${proto}//${location.host}/ws/ai`)
    wsRef.current = ws

    ws.onopen = () => { setConnected(true); setConnecting(false) }
    ws.onclose = () => {
      setConnected(false)
      setConnecting(false)
      wsRef.current = null
    }
    ws.onerror = () => {
      setConnected(false)
      setConnecting(false)
      wsRef.current = null
    }
    ws.onmessage = (e) => {
      const msg = JSON.parse(e.data)
      if (msg.type === 'token') {
        setMessages((prev) => {
          const last = prev[prev.length - 1]
          if (last?.role === 'assistant' && last.streaming) {
            return [...prev.slice(0, -1), { ...last, content: last.content + msg.token }]
          }
          return [...prev, { role: 'assistant', content: msg.token, streaming: true }]
        })
      } else if (msg.type === 'done') {
        isStreaming.current = false
        setMessages((prev) => {
          const last = prev[prev.length - 1]
          if (last?.role === 'assistant' && last.streaming) {
            return [...prev.slice(0, -1), { ...last, streaming: false }]
          }
          return prev
        })
      }
    }
  }, [])

  useEffect(() => {
    connect()
    return () => {
      wsRef.current?.close()
      wsRef.current = null
    }
  }, [connect])

  function send() {
    const prompt = input.trim()
    if (!prompt || !connected || isStreaming.current) return
    setInput('')
    isStreaming.current = true
    setMessages((prev) => [...prev, { role: 'user', content: prompt }])
    wsRef.current?.send(JSON.stringify({ type: 'generate', prompt }))
  }

  function handleKey(e: React.KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      send()
    }
  }

  return (
    <div className="flex flex-col h-[calc(100vh-64px)] p-6">
      <div className="mb-4">
        <h1 className="text-xl font-semibold text-white">AI Assistant</h1>
        <p className="text-sm text-gray-400 mt-1">
          Powered by Ollama — ask about your server, logs, or commands.
          {connecting && <span className="ml-2 text-gray-600">Connecting…</span>}
          {connected && <span className="ml-2 text-green-500">● Connected</span>}
          {!connected && !connecting && (
            <span className="ml-2 text-red-500">● Disconnected —{' '}
              <button onClick={connect} className="underline hover:no-underline">reconnect</button>
            </span>
          )}
        </p>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto space-y-4 pr-1">
        {messages.length === 0 && (
          <div className="flex flex-col items-center justify-center h-full text-gray-600">
            <Bot size={48} className="mb-4" />
            <p className="text-sm">Ask about your server, analyze logs, or get command suggestions.</p>
          </div>
        )}
        {messages.map((msg, i) => (
          <div key={i} className={`flex gap-3 ${msg.role === 'user' ? 'justify-end' : ''}`}>
            {msg.role === 'assistant' && (
              <div className="w-7 h-7 rounded-full bg-blue-900 flex items-center justify-center shrink-0 mt-0.5">
                <Bot size={14} className="text-blue-400" />
              </div>
            )}
            <div
              className={`max-w-[80%] rounded-xl px-4 py-3 text-sm ${
                msg.role === 'user'
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-800 text-gray-200'
              }`}
            >
              {msg.role === 'assistant' ? (
                <MessageContent content={msg.content} />
              ) : (
                <p className="whitespace-pre-wrap">{msg.content}</p>
              )}
              {msg.streaming && (
                <Loader2 size={12} className="inline animate-spin ml-1 text-gray-500" />
              )}
            </div>
            {msg.role === 'user' && (
              <div className="w-7 h-7 rounded-full bg-gray-700 flex items-center justify-center shrink-0 mt-0.5">
                <User size={14} className="text-gray-400" />
              </div>
            )}
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      {/* Input */}
      <div className="mt-4 flex gap-2">
        <textarea
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKey}
          placeholder="Ask about your server…"
          rows={2}
          disabled={!connected || isStreaming.current}
          className="flex-1 bg-gray-800 border border-gray-700 rounded-xl px-4 py-2.5 text-sm text-white placeholder-gray-600 resize-none focus:outline-none focus:border-blue-500 disabled:opacity-50"
        />
        <button
          onClick={send}
          disabled={!connected || !input.trim() || isStreaming.current}
          className="self-end px-4 py-2.5 bg-blue-600 hover:bg-blue-500 disabled:opacity-40 text-white rounded-xl transition-colors"
          title="Send (Enter)"
        >
          <Send size={16} />
        </button>
      </div>
      <p className="text-xs text-gray-700 mt-2 text-center">
        AI suggestions are for reference only — verify before executing commands.
      </p>
    </div>
  )
}
