import { useEffect, useRef, useState, useCallback } from 'react'
import * as THREE from 'three'

// 模拟矿井监测点位置
const MONITOR_POINTS = [
  { id: 'A01', name: '主井东巷', x: -40, z: -30 },
  { id: 'A02', name: '主井西巷', x: 40, z: -30 },
  { id: 'B01', name: '副井底车场', x: -20, z: 10 },
  { id: 'B02', name: '回风巷道', x: 20, z: 10 },
  { id: 'C01', name: '运输大巷', x: -35, z: -50 },
  { id: 'C02', name: '采煤工作面', x: 35, z: -50 },
  { id: 'D01', name: '中央变电所', x: 0, z: -20 },
  { id: 'D02', name: '水泵房', x: 0, z: 20 },
  { id: 'E01', name: '避难硐室', x: -50, z: 0 },
  { id: 'E02', name: '炸药库', x: 50, z: 0 },
]

const GAS_TYPES = [
  { code: 'T', name: '甲烷', unit: '%', thresholds: [0.5, 0.8, 1.0, 1.5, 2.0] },
  { code: 'CO', name: '一氧化碳', unit: 'ppm', thresholds: [12, 24, 37, 50, 100] },
  { code: 'CO2', name: '二氧化碳', unit: '%', thresholds: [0.5, 1.0, 1.5, 2.0, 3.0] },
  { code: 'O2', name: '氧气', unit: '%', thresholds: [18, 19, 20.5, 21, 23] },
]

function getLevel(gasCode: string, value: number): number {
  const gas = GAS_TYPES.find(g => g.code === gasCode)
  if (!gas) return 1
  if (gas.code === 'O2') {
    if (value < gas.thresholds[0]) return 5
    if (value < gas.thresholds[1]) return 4
    if (value < gas.thresholds[2]) return 3
    if (value < gas.thresholds[3]) return 2
    return 1
  }
  if (value >= gas.thresholds[4]) return 5
  if (value >= gas.thresholds[3]) return 4
  if (value >= gas.thresholds[2]) return 3
  if (value >= gas.thresholds[1]) return 2
  return 1
}

function getGasColor(level: number): number {
  const colors = [0x22c55e, 0x84cc16, 0xeab308, 0xf97316, 0xef4444]
  return colors[Math.min(level - 1, 4)]
}

export interface SensorReading {
  id: string
  location: string
  locationName: string
  gasType: string
  gasName: string
  value: number
  unit: string
  level: number
  time: string
}

export interface Alert {
  id: string
  location: string
  locationName: string
  gasType: string
  gasName: string
  value: number
  unit: string
  level: number
  status: string
  createdAt: string
}

export interface DashboardStats {
  totalSensors: number
  onlineSensors: number
  normalCount: number
  warningCount: number
  dangerCount: number
  todayAlerts: number
}

function genId() {
  return Date.now().toString(36) + Math.random().toString(36).slice(2, 6)
}

function fmt(v: number) {
  return Number.isInteger(v) ? v.toString() : v.toFixed(2)
}

export default function MineViewer() {
  const mountRef = useRef<HTMLDivElement>(null)
  const rendererRef = useRef<THREE.WebGLRenderer | null>(null)
  const sceneRef = useRef<THREE.Scene | null>(null)
  const cameraRef = useRef<THREE.PerspectiveCamera | null>(null)
  const markersRef = useRef<{ mesh: THREE.Mesh; beam: THREE.Mesh; alertId: string }[]>([])
  const animationRef = useRef<number>(0)
  const sensorMeshesRef = useRef<Map<string, THREE.Mesh>>(new Map())

  const [alerts, setAlerts] = useState<Alert[]>([])
  const [sensors, setSensors] = useState<SensorReading[]>([])
  const [stats, setStats] = useState<DashboardStats>({ totalSensors: 10, onlineSensors: 10, normalCount: 10, warningCount: 0, dangerCount: 0, todayAlerts: 0 })
  const [demoMode, setDemoMode] = useState(true)

  // --- 3D Scene ---
  const initScene = useCallback(() => {
    const mount = mountRef.current!
    const w = mount.clientWidth
    const h = mount.clientHeight

    const scene = new THREE.Scene()
    scene.background = new THREE.Color(0x050810)
    scene.fog = new THREE.FogExp2(0x050810, 0.006)
    sceneRef.current = scene

    const camera = new THREE.PerspectiveCamera(60, w / h, 0.1, 1000)
    camera.position.set(0, 80, 120)
    camera.lookAt(0, 0, 0)
    cameraRef.current = camera

    const renderer = new THREE.WebGLRenderer({ antialias: true, powerPreference: 'high-performance' })
    renderer.setSize(w, h)
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
    renderer.shadowMap.enabled = true
    mount.appendChild(renderer.domElement)
    rendererRef.current = renderer

    // Lights
    scene.add(new THREE.AmbientLight(0x112244, 2))
    const sun = new THREE.DirectionalLight(0xffffff, 1.5)
    sun.position.set(50, 100, 50)
    sun.castShadow = true
    scene.add(sun)
    const blueLight = new THREE.PointLight(0x4488ff, 1.5, 300)
    blueLight.position.set(-30, 40, 20)
    scene.add(blueLight)
    const orangeLight = new THREE.PointLight(0xff6600, 0.8, 200)
    orangeLight.position.set(30, 30, -30)
    scene.add(orangeLight)

    // Ground
    const groundGeo = new THREE.PlaneGeometry(300, 300, 30, 30)
    const groundMat = new THREE.MeshLambertMaterial({ color: 0x0d1117 })
    const ground = new THREE.Mesh(groundGeo, groundMat)
    ground.rotation.x = -Math.PI / 2
    ground.receiveShadow = true
    scene.add(ground)

    // Grid
    const grid = new THREE.GridHelper(300, 60, 0x1a2332, 0x111827)
    grid.position.y = 0.05
    scene.add(grid)

    // Tunnel walls (矿井巷道)
    const wallMat = new THREE.MeshLambertMaterial({ color: 0x1c2333 })
    const tunnels: [number, number, number, number, number, number][] = [
      // [x, y, z, w, h, d] — main corridors
      [-30, 2, -20, 6, 4, 80],  // West corridor
      [30, 2, -20, 6, 4, 80],  // East corridor
      [0, 2, -55, 80, 4, 6],   // South corridor
      [0, 2, 15, 80, 4, 6],     // North corridor
      [-30, 2, -55, 6, 4, 30],  // SW connector
      [30, 2, -55, 6, 4, 30],   // SE connector
    ]
    tunnels.forEach(([x, y, z, w, h, d]) => {
      const geo = new THREE.BoxGeometry(w, h, d)
      const mesh = new THREE.Mesh(geo, wallMat)
      mesh.position.set(x, y, z)
      mesh.castShadow = true
      scene.add(mesh)
    })

    // Mine shaft (主井)
    const shaftGeo = new THREE.CylinderGeometry(8, 8, 30, 16)
    const shaftMat = new THREE.MeshLambertMaterial({ color: 0x2a3550, wireframe: false })
    const shaft = new THREE.Mesh(shaftGeo, shaftMat)
    shaft.position.set(0, 15, 0)
    shaft.castShadow = true
    scene.add(shaft)

    // Shaft headframe (井架)
    const frameMat = new THREE.MeshLambertMaterial({ color: 0x334466 })
    const posts = [[-10,15,-10],[10,15,-10],[-10,15,10],[10,15,10]] as [number,number,number][]
    posts.forEach(([x,y,z]) => {
      const g = new THREE.CylinderGeometry(0.5, 0.5, 30, 6)
      const m = new THREE.Mesh(g, frameMat)
      m.position.set(x, y, z)
      scene.add(m)
    })
    const beamGeo = new THREE.BoxGeometry(22, 1.5, 22)
    const beam = new THREE.Mesh(beamGeo, frameMat)
    beam.position.set(0, 30, 0)
    scene.add(beam)

    // Ventilation/fans (通风设施)
    const fanMat = new THREE.MeshLambertMaterial({ color: 0x4488aa })
    const fanPositions: [number, number, number][] = [[-30, 4, -55], [30, 4, -55], [0, 4, 15]]
    fanPositions.forEach(([x,y,z]) => {
      const g = new THREE.CylinderGeometry(2, 2, 1, 12)
      const m = new THREE.Mesh(g, fanMat)
      m.position.set(x, y, z)
      scene.add(m)
    })

    // Sensor monitoring points
    const sensorGeo = new THREE.OctahedronGeometry(1.2, 0)
    MONITOR_POINTS.forEach(pt => {
      const mat = new THREE.MeshLambertMaterial({ color: 0x22c55e, emissive: 0x22c55e, emissiveIntensity: 0.3 })
      const mesh = new THREE.Mesh(sensorGeo, mat)
      mesh.position.set(pt.x, 3, pt.z)
      mesh.castShadow = true
      scene.add(mesh)
      sensorMeshesRef.current.set(pt.id, mesh)

      // Label sprite
      const canvas = document.createElement('canvas')
      canvas.width = 128; canvas.height = 32
      const ctx = canvas.getContext('2d')!
      ctx.fillStyle = '#4f8eff'
      ctx.font = 'bold 18px monospace'
      ctx.fillText(pt.id, 0, 22)
      const texture = new THREE.CanvasTexture(canvas)
      const spriteMat = new THREE.SpriteMaterial({ map: texture })
      const sprite = new THREE.Sprite(spriteMat)
      sprite.position.set(pt.x, 6, pt.z)
      sprite.scale.set(6, 1.5, 1)
      scene.add(sprite)
    })

    // Camera controls
    let isDragging = false
    let prev = { x: 0, y: 0 }
    let theta = 0.3, phi = 0.9, radius = 140
    function updateCamera() {
      camera.position.x = radius * Math.sin(phi) * Math.cos(theta)
      camera.position.y = radius * Math.cos(phi)
      camera.position.z = radius * Math.sin(phi) * Math.sin(theta)
      camera.lookAt(0, 0, 0)
    }
    updateCamera()
    renderer.domElement.addEventListener('mousedown', e => { isDragging = true; prev = { x: e.clientX, y: e.clientY } })
    window.addEventListener('mousemove', e => {
      if (!isDragging) return
      theta -= (e.clientX - prev.x) * 0.004
      phi -= (e.clientY - prev.y) * 0.004
      phi = Math.max(0.2, Math.min(Math.PI / 2 - 0.1, phi))
      prev = { x: e.clientX, y: e.clientY }
      updateCamera()
    })
    window.addEventListener('mouseup', () => { isDragging = false })
    renderer.domElement.addEventListener('wheel', e => {
      radius = Math.max(40, Math.min(250, radius + e.deltaY * 0.12))
      updateCamera()
    })

    const onResize = () => {
      if (!mount) return
      camera.aspect = mount.clientWidth / mount.clientHeight
      camera.updateProjectionMatrix()
      renderer.setSize(mount.clientWidth, mount.clientHeight)
    }
    window.addEventListener('resize', onResize)

    let time = 0
    function animate() {
      animationRef.current = requestAnimationFrame(animate)
      time += 0.01

      // Animate sensor nodes
      sensorMeshesRef.current.forEach((mesh, id) => {
        mesh.rotation.y += 0.02
        mesh.position.y = 3 + Math.sin(time * 2 + id.charCodeAt(0)) * 0.3
      })

      // Animate markers
      markersRef.current.forEach(({ mesh, beam: b }) => {
        mesh.scale.setScalar(1 + 0.08 * Math.sin(time * 4))
        if (b) b.scale.y = 1 + 0.1 * Math.sin(time * 3)
      })

      renderer.render(scene, camera)
    }
    animate()

    return () => {
      cancelAnimationFrame(animationRef.current)
      renderer.dispose()
      window.removeEventListener('resize', onResize)
    }
  }, [])

  useEffect(() => {
    const cleanup = initScene()
    return cleanup
  }, [initScene])

  // --- Demo simulation ---
  useEffect(() => {
    if (!demoMode) return

    // Generate initial sensors
    const initSensors = MONITOR_POINTS.map(pt => {
      const gas = GAS_TYPES[Math.floor(Math.random() * GAS_TYPES.length)]
      const base = gas.thresholds[0] * (0.3 + Math.random() * 0.5)
      const value = Number(base.toFixed(gas.code === 'T' || gas.code === 'CO2' ? 2 : 1))
      const level = getLevel(gas.code, value)
      return {
        id: pt.id, location: pt.id, locationName: pt.name,
        gasType: gas.code, gasName: gas.name,
        value, unit: gas.unit, level,
        time: new Date().toLocaleTimeString(),
      }
    })
    setSensors(initSensors)
    updateStats(initSensors, [])

    // Interval: update sensor values every 3s
    const simInterval = setInterval(() => {
      setSensors(prev => {
        const next = prev.map(s => {
          const gas = GAS_TYPES.find(g => g.code === s.gasType) || GAS_TYPES[0]
          const delta = (Math.random() - 0.48) * (gas.thresholds[4] * 0.15)
          let value = Math.max(0, Number((s.value + delta).toFixed(2)))
          // O2 special case
          if (s.gasType === 'O2') {
            value = Math.min(23, Math.max(16, 20.9 + (Math.random() - 0.5) * 1.5))
            value = Number(value.toFixed(1))
          }
          const level = getLevel(s.gasType, value)
          return { ...s, value, level, time: new Date().toLocaleTimeString() }
        })
        updateStats(next, [])
        return next
      })

      // Random alert every 8-15s
      if (Math.random() < 0.25) {
        const pt = MONITOR_POINTS[Math.floor(Math.random() * MONITOR_POINTS.length)]
        const gas = GAS_TYPES[Math.floor(Math.random() * GAS_TYPES.length)]
        const value = Number((gas.thresholds[2] * (0.8 + Math.random() * 0.4)).toFixed(gas.code === 'T' || gas.code === 'CO2' ? 2 : 1))
        const level = getLevel(gas.code, value)
        const alert: Alert = {
          id: genId(), location: pt.id, locationName: pt.name,
          gasType: gas.code, gasName: gas.name,
          value, unit: gas.unit, level, status: 'open',
          createdAt: new Date().toISOString(),
        }
        setAlerts(prev => [alert, ...prev].slice(0, 30))
        setStats(s => ({ ...s, todayAlerts: s.todayAlerts + 1, dangerCount: level >= 4 ? s.dangerCount + 1 : s.dangerCount }))
        add3DAlert(alert)
      }
    }, 3000)

    return () => clearInterval(simInterval)
  }, [demoMode])

  const add3DAlert = useCallback((alert: Alert) => {
    const scene = sceneRef.current
    if (!scene) return

    // Remove old marker for same location
    markersRef.current
      .filter(m => m.alertId === alert.location)
      .forEach(({ mesh, beam }) => { scene.remove(mesh); scene.remove(beam) })
    markersRef.current = markersRef.current.filter(m => m.alertId !== alert.location)

    const pt = MONITOR_POINTS.find(p => p.id === alert.location)
    if (!pt) return
    const x = pt.x, z = pt.z
    const color = getGasColor(alert.level)

    const sphereGeo = new THREE.SphereGeometry(alert.level * 1.2 + 1.5, 20, 20)
    const sphereMat = new THREE.MeshLambertMaterial({
      color, emissive: color, emissiveIntensity: 0.8,
      transparent: true, opacity: 0.85,
    })
    const sphere = new THREE.Mesh(sphereGeo, sphereMat)
    sphere.position.set(x, 4, z)
    scene.add(sphere)

    const beamGeo = new THREE.CylinderGeometry(0.3, 0.3, 40, 10)
    const beamMat = new THREE.MeshLambertMaterial({ color, transparent: true, opacity: 0.25 })
    const beam = new THREE.Mesh(beamGeo, beamMat)
    beam.position.set(x, 22, z)
    scene.add(beam)

    // Ring pulse
    const ringGeo = new THREE.RingGeometry(2, 2.5, 32)
    const ringMat = new THREE.MeshBasicMaterial({ color, transparent: true, opacity: 0.5, side: THREE.DoubleSide })
    const ring = new THREE.Mesh(ringGeo, ringMat)
    ring.rotation.x = -Math.PI / 2
    ring.position.set(x, 0.2, z)
    scene.add(ring)

    markersRef.current.push({ mesh: sphere, beam, alertId: alert.location })
    markersRef.current.push({ mesh: ring, beam: beam, alertId: alert.location + '_ring' })
  }, [])

  const updateStats = (sensorList: SensorReading[], _alertList: Alert[]) => {
    setStats(s => ({
      ...s,
      normalCount: sensorList.filter(s => s.level <= 2).length,
      warningCount: sensorList.filter(s => s.level === 3).length,
      dangerCount: sensorList.filter(s => s.level >= 4).length,
    }))
  }

  const addManualAlert = () => {
    const pt = MONITOR_POINTS[Math.floor(Math.random() * MONITOR_POINTS.length)]
    const gas = GAS_TYPES[Math.floor(Math.random() * GAS_TYPES.length)]
    const value = Number((gas.thresholds[3] * (0.9 + Math.random() * 0.3)).toFixed(2))
    const level = getLevel(gas.code, value)
    const alert: Alert = {
      id: genId(), location: pt.id, locationName: pt.name,
      gasType: gas.code, gasName: gas.name,
      value, unit: gas.unit, level, status: 'open',
      createdAt: new Date().toISOString(),
    }
    setAlerts(prev => [alert, ...prev].slice(0, 30))
    setStats(s => ({ ...s, todayAlerts: s.todayAlerts + 1 }))
    add3DAlert(alert)
  }

  const levelLabel = (l: number) => ['正常', '注意', '警戒', '警告', '危险'][(l - 1) || 0]
  const levelColor = (l: number) => ['#22c55e', '#84cc16', '#eab308', '#f97316', '#ef4444'][Math.min(l - 1, 4)]

  return (
    <div style={{ position: 'relative', width: '100%', height: 'calc(100vh - 52px)', background: '#050810', overflow: 'hidden' }}>

      {/* ─── 3D Canvas ─── */}
      <div ref={mountRef} style={{ width: '100%', height: '100%' }} />

      {/* ─── Top Stats Bar ─── */}
      <div style={{
        position: 'absolute', top: 0, left: 0, right: 0,
        display: 'flex', gap: 0, background: 'rgba(5,8,16,0.92)',
        borderBottom: '1px solid rgba(79,142,255,0.2)',
        backdropFilter: 'blur(12px)',
        zIndex: 10,
      }}>
        {[
          { label: '在线监测点', value: stats.onlineSensors, unit: `/${stats.totalSensors}`, color: '#4f8eff' },
          { label: '正常', value: stats.normalCount, color: '#22c55e' },
          { label: '注意', value: stats.warningCount, color: '#eab308' },
          { label: '危险', value: stats.dangerCount, color: '#ef4444' },
          { label: '今日报警', value: stats.todayAlerts, color: '#f97316' },
        ].map((s, i) => (
          <div key={i} style={{
            flex: 1, textAlign: 'center', padding: '10px 0',
            borderRight: i < 4 ? '1px solid rgba(255,255,255,0.05)' : 'none',
          }}>
            <div style={{ fontSize: 22, fontWeight: 700, color: s.color, lineHeight: 1.2 }}>{s.value}<span style={{ fontSize: 13, fontWeight: 400 }}>{s.unit || ''}</span></div>
            <div style={{ fontSize: 11, color: '#6b7280', marginTop: 2 }}>{s.label}</div>
          </div>
        ))}
      </div>

      {/* ─── Right Panel: Real-time Sensors ─── */}
      <div style={{
        position: 'absolute', top: 54, right: 12, width: 260,
        background: 'rgba(5,8,16,0.9)', border: '1px solid rgba(79,142,255,0.2)',
        borderRadius: 10, backdropFilter: 'blur(12px)',
        overflow: 'hidden', zIndex: 10,
      }}>
        <div style={{ padding: '10px 14px', borderBottom: '1px solid rgba(255,255,255,0.06)', fontSize: 12, fontWeight: 600, color: '#4f8eff', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span>📡 实时监测数据</span>
          <span style={{ color: demoMode ? '#22c55e' : '#6b7280', fontSize: 10 }}>{demoMode ? '● 模拟' : '○ 实时'}</span>
        </div>
        <div style={{ maxHeight: 280, overflowY: 'auto' }}>
          {sensors.map(s => (
            <div key={s.id} style={{
              padding: '7px 14px', borderBottom: '1px solid rgba(255,255,255,0.04)',
              display: 'flex', justifyContent: 'space-between', alignItems: 'center',
              borderLeft: `3px solid ${levelColor(s.level)}`,
            }}>
              <div>
                <div style={{ fontSize: 11, color: '#9ca3af' }}>{s.locationName}</div>
                <div style={{ fontSize: 11, color: '#6b7280' }}>{s.gasName} ({s.gasType})</div>
              </div>
              <div style={{ textAlign: 'right' }}>
                <div style={{ fontSize: 15, fontWeight: 700, color: levelColor(s.level) }}>{fmt(s.value)}<span style={{ fontSize: 10, marginLeft: 2 }}>{s.unit}</span></div>
                <div style={{ fontSize: 10, color: levelColor(s.level) }}>{levelLabel(s.level)}</div>
              </div>
            </div>
          ))}
        </div>
        <button onClick={addManualAlert} style={{
          width: '100%', padding: '8px', background: 'rgba(79,142,255,0.15)',
          border: 'none', borderTop: '1px solid rgba(79,142,255,0.2)',
          color: '#4f8eff', fontSize: 12, cursor: 'pointer',
        }}>+ 模拟触发报警</button>
      </div>

      {/* ─── Right Panel: Alert List ─── */}
      <div style={{
        position: 'absolute', top: 360, right: 12, width: 260,
        background: 'rgba(5,8,16,0.9)', border: '1px solid rgba(249,115,22,0.3)',
        borderRadius: 10, backdropFilter: 'blur(12px)',
        overflow: 'hidden', zIndex: 10,
      }}>
        <div style={{ padding: '10px 14px', borderBottom: '1px solid rgba(255,255,255,0.06)', fontSize: 12, fontWeight: 600, color: '#f97316', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span>🚨 报警记录 ({alerts.length})</span>
        </div>
        <div style={{ maxHeight: 240, overflowY: 'auto' }}>
          {alerts.length === 0 && (
            <div style={{ padding: '20px 14px', textAlign: 'center', color: '#444', fontSize: 12 }}>暂无报警</div>
          )}
          {alerts.map(a => (
            <div key={a.id} style={{
              padding: '7px 14px', borderBottom: '1px solid rgba(255,255,255,0.04)',
              borderLeft: `3px solid ${levelColor(a.level)}`,
            }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11 }}>
                <span style={{ color: '#9ca3af' }}>{a.locationName}</span>
                <span style={{ color: levelColor(a.level), fontWeight: 600 }}>{a.gasName} {fmt(a.value)}{a.unit}</span>
              </div>
              <div style={{ fontSize: 10, color: '#6b7280', marginTop: 2 }}>
                <span style={{ color: levelColor(a.level) }}>⚠ {levelLabel(a.level)}</span>
                <span style={{ marginLeft: 8 }}>{new Date(a.createdAt).toLocaleTimeString()}</span>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* ─── Bottom left: Legend ─── */}
      <div style={{
        position: 'absolute', bottom: 40, left: 12, width: 200,
        background: 'rgba(5,8,16,0.85)', border: '1px solid rgba(255,255,255,0.08)',
        borderRadius: 8, backdropFilter: 'blur(8px)',
        padding: '10px 14px', fontSize: 11, zIndex: 10,
      }}>
        <div style={{ color: '#6b7280', marginBottom: 6, fontWeight: 600 }}>图例</div>
        {[['#22c55e', '正常'], ['#84cc16', '注意'], ['#eab308', '警戒'], ['#f97316', '警告'], ['#ef4444', '危险']].map(([c, l]) => (
          <div key={l} style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 3, color: '#9ca3af' }}>
            <div style={{ width: 8, height: 8, borderRadius: '50%', background: c, boxShadow: `0 0 6px ${c}` }} />
            {l}
          </div>
        ))}
        <div style={{ color: '#4b5563', marginTop: 6 }}>🖱️ 拖拽旋转 · 滚轮缩放</div>
      </div>

      {/* ─── Demo Mode Toggle ─── */}
      <div style={{
        position: 'absolute', bottom: 40, right: 12,
        zIndex: 10,
      }}>
        <button onClick={() => setDemoMode(d => !d)} style={{
          padding: '6px 14px', borderRadius: 6,
          background: demoMode ? 'rgba(34,197,94,0.2)' : 'rgba(107,114,128,0.2)',
          border: `1px solid ${demoMode ? 'rgba(34,197,94,0.4)' : 'rgba(107,114,128,0.3)'}`,
          color: demoMode ? '#22c55e' : '#9ca3af',
          fontSize: 12, cursor: 'pointer',
        }}>
          {demoMode ? '🟢 演示模式' : '⚪ 实时模式'}
        </button>
      </div>

    </div>
  )
}
