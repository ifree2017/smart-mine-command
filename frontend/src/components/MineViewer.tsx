import { useEffect, useRef, useState, useCallback } from 'react'
import * as THREE from 'three'

const MONITOR_POINTS = [
  { id: 'A01', name: '主井东巷', x: -45, z: -35 },
  { id: 'A02', name: '主井西巷', x: 45, z: -35 },
  { id: 'B01', name: '副井底车场', x: -25, z: 5 },
  { id: 'B02', name: '回风巷道', x: 25, z: 5 },
  { id: 'C01', name: '运输大巷', x: -45, z: -60 },
  { id: 'C02', name: '采煤工作面①', x: 45, z: -60 },
  { id: 'D01', name: '中央变电所', x: 0, z: -25 },
  { id: 'D02', name: '水泵房', x: 0, z: 25 },
  { id: 'E01', name: '避难硐室', x: -55, z: 0 },
  { id: 'E02', name: '炸药库', x: 55, z: 0 },
  { id: 'F01', name: '采煤工作面②', x: -35, z: -70 },
  { id: 'F02', name: '掘进工作面', x: 35, z: -70 },
]

const GAS_TYPES = [
  { code: 'T', name: '甲烷', unit: '%LEL', thresholds: [10, 25, 50, 75, 100], max: 100, color: '#34d399' },
  { code: 'CO', name: '一氧化碳', unit: 'ppm', thresholds: [12, 24, 37, 50, 100], max: 150, color: '#fbbf24' },
  { code: 'CO2', name: '二氧化碳', unit: '%', thresholds: [0.5, 1.0, 1.5, 2.0, 3.0], max: 5, color: '#60a5fa' },
  { code: 'O2', name: '氧气浓度', unit: '%', thresholds: [18, 19, 20.5, 21, 23], max: 25, color: '#a78bfa' },
]

function getLevel(code: string, v: number): number {
  const g = GAS_TYPES.find(x => x.code === code)!
  if (code === 'O2') {
    if (v < g.thresholds[0]) return 5
    if (v < g.thresholds[1]) return 4
    if (v < g.thresholds[2]) return 3
    if (v < g.thresholds[3]) return 2
    return 1
  }
  if (v >= g.thresholds[4]) return 5
  if (v >= g.thresholds[3]) return 4
  if (v >= g.thresholds[2]) return 3
  if (v >= g.thresholds[1]) return 2
  return 1
}

function fmt(v: number) { return Number.isInteger(v) ? v.toString() : v.toFixed(1) }
function levelColor(l: number) { return ['#22c55e', '#84cc16', '#eab308', '#f97316', '#ef4444'][Math.min(l - 1, 4)] }
function levelLabel(l: number) { return ['正常', '关注', '警戒', '警告', '危险'][Math.min(l - 1, 4)] }
function genId() { return Date.now().toString(36) + Math.random().toString(36).slice(2, 5) }

export interface Sensor {
  id: string; pointId: string; pointName: string
  gasCode: string; gasName: string; unit: string
  value: number; level: number; history: number[]; trend: 'up' | 'down' | 'stable'
}
export interface Alert {
  id: string; pointId: string; pointName: string
  gasCode: string; gasName: string
  value: number; unit: string; level: number; status: string; createdAt: string
}

function Gauge({ value, max, label, unit, color, level }: { value: number; max: number; label: string; unit: string; color: string; level: number }) {
  const pct = Math.min(value / max, 1)
  const cx = 50, cy = 50, r = 40
  const angle = -135 + pct * 270
  const rad = (angle * Math.PI) / 180
  const ex = cx + r * Math.cos(rad), ey = cy + r * Math.sin(rad)
  const large = pct > 0.375 ? 1 : 0
  const d = pct < 0.002 ? `M${cx},${cy - r} A${r},${r} 0 0,1 ${cx},${cy - r}` : `M${cx},${cy - r} A${r},${r} 0 ${large},1 ${ex},${ey}`
  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', padding: '8px 4px', background: '#f8fafc', border: `1px solid ${level >= 3 ? levelColor(level) + '44' : '#1e2535'}`, borderRadius: 8, minWidth: 82 }}>
      <svg width="100" height="66" viewBox="0 0 100 66" style={{ overflow: 'visible' }}>
        <path d={`M${cx},${cy - r} A${r},${r} 0 1,1 ${cx + 0.01},${cy - r}`} fill="none" stroke="#1e2535" strokeWidth="7" strokeLinecap="round" />
        {pct >= 0.002 && <path d={d} fill="none" stroke={color} strokeWidth="7" strokeLinecap="round" style={{ filter: `drop-shadow(0 0 4px ${color}80)` }} />}
        {[0, 0.25, 0.5, 0.75, 1.0].map((t, i) => {
          const a = (-135 + t * 270) * Math.PI / 180
          return <line key={i} x1={cx + (r - 5) * Math.cos(a)} y1={cy + (r - 5) * Math.sin(a)} x2={cx + (r + 1) * Math.cos(a)} y2={cy + (r + 1) * Math.sin(a)} stroke="#2a3447" strokeWidth="1.5" />
        })}
        <text x={cx} y={cy + 14} textAnchor="middle" fill={color} fontSize="13" fontWeight="bold" fontFamily="monospace">{fmt(value)}</text>
        <text x={cx} y={cy + 26} textAnchor="middle" fill="#4b5563" fontSize="7" fontFamily="monospace">{unit}</text>
      </svg>
      <div style={{ fontSize: 9, color: '#6b7280', textAlign: 'center', marginTop: 2 }}>{label}</div>
      <div style={{ fontSize: 9, color: levelColor(level), fontWeight: 700, marginTop: 1 }}>{levelLabel(level)}</div>
    </div>
  )
}

export default function MineViewer() {
  const mountRef = useRef<HTMLDivElement>(null)
  const rendererRef = useRef<THREE.WebGLRenderer | null>(null)
  const sceneRef = useRef<THREE.Scene | null>(null)
  const cameraRef = useRef<THREE.PerspectiveCamera | null>(null)
  const markerMeshes = useRef<Map<string, { sphere: THREE.Mesh; ring: THREE.Mesh; pulse: THREE.Mesh; beam: THREE.Mesh }>>(new Map())
  const animRef = useRef<number>(0)
  const sensorMeshes = useRef<Map<string, THREE.Mesh>>(new Map())

  const [sensors, setSensors] = useState<Sensor[]>([])
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [stats, setStats] = useState({ total: 12, normal: 12, warning: 0, danger: 0, offline: 0 })
  const [selectedPoint, setSelectedPoint] = useState<string | null>(null)
  const [demoMode] = useState(true)
  const [time, setTime] = useState(new Date())

  const initScene = useCallback(() => {
    const mount = mountRef.current!
    const W = mount.clientWidth, H = mount.clientHeight
    const scene = new THREE.Scene()
    scene.background = new THREE.Color(0xd8e8f0)
    scene.fog = new THREE.FogExp2(0xd8e8f0, 0.004)
    sceneRef.current = scene

    const camera = new THREE.PerspectiveCamera(55, W / H, 0.1, 1200)
    camera.position.set(0, 95, 130); camera.lookAt(0, 0, 0)
    cameraRef.current = camera

    const renderer = new THREE.WebGLRenderer({ antialias: true, powerPreference: 'high-performance' })
    renderer.setSize(W, H); renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
    renderer.shadowMap.enabled = true; renderer.shadowMap.type = THREE.PCFSoftShadowMap
    mount.appendChild(renderer.domElement); rendererRef.current = renderer

    scene.add(new THREE.AmbientLight(0x0d1a2e, 3))
    const sun = new THREE.DirectionalLight(0x4488ff, 0.8)
    sun.position.set(-50, 80, 30); sun.castShadow = true; scene.add(sun)
    const fillLight = new THREE.PointLight(0x2244aa, 1.5, 400); fillLight.position.set(0, 60, 0); scene.add(fillLight)

    const ground = new THREE.Mesh(new THREE.PlaneGeometry(350, 350, 40, 40), new THREE.MeshLambertMaterial({ color: 0x080e1a }))
    ground.rotation.x = -Math.PI / 2; ground.receiveShadow = true; scene.add(ground)
    scene.add(new THREE.GridHelper(350, 70, 0x0d2035, 0x0a1828))

    const wallEdgesMat = new THREE.LineBasicMaterial({ color: 0x1e3a5f })
    const addBox = (x: number, y: number, z: number, w: number, h: number, d: number) => {
      const mesh = new THREE.Mesh(new THREE.BoxGeometry(w, h, d), new THREE.MeshLambertMaterial({ color: 0x111827 }))
      mesh.position.set(x, y, z); mesh.castShadow = true; scene.add(mesh)
      const edges = new THREE.LineSegments(new THREE.EdgesGeometry(mesh.geometry), wallEdgesMat)
      edges.position.copy(mesh.position); scene.add(edges)
    }
    addBox(-30, 2, -20, 6, 4, 90); addBox(30, 2, -20, 6, 4, 90); addBox(0, 2, -60, 80, 4, 6); addBox(0, 2, 15, 80, 4, 6)
    addBox(-30, 2, -60, 6, 4, 30); addBox(30, 2, -60, 6, 4, 30); addBox(-45, 2, -5, 6, 4, 20); addBox(45, 2, -5, 6, 4, 20)

    const shaft = new THREE.Mesh(new THREE.CylinderGeometry(9, 9, 35, 20), new THREE.MeshLambertMaterial({ color: 0x1a2740 }))
    shaft.position.set(0, 17, 0); shaft.castShadow = true; scene.add(shaft)
    const ringGeo = new THREE.TorusGeometry(9, 0.8, 8, 24)
    const ringMat = new THREE.MeshLambertMaterial({ color: 0x2563eb, emissive: 0x1d4ed8, emissiveIntensity: 0.5 })
    ;[-1, 12, 25].forEach(y => { const r = new THREE.Mesh(ringGeo, ringMat); r.position.set(0, y, 0); scene.add(r) })
    const frameMat = new THREE.MeshLambertMaterial({ color: 0x1e3a5f })
    ;[[-12, 0, -12], [12, 0, -12], [-12, 0, 12], [12, 0, 12]].forEach(([x, , z]) => {
      const m = new THREE.Mesh(new THREE.CylinderGeometry(0.8, 1.0, 38, 6), frameMat); m.position.set(x, 19, z); scene.add(m)
    })
    const beamTop = new THREE.Mesh(new THREE.BoxGeometry(26, 1.8, 26), frameMat); beamTop.position.y = 38; scene.add(beamTop)

    const fanMat = new THREE.MeshLambertMaterial({ color: 0x1e4d7a, emissive: 0x1e4d7a, emissiveIntensity: 0.2 })
    ;[[-30, 5, -58], [30, 5, -58], [0, 5, 17]].forEach(([x, y, z]) => {
      const fan = new THREE.Mesh(new THREE.CylinderGeometry(2.5, 2.5, 1.5, 16), fanMat)
      fan.position.set(x, y, z); scene.add(fan)
    })

    const nodeGeo = new THREE.OctahedronGeometry(1.5, 0)
    MONITOR_POINTS.forEach(pt => {
      const mat = new THREE.MeshLambertMaterial({ color: 0x22c55e, emissive: 0x22c55e, emissiveIntensity: 0.4 })
      const mesh = new THREE.Mesh(nodeGeo, mat); mesh.position.set(pt.x, 4, pt.z); mesh.castShadow = true
      scene.add(mesh); sensorMeshes.current.set(pt.id, mesh)
      const canvas = document.createElement('canvas'); canvas.width = 160; canvas.height = 36
      const ctx = canvas.getContext('2d')!; ctx.fillStyle = '#4f8eff'; ctx.font = 'bold 20px monospace'; ctx.fillText(pt.id, 0, 24)
      const sprite = new THREE.Sprite(new THREE.SpriteMaterial({ map: new THREE.CanvasTexture(canvas) }))
      sprite.position.set(pt.x, 7.5, pt.z); sprite.scale.set(8, 1.8, 1); scene.add(sprite)
    })

    let dragging = false, prev = { x: 0, y: 0 }, theta = 0.3, phi = 0.95, radius = 160
    const updateCamera = () => {
      camera.position.x = radius * Math.sin(phi) * Math.cos(theta)
      camera.position.y = radius * Math.cos(phi)
      camera.position.z = radius * Math.sin(phi) * Math.sin(theta)
      camera.lookAt(0, 0, 0)
    }
    updateCamera()
    renderer.domElement.addEventListener('mousedown', e => { dragging = true; prev = { x: e.clientX, y: e.clientY } })
    window.addEventListener('mousemove', e => {
      if (!dragging) return; theta -= (e.clientX - prev.x) * 0.004; phi -= (e.clientY - prev.y) * 0.004
      phi = Math.max(0.2, Math.min(Math.PI / 2 - 0.05, phi)); prev = { x: e.clientX, y: e.clientY }; updateCamera()
    })
    window.addEventListener('mouseup', () => { dragging = false })
    renderer.domElement.addEventListener('wheel', e => { radius = Math.max(50, Math.min(300, radius + e.deltaY * 0.15)); updateCamera() })

    let t = 0
    function animate() {
      animRef.current = requestAnimationFrame(animate); t += 0.015
      sensorMeshes.current.forEach((m, id) => { m.rotation.y += 0.02; m.position.y = 4 + Math.sin(t * 1.5 + id.charCodeAt(0) * 0.3) * 0.4 })
      markerMeshes.current.forEach(({ sphere, pulse, beam }) => {
        sphere.scale.setScalar(1 + 0.12 * Math.sin(t * 4)); sphere.rotation.y += 0.03
        if (pulse) { pulse.scale.setScalar(1 + 0.2 * Math.sin(t * 3)); (pulse.material as THREE.MeshBasicMaterial).opacity = 0.3 * (1 - (t % (Math.PI * 2 / 3)) / (Math.PI * 2 / 3)) }
        if (beam) beam.scale.y = 1 + 0.08 * Math.sin(t * 2.5)
      })
      renderer.render(scene, camera)
    }
    animate()
    const onResize = () => { if (!mount) return; camera.aspect = mount.clientWidth / mount.clientHeight; camera.updateProjectionMatrix(); renderer.setSize(mount.clientWidth, mount.clientHeight) }
    window.addEventListener('resize', onResize)
    return () => { cancelAnimationFrame(animRef.current); renderer.dispose(); window.removeEventListener('resize', onResize) }
  }, [])

  useEffect(() => { const c = initScene(); return c }, [initScene])

  useEffect(() => {
    if (!demoMode) return
    const init: Sensor[] = MONITOR_POINTS.map(pt => {
      const gas = GAS_TYPES[Math.floor(Math.random() * GAS_TYPES.length)]
      const base = gas.thresholds[0] * (0.2 + Math.random() * 0.5)
      const val = Number(base.toFixed(gas.code === 'CO' || gas.code === 'O2' ? 1 : 2))
      const hist = Array.from({ length: 20 }, () => val * (0.8 + Math.random() * 0.4))
      return { id: pt.id + '_' + gas.code, pointId: pt.id, pointName: pt.name, gasCode: gas.code, gasName: gas.name, unit: gas.unit, value: val, level: getLevel(gas.code, val), history: hist, trend: ('stable') }
    })
    setSensors(init)
    const iv = setInterval(() => {
      setTime(new Date())
      setSensors(prev => {
        const next = prev.map(s => {
          const gas = GAS_TYPES.find(g => g.code === s.gasCode) || GAS_TYPES[0]
          const delta = (Math.random() - 0.47) * (gas.max * 0.07)
          let v = Math.max(0, Number((s.value + delta).toFixed(gas.code === 'CO' || gas.code === 'O2' ? 1 : 2)))
          v = Math.min(v, gas.max)
          const hist = [...s.history.slice(-19), v]
          const trend: 'up' | 'down' | 'stable' = v > s.history[s.history.length - 1] ? 'up' : v < s.history[s.history.length - 1] ? 'down' : 'stable'
          return { ...s, value: v, level: getLevel(s.gasCode, v), history: hist, trend }
        })
        setStats({ total: next.length, normal: next.filter(s => s.level <= 2).length, warning: next.filter(s => s.level === 3).length, danger: next.filter(s => s.level >= 4).length, offline: 0 })
        if (Math.random() < 0.18) {
          const s = next[Math.floor(Math.random() * next.length)]
          if (s.level >= 3) {
            const a: Alert = { id: genId(), pointId: s.pointId, pointName: s.pointName, gasCode: s.gasCode, gasName: s.gasName, value: s.value, unit: s.unit, level: s.level, status: 'open', createdAt: new Date().toISOString() }
            setAlerts(prev => [a, ...prev].slice(0, 50))
            add3DMarker(a)
          }
        }
        return next
      })
    }, 2500)
    return () => clearInterval(iv)
  }, [demoMode])

  const add3DMarker = useCallback((alert: Alert) => {
    const scene = sceneRef.current
    if (!scene) return
    const old = markerMeshes.current.get(alert.pointId)
    if (old) { scene.remove(old.sphere); scene.remove(old.ring); scene.remove(old.pulse); scene.remove(old.beam) }
    markerMeshes.current.delete(alert.pointId)
    const pt = MONITOR_POINTS.find(p => p.id === alert.pointId)
    if (!pt) return
    const color = levelColor(alert.level)
    const mat = new THREE.MeshLambertMaterial({ color, emissive: color, emissiveIntensity: 0.9, transparent: true, opacity: 0.85 })
    const sphere = new THREE.Mesh(new THREE.SphereGeometry(alert.level * 1.2 + 2, 20, 20), mat)
    sphere.position.set(pt.x, 4.5, pt.z); scene.add(sphere)
    const ring = new THREE.Mesh(new THREE.TorusGeometry(alert.level * 2 + 4, 0.4, 8, 32), new THREE.MeshBasicMaterial({ color, transparent: true, opacity: 0.6 }))
    ring.rotation.x = -Math.PI / 2; ring.position.set(pt.x, 0.3, pt.z); scene.add(ring)
    const pulse = new THREE.Mesh(new THREE.SphereGeometry(alert.level * 1.5 + 3, 16, 16), new THREE.MeshBasicMaterial({ color, transparent: true, opacity: 0.3, wireframe: true }))
    pulse.position.set(pt.x, 4.5, pt.z); scene.add(pulse)
    const beam = new THREE.Mesh(new THREE.CylinderGeometry(0.3, 0.3, 50, 10), new THREE.MeshLambertMaterial({ color, transparent: true, opacity: 0.2 }))
    beam.position.set(pt.x, 27, pt.z); scene.add(beam)
    markerMeshes.current.set(alert.pointId, { sphere, ring, pulse, beam })
  }, [])

  const triggerAlert = useCallback(() => {
    const pt = MONITOR_POINTS[Math.floor(Math.random() * MONITOR_POINTS.length)]
    const gas = GAS_TYPES[Math.floor(Math.random() * GAS_TYPES.length)]
    const v = Number((gas.thresholds[3] * (0.85 + Math.random() * 0.3)).toFixed(gas.code === 'CO' || gas.code === 'O2' ? 1 : 2))
    const level = getLevel(gas.code, v)
    const a: Alert = { id: genId(), pointId: pt.id, pointName: pt.name, gasCode: gas.code, gasName: gas.name, value: v, unit: gas.unit, level, status: 'open', createdAt: new Date().toISOString() }
    setAlerts(prev => [a, ...prev].slice(0, 50)); add3DMarker(a)
    setStats(s => ({ ...s, danger: s.danger + (level >= 4 ? 1 : 0), warning: s.warning + (level === 3 ? 1 : 0) }))
  }, [add3DMarker])

  return (
    <div style={{ position: 'relative', width: '100%', height: '100vh', background: '#f0f4f8', overflow: 'hidden', fontFamily: '"PingFang SC","Microsoft YaHei",sans-serif' }}>
      <div ref={mountRef} style={{ width: '100%', height: '100%' }} />

      {/* TOP BAR */}
      <div style={{ position: 'absolute', top: 0, left: 0, right: 0, height: 52, background: 'rgba(5,8,16,0.95)', borderBottom: '1px solid rgba(59,130,246,0.25)', display: 'flex', alignItems: 'center', padding: '0 16px', gap: 20, zIndex: 20, backdropFilter: 'blur(12px)' }}>
        <div style={{ fontSize: 15, fontWeight: 700, color: '#3b82f6', letterSpacing: 1 }}>🕳️ 智慧矿山综合监控平台</div>
        <div style={{ fontSize: 11, color: '#374151' }}>|</div>
        <div style={{ fontSize: 11, color: '#6b7280' }}>{time.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })} {time.toLocaleTimeString('zh-CN')}</div>
        <div style={{ marginLeft: 'auto', display: 'flex', gap: 24, fontSize: 12 }}>
          {[{ label: '监测点', value: stats.total, color: '#60a5fa' }, { label: '正常', value: stats.normal, color: '#22c55e' }, { label: '预警', value: stats.warning, color: '#eab308' }, { label: '危险', value: stats.danger, color: '#ef4444' }, { label: '离线', value: stats.offline, color: '#6b7280' }].map((s, i) => (
            <div key={i} style={{ textAlign: 'center' }}>
              <div style={{ fontSize: 18, fontWeight: 700, color: s.color, lineHeight: 1.2 }}>{s.value}</div>
              <div style={{ fontSize: 10, color: '#4b5563' }}>{s.label}</div>
            </div>
          ))}
        </div>
      </div>

      {/* LEFT: Gas gauges */}
      <div style={{ position: 'absolute', top: 62, left: 12, width: 410, background: 'rgba(5,8,16,0.92)', border: '1px solid rgba(59,130,246,0.2)', borderRadius: 12, backdropFilter: 'blur(12px)', padding: '10px 12px', zIndex: 10 }}>
        <div style={{ fontSize: 12, fontWeight: 600, color: '#3b82f6', marginBottom: 8, paddingBottom: 6, borderBottom: '1px solid rgba(59,130,246,0.15)' }}>📊 气体监测总览</div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 6 }}>
          {GAS_TYPES.map(g => {
            const rel = sensors.filter(s => s.gasCode === g.code)
            const avg = rel.length ? rel.reduce((a, b) => a + b.value, 0) / rel.length : 0
            const maxL = rel.length ? Math.max(...rel.map(s => s.level)) : 1
            return <Gauge key={g.code} value={avg} max={g.max} label={g.name} unit={g.unit} color={g.color} level={maxL} />
          })}
        </div>
      </div>

      {/* LEFT BOTTOM: Sensor table */}
      <div style={{ position: 'absolute', bottom: 12, left: 12, width: 410, background: 'rgba(5,8,16,0.92)', border: '1px solid rgba(59,130,246,0.2)', borderRadius: 12, backdropFilter: 'blur(12px)', padding: '10px 12px', zIndex: 10, maxHeight: 280, overflowY: 'auto' }}>
        <div style={{ fontSize: 12, fontWeight: 600, color: '#3b82f6', marginBottom: 8, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span>📡 {sensors.length} 个监测点</span>
          <button onClick={triggerAlert} style={{ fontSize: 10, padding: '2px 8px', background: 'rgba(239,68,68,0.15)', border: '1px solid rgba(239,68,68,0.4)', borderRadius: 4, color: '#ef4444', cursor: 'pointer' }}>🚨 模拟报警</button>
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 4 }}>
          {sensors.map(s => (
            <div key={s.id} onClick={() => setSelectedPoint(selectedPoint === s.pointId ? null : s.pointId)} style={{
              padding: '5px 7px', borderRadius: 6, cursor: 'pointer',
              background: selectedPoint === s.pointId ? 'rgba(59,130,246,0.12)' : 'rgba(255,255,255,0.02)',
              border: `1px solid ${selectedPoint === s.pointId ? 'rgba(59,130,246,0.5)' : s.level >= 4 ? levelColor(s.level) + '44' : 'rgba(255,255,255,0.04)'}`,
            }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 2 }}>
                <span style={{ fontSize: 10, fontWeight: 700, color: '#4f8eff' }}>{s.pointId}</span>
                <span style={{ fontSize: 9, color: levelColor(s.level), fontWeight: 700 }}>{levelLabel(s.level)}</span>
              </div>
              <div style={{ fontSize: 9, color: '#6b7280', marginBottom: 2 }}>{s.gasName}</div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span style={{ fontSize: 13, fontWeight: 700, color: levelColor(s.level), fontFamily: 'monospace' }}>{fmt(s.value)}<span style={{ fontSize: 8, marginLeft: 1 }}>{s.unit}</span></span>
                <span style={{ fontSize: 9 }}>{s.trend === 'up' ? '↑' : s.trend === 'down' ? '↓' : '→'}</span>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* RIGHT TOP: Alert panel */}
      <div style={{ position: 'absolute', top: 62, right: 12, width: 310, background: 'rgba(5,8,16,0.92)', border: '1px solid rgba(249,115,22,0.2)', borderRadius: 12, backdropFilter: 'blur(12px)', padding: '10px 12px', zIndex: 10, maxHeight: 300, overflowY: 'auto' }}>
        <div style={{ fontSize: 12, fontWeight: 600, color: '#f97316', marginBottom: 8, paddingBottom: 6, borderBottom: '1px solid rgba(249,115,22,0.15)' }}>
          🚨 报警记录 ({alerts.length})
        </div>
        {alerts.length === 0 && <div style={{ color: '#374151', fontSize: 12, textAlign: 'center', padding: '16px 0' }}>暂无报警</div>}
        {alerts.slice(0, 30).map(a => (
          <div key={a.id} style={{ padding: '6px 0', borderBottom: '1px solid rgba(255,255,255,0.04)', borderLeft: `3px solid ${levelColor(a.level)}`, paddingLeft: 8, marginBottom: 4 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11 }}>
              <span style={{ color: '#9ca3af' }}>{a.pointName}</span>
              <span style={{ color: levelColor(a.level), fontWeight: 700 }}>{a.gasName} {fmt(a.value)}{a.unit}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 10, color: '#4b5563', marginTop: 2 }}>
              <span>{levelLabel(a.level)}</span>
              <span>{new Date(a.createdAt).toLocaleTimeString()}</span>
            </div>
          </div>
        ))}
      </div>

      {/* BOTTOM RIGHT: Legend */}
      <div style={{ position: 'absolute', bottom: 12, right: 12, width: 180, background: 'rgba(5,8,16,0.88)', border: '1px solid rgba(255,255,255,0.07)', borderRadius: 8, backdropFilter: 'blur(8px)', padding: '10px 14px', fontSize: 11, color: '#6b7280', zIndex: 10 }}>
        <div style={{ marginBottom: 6, fontWeight: 600, color: '#4b5563' }}>图例</div>
        {[['#22c55e', '正常'], ['#84cc16', '关注'], ['#eab308', '警戒'], ['#f97316', '警告'], ['#ef4444', '危险']].map(([c, l]) => (
          <div key={l} style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 3 }}>
            <div style={{ width: 8, height: 8, borderRadius: '50%', background: c, boxShadow: `0 0 5px ${c}` }} />
            {l}
          </div>
        ))}
        <div style={{ marginTop: 8, color: '#374151', fontSize: 10 }}>🖱️ 拖拽旋转 · 滚轮缩放</div>
      </div>

    </div>
  )
}
