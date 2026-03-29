import { BrowserRouter, Routes, Route, Link } from 'react-router-dom'
import MineViewer from './components/MineViewer'
import AlertPanel from './components/AlertPanel'

export default function App() {
  return (
    <BrowserRouter>
      <nav>
        <Link to="/">🗺️ 3D一张图</Link>
        <Link to="/alerts">🚨 报警列表</Link>
      </nav>
      <Routes>
        <Route path="/" element={<MineViewer />} />
        <Route path="/alerts" element={<AlertPanel />} />
      </Routes>
    </BrowserRouter>
  )
}
