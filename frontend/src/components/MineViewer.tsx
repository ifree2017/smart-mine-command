import { useEffect, useRef, useState, useCallback } from 'react'
import * as THREE from 'three'

export interface Alert {
  id: string
  location: string
  gasType: string
  value: number
  level: number
  status: string
  createdAt: string
}

export default function MineViewer() {
  const mountRef = useRef<HTMLDivElement>(null)
  const rendererRef = useRef<THREE.WebGLRenderer | null>(null)
  const sceneRef = useRef<THREE.Scene | null>(null)
  const cameraRef = useRef<THREE.PerspectiveCamera | null>(null)
  const markersRef = useRef<THREE.Mesh[]>([])
  const animationRef = useRef<number>(0)
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [wsConnected, setWsConnected] = useState(false)

  const addAlertMarker = useCallback((alert: Alert) => {
    const scene = sceneRef.current
    if (!scene) return

    // Remove old marker for same location
    markersRef.current.forEach(m => scene.remove(m))
    markersRef.current = []

    // Color based on level
    const colors = [0x2d5a2d, 0x5a7a2d, 0x7a6a2d, 0x7a3a2d, 0x7a1a1a]
    const color = colors[Math.min(alert.level - 1, 4)]

    // Position: derive from location string hash for demo
    const hash = alert.location.split('').reduce((acc, c) => acc + c.charCodeAt(0), 0)
    const x = (hash % 100) - 50
    const z = ((hash * 7) % 100) - 50

    // Sphere marker
    const geo = new THREE.SphereGeometry(alert.level * 0.8 + 0.5, 16, 16)
    const mat = new THREE.MeshLambertMaterial({ color, emissive: color, emissiveIntensity: 0.5 })
    const mesh = new THREE.Mesh(geo, mat)
    mesh.position.set(x, 2, z)
    scene.add(mesh)
    markersRef.current.push(mesh)

    // Vertical beam
    const beamGeo = new THREE.CylinderGeometry(0.2, 0.2, 20, 8)
    const beamMat = new THREE.MeshLambertMaterial({ color, transparent: true, opacity: 0.3 })
    const beam = new THREE.Mesh(beamGeo, beamMat)
    beam.position.set(x, 10, z)
    scene.add(beam)
    markersRef.current.push(beam)
  }, [])

  useEffect(() => {
    const mount = mountRef.current
    if (!mount) return

    const w = mount.clientWidth
    const h = mount.clientHeight

    // Scene
    const scene = new THREE.Scene()
    scene.background = new THREE.Color(0x0a0a0f)
    scene.fog = new THREE.Fog(0x0a0a0f, 80, 200)
    sceneRef.current = scene

    // Camera
    const camera = new THREE.PerspectiveCamera(75, w / h, 0.1, 1000)
    camera.position.set(0, 50, 100)
    camera.lookAt(0, 0, 0)
    cameraRef.current = camera

    // Renderer
    const renderer = new THREE.WebGLRenderer({ antialias: true })
    renderer.setSize(w, h)
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
    mount.appendChild(renderer.domElement)
    rendererRef.current = renderer

    // Lights
    const ambient = new THREE.AmbientLight(0x334455, 1.5)
    scene.add(ambient)
    const directional = new THREE.DirectionalLight(0xffffff, 1)
    directional.position.set(50, 100, 50)
    scene.add(directional)
    const pointLight = new THREE.PointLight(0x4488ff, 1, 200)
    pointLight.position.set(0, 30, 0)
    scene.add(pointLight)

    // Floor
    const floorGeo = new THREE.PlaneGeometry(200, 200, 20, 20)
    const floorMat = new THREE.MeshLambertMaterial({ color: 0x1a1a2e })
    const floor = new THREE.Mesh(floorGeo, floorMat)
    floor.rotation.x = -Math.PI / 2
    scene.add(floor)

    // Grid helper
    const grid = new THREE.GridHelper(200, 40, 0x223344, 0x1a2233)
    grid.position.y = 0.1
    scene.add(grid)

    // Mine tunnel representations (simplified corridor layout)
    const tunnelMat = new THREE.MeshLambertMaterial({ color: 0x2a2a3e })
    const tunnelPositions: [number, number, number, number, number][] = [
      [-30, 1, -20, 4, 60],  // tunnel 1 - along Z
      [0, 1, -20, 4, 60],    // tunnel 2
      [30, 1, -20, 4, 60],   // tunnel 3
      [-45, 1, 10, 80, 4],   // cross tunnel
      [45, 1, 10, 80, 4],    // cross tunnel
    ]
    tunnelPositions.forEach(([x, y, z, w, d]) => {
      const geo = new THREE.BoxGeometry(w, 2, d)
      const mesh = new THREE.Mesh(geo, tunnelMat)
      mesh.position.set(x, y, z)
      scene.add(mesh)
    })

    // Main shaft
    const shaftGeo = new THREE.CylinderGeometry(5, 5, 20, 16)
    const shaftMat = new THREE.MeshLambertMaterial({ color: 0x333355, wireframe: false })
    const shaft = new THREE.Mesh(shaftGeo, shaftMat)
    shaft.position.set(0, 10, 0)
    scene.add(shaft)

    // OrbitControls (manual implementation - simple mouse drag)
    let isDragging = false
    let prevMouse = { x: 0, y: 0 }
    let theta = 0
    let phi = Math.PI / 4
    let radius = 120

    function updateCamera() {
      camera.position.x = radius * Math.sin(phi) * Math.cos(theta)
      camera.position.y = radius * Math.cos(phi)
      camera.position.z = radius * Math.sin(phi) * Math.sin(theta)
      camera.lookAt(0, 0, 0)
    }
    updateCamera()

    renderer.domElement.addEventListener('mousedown', (e) => {
      isDragging = true
      prevMouse = { x: e.clientX, y: e.clientY }
    })
    window.addEventListener('mousemove', (e) => {
      if (!isDragging) return
      const dx = e.clientX - prevMouse.x
      const dy = e.clientY - prevMouse.y
      theta -= dx * 0.005
      phi -= dy * 0.005
      phi = Math.max(0.1, Math.min(Math.PI - 0.1, phi))
      prevMouse = { x: e.clientX, y: e.clientY }
      updateCamera()
    })
    window.addEventListener('mouseup', () => { isDragging = false })
    renderer.domElement.addEventListener('wheel', (e) => {
      radius += e.deltaY * 0.1
      radius = Math.max(30, Math.min(200, radius))
      updateCamera()
    })

    // Resize handler
    const onResize = () => {
      if (!mount) return
      const nw = mount.clientWidth
      const nh = mount.clientHeight
      camera.aspect = nw / nh
      camera.updateProjectionMatrix()
      renderer.setSize(nw, nh)
    }
    window.addEventListener('resize', onResize)

    // WebSocket
    let ws: WebSocket | null = null
    try {
      const wsUrl = `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/ws`
      ws = new WebSocket(wsUrl)
      ws.onopen = () => setWsConnected(true)
      ws.onclose = () => setWsConnected(false)
      ws.onerror = () => setWsConnected(false)
      ws.onmessage = (e) => {
        try {
          const msg = JSON.parse(e.data)
          if (msg.type === 'alert') {
            setAlerts(prev => [msg.data, ...prev].slice(0, 50))
            addAlertMarker(msg.data)
          }
        } catch {}
      }
    } catch {}

    // Animation loop
    function animate() {
      animationRef.current = requestAnimationFrame(animate)
      // Pulse markers
      markersRef.current.forEach((m, i) => {
        m.scale.setScalar(1 + 0.1 * Math.sin(Date.now() * 0.003 + i))
      })
      renderer.render(scene, camera)
    }
    animate()

    return () => {
      cancelAnimationFrame(animationRef.current)
      ws?.close()
      renderer.dispose()
      window.removeEventListener('resize', onResize)
    }
  }, [addAlertMarker])

  return (
    <div style={{ position: 'relative', width: '100%', height: 'calc(100vh - 52px)' }}>
      <div ref={mountRef} style={{ width: '100%', height: '100%' }} />

      {/* Connection status */}
      <div style={{
        position: 'absolute', top: 12, left: 12, fontSize: 12,
        padding: '4px 10px', borderRadius: 4,
        background: wsConnected ? 'rgba(45,90,45,0.9)' : 'rgba(122,26,26,0.9)',
        color: '#fff'
      }}>
        {wsConnected ? '🟢 WebSocket已连接' : '🔴 WebSocket未连接'}
      </div>

      {/* 3D scene label */}
      <div style={{
        position: 'absolute', bottom: 12, left: 12, fontSize: 12,
        color: 'rgba(255,255,255,0.5)', padding: '4px 10px',
        background: 'rgba(0,0,0,0.4)', borderRadius: 4
      }}>
        🖱️ 拖拽旋转 | 滚轮缩放
      </div>

      {/* Alert overlay */}
      <div className="alert-overlay">
        <h3>🚨 实时报警 ({alerts.length})</h3>
        {alerts.length === 0 && (
          <div style={{ color: '#666', fontSize: 12, padding: '8px 0' }}>暂无报警数据</div>
        )}
        {alerts.map(a => (
          <div key={a.id} className={`alert-item level-${a.level}`}>
            <span className="location">{a.location}</span>
            <span className="value">{a.gasType}={a.value}</span>
            <span className="time">{new Date(a.createdAt).toLocaleTimeString()}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
