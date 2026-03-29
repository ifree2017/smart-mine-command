# SmartMine-Command — STATE.md

## 项目信息
- **项目名**：智慧矿山指挥调度平台
- **代号**：SMC-001
- **目录**：`smart-mine-command/`
- **状态**：SDD 编写中
- **创建时间**：2026-03-29
- **更新**：2026-03-29

## 里程碑（M1-M5）

| 里程碑 | 状态 | 完成日期 |
|--------|------|---------|
| SDD 评审 | ✅ 通过（GAN v2）| 2026-03-29 |
| M1 | ✅ 完成 | 2026-03-29 |
| M2 | ⏳ 待开始 | — |
| M3 | ⏳ 待开始 | — |
| M4 | ⏳ 待开始 | — |
| M5 | ⏳ 待开始 | — |

## 资产复用

| 已有项目 | 复用内容 |
|---------|---------|
| digital-twin-mine | 3D可视化引擎，WebSocket实时数据，仿真器 |
| free-dispatch | 调度内核，事件总线，协议解析，kj90x对接 |
| QuantAgent | AI分析流程，知识库架构 |
| TCM-AI | 知识库+推荐算法 |
| IceBreak | WebSocket推送，实时通信 |

## 技术栈

| 模块 | 技术 |
|------|------|
| 前端 | React + TypeScript + Vite（复用数字孪生）|
| 后端 | Go 1.24 + Gin |
| 3D渲染 | Three.js（复用）|
| AI | LLM API（OpenAI/Groq）+ 本地知识库 |
| 实时通信 | WebSocket |
| 调度内核 | FreeDispatch Go（复用）|
| 数据库 | PostgreSQL |

## M1 任务

| Task | 描述 | 状态 |
|------|------|------|
| T1 | 项目初始化（Go模块）| ✅ |
| T2 | 事件总线（EventBus）| ✅ |
| T3 | kj90x协议解析 | ✅ |
| T4 | 调度内核 | ✅ |
| T5 | 数据模型 | ✅ |
| T6 | 存储层（JSON持久化）| ✅ |
| T7 | HTTP API + WebSocket | ✅ |
| T8 | TCP Server（kj90x接入）| ✅ |
| T9 | 前端3D一张图 | ✅ |
| T10 | 前端报警面板 | ✅ |

## GitHub
待创建仓库：https://github.com/ifree2017/smart-mine-command
本地已提交：`3e6d10d`
