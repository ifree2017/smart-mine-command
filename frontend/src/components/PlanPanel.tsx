import { useState } from 'react'

interface Plan {
  id: string
  name: string
  level: number
  commands: any[]
}

export default function PlanPanel() {
  const [plans] = useState<Plan[]>([
    { id: '1', name: '一级报警预案', level: 4, commands: [] },
    { id: '2', name: '二级报警预案', level: 3, commands: [] },
    { id: '3', name: '三级报警预案', level: 2, commands: [] },
  ])

  return (
    <div className="plan-panel">
      <h2>📋 数字预案管理</h2>
      <div className="plans-grid">
        {plans.map(plan => (
          <div key={plan.id} className={`plan-card level-${plan.level}`}>
            <h3>{plan.name}</h3>
            <p>触发等级：⭐ x {plan.level}</p>
            <div className="plan-actions">
              <button>▶️ 触发演练</button>
              <button>✏️ 编辑</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
