import { useState } from 'react'

interface AnalysisResult {
  summary: string
  risk_level: number
  factors: string[]
  recommendation: string
}

interface Plan {
  id: string
  name: string
  trigger: string
  commands: any[]
}

export default function AIPanel() {
  const [form, setForm] = useState({
    location: '',
    gasType: 'T',
    value: '',
    level: 2,
  })
  const [analyzing, setAnalyzing] = useState(false)
  const [result, setResult] = useState<AnalysisResult | null>(null)
  const [plans, setPlans] = useState<Plan[]>([])

  async function handleAnalyze(e: React.FormEvent) {
    e.preventDefault()
    setAnalyzing(true)
    setResult(null)
    setPlans([])

    try {
      const r = await fetch('/api/ai/analyze', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          location: form.location,
          gasType: form.gasType,
          value: parseFloat(form.value),
          level: form.level,
        }),
      })
      const d = await r.json()
      setResult(d)

      const r2 = await fetch('/api/ai/recommend', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          location: form.location,
          gasType: form.gasType,
          value: parseFloat(form.value),
          level: form.level,
        }),
      })
      const d2 = await r2.json()
      setPlans(d2.plans || [])
    } catch (err) {
      console.error(err)
    } finally {
      setAnalyzing(false)
    }
  }

  return (
    <div className="ai-panel">
      <h2>🤖 AI安全分析</h2>

      <form className="ai-form" onSubmit={handleAnalyze}>
        <div className="form-row">
          <label>位置：</label>
          <input
            value={form.location}
            onChange={e => setForm({...form, location: e.target.value})}
            placeholder="如：一号井东巷"
          />
        </div>
        <div className="form-row">
          <label>气体类型：</label>
          <select value={form.gasType} onChange={e => setForm({...form, gasType: e.target.value})}>
            <option value="T">甲烷(T)</option>
            <option value="CO">一氧化碳(CO)</option>
            <option value="CO2">二氧化碳(CO2)</option>
            <option value="O2">氧气(O2)</option>
          </select>
        </div>
        <div className="form-row">
          <label>浓度值：</label>
          <input
            type="number"
            step="0.01"
            value={form.value}
            onChange={e => setForm({...form, value: e.target.value})}
            placeholder="如：0.5"
          />
        </div>
        <div className="form-row">
          <label>风险等级：</label>
          <select value={form.level} onChange={e => setForm({...form, level: parseInt(e.target.value)})}>
            <option value="1">⭐ 低危</option>
            <option value="2">⭐⭐ 中危</option>
            <option value="3">⭐⭐⭐ 高危</option>
            <option value="4">⭐⭐⭐⭐ 严重</option>
            <option value="5">⭐⭐⭐⭐⭐ 特大</option>
          </select>
        </div>
        <button type="submit" disabled={analyzing || !form.location || !form.value}>
          {analyzing ? '🔍 分析中...' : '🔬 AI分析'}
        </button>
      </form>

      {result && (
        <div className={`analysis-result level-${result.risk_level}`}>
          <div className="result-header">
            <h3>📊 AI分析结果</h3>
            <span className="risk-badge">{renderRiskLevel(result.risk_level)}</span>
          </div>
          <p className="summary">{result.summary}</p>
          <div className="factors">
            <strong>影响因素：</strong>
            <ul>
              {result.factors?.map((f, i) => <li key={i}>{f}</li>)}
            </ul>
          </div>
          <div className="recommendation">
            <strong>💡 建议：</strong>{result.recommendation}
          </div>
        </div>
      )}

      {plans.length > 0 && (
        <div className="recommended-plans">
          <h3>📋 推荐预案</h3>
          {plans.map((plan, i) => (
            <div key={i} className="plan-item">
              <h4>{plan.name}</h4>
              <p>触发条件：{plan.trigger}</p>
              <div className="plan-commands">
                {plan.commands?.map((cmd: any, j: number) => (
                  <span key={j} className="cmd-tag">{cmd.type}</span>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

function renderRiskLevel(level: number) {
  const labels = ['', '低危', '中危', '高危', '严重', '特大']
  const colors = ['', '#2ecc71', '#f1c40f', '#e67e22', '#e74c3c', '#8e44ad']
  return <span style={{ color: colors[level], fontWeight: 'bold' }}>{labels[level]}</span>
}
