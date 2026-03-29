import { useEffect, useState } from 'react'

export interface Alert {
  id: string
  location: string
  gasType: string
  value: number
  level: number
  status: string
  createdAt: string
}

export default function AlertPanel() {
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/alerts')
      .then(r => r.json())
      .then(d => {
        setAlerts(d.alerts || d || [])
        setLoading(false)
      })
      .catch(() => {
        // fallback: try mock data for demo
        setAlerts([])
        setLoading(false)
      })
  }, [])

  async function handleAck(id: string) {
    try {
      await fetch(`/api/alerts/ack/${id}`, { method: 'PUT' })
      setAlerts(prev => prev.map(a => a.id === id ? { ...a, status: 'acknowledged' } : a))
    } catch {}
  }

  if (loading) {
    return <div className="alert-panel"><div className="loading">加载中...</div></div>
  }

  return (
    <div className="alert-panel">
      <h2>🚨 报警列表</h2>
      {alerts.length === 0 ? (
        <div style={{ color: '#888', padding: '20px 0' }}>暂无报警记录</div>
      ) : (
        <table>
          <thead>
            <tr>
              <th>位置</th>
              <th>气体类型</th>
              <th>数值</th>
              <th>等级</th>
              <th>时间</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {alerts.map(a => (
              <tr key={a.id} className={`level-${a.level}`}>
                <td>{a.location}</td>
                <td>{a.gasType}</td>
                <td style={{ fontFamily: 'monospace' }}>{a.value}</td>
                <td>{'⭐'.repeat(a.level)}</td>
                <td>{new Date(a.createdAt).toLocaleString()}</td>
                <td>
                  <span style={{
                    padding: '2px 8px', borderRadius: 4, fontSize: 12,
                    background: a.status === 'open' ? '#7a1a1a' : '#2d5a2d',
                    color: '#fff'
                  }}>
                    {a.status === 'open' ? '未处理' : '已确认'}
                  </span>
                </td>
                <td>
                  {a.status === 'open' && (
                    <button onClick={() => handleAck(a.id)}>确认</button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
