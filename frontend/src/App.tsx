import { BrowserRouter, Routes, Route, Link } from 'react-router-dom'
import MineViewer from './components/MineViewer'
import AlertPanel from './components/AlertPanel'
import CommandPanel from './components/CommandPanel'
import PlanPanel from './components/PlanPanel'
import AIPanel from './components/AIPanel'

export default function App() {
  return (
    <BrowserRouter>
      <nav>
        <Link to="/">🗺️ 3D一张图</Link>
        <Link to="/alerts">🚨 报警</Link>
        <Link to="/commands">📢 指令</Link>
        <Link to="/plans">📋 预案</Link>
        <Link to="/ai">🤖 AI分析</Link>
      </nav>
      <Routes>
        <Route path="/" element={<MineViewer />} />
        <Route path="/alerts" element={<AlertPanel />} />
        <Route path="/commands" element={<CommandPanel />} />
        <Route path="/plans" element={<PlanPanel />} />
        <Route path="/ai" element={<AIPanel />} />
      </Routes>
    </BrowserRouter>
  )
}
