import { useState } from 'react'

interface Command {
  id: string
  type: 'broadcast' | 'ws' | 'call'
  target: string
  content: string
  status: 'pending' | 'executing' | 'done' | 'failed'
  createdAt: string
}

export default function CommandPanel() {
  const [commands, setCommands] = useState<Command[]>([])
  const [form, setForm] = useState({ type: 'broadcast', target: 'all', content: '' })
  const [loading, setLoading] = useState(false)

  async function fetchCommands() {
    const r = await fetch('/api/commands')
    const d = await r.json()
    setCommands(d.commands || [])
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    try {
      const r = await fetch('/api/commands', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(form),
      })
      const d = await r.json()
      if (d.command) {
        setCommands(prev => [d.command, ...prev])
        setForm({ ...form, content: '' })
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="command-panel">
      <div className="panel-header">
        <h2>📢 融合通信指令</h2>
        <button onClick={fetchCommands} className="btn-refresh">🔄 刷新</button>
      </div>

      {/* 发送表单 */}
      <form className="command-form" onSubmit={handleSubmit}>
        <div className="form-row">
          <label>指令类型：</label>
          <select value={form.type} onChange={e => setForm({...form, type: e.target.value})}>
            <option value="broadcast">📢 广播</option>
            <option value="ws">💬 WS推送</option>
            <option value="call">📞 呼叫</option>
          </select>
        </div>
        <div className="form-row">
          <label>目标：</label>
          <input
            value={form.target}
            onChange={e => setForm({...form, target: e.target.value})}
            placeholder="all / 群组ID / 人员ID"
          />
        </div>
        <div className="form-row">
          <label>内容：</label>
          <textarea
            value={form.content}
            onChange={e => setForm({...form, content: e.target.value})}
            placeholder="输入指令内容..."
            rows={3}
          />
        </div>
        <button type="submit" disabled={loading || !form.content}>
          {loading ? '发送中...' : '🚀 发送指令'}
        </button>
      </form>

      {/* 指令历史 */}
      <div className="command-history">
        <h3>指令历史</h3>
        {commands.length === 0 ? (
          <p className="empty">暂无指令记录</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>类型</th><th>目标</th><th>内容</th><th>状态</th><th>时间</th>
              </tr>
            </thead>
            <tbody>
              {commands.map(cmd => (
                <tr key={cmd.id}>
                  <td>{typeIcon(cmd.type)} {cmd.type}</td>
                  <td>{cmd.target}</td>
                  <td className="content">{cmd.content}</td>
                  <td><span className={`status ${cmd.status}`}>{statusText(cmd.status)}</span></td>
                  <td>{new Date(cmd.createdAt).toLocaleTimeString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}

function typeIcon(type_: string) {
  switch(type_) {
    case 'broadcast': return '📢'
    case 'ws': return '💬'
    case 'call': return '📞'
    default: return '📝'
  }
}

function statusText(status: string) {
  const map: Record<string, string> = {
    pending: '⏳ 待执行',
    executing: '⚙️ 执行中',
    done: '✅ 完成',
    failed: '❌ 失败',
  }
  return map[status] || status
}
