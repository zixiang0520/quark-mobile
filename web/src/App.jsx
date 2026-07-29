import React, { useState, useEffect, useCallback } from 'react'
import FileBrowser from './components/FileBrowser.jsx'
import TransferPanel from './components/TransferPanel.jsx'
import TaskList from './components/TaskList.jsx'
import LoginPage from './components/LoginPage.jsx'
import BindingPage from './components/BindingPage.jsx'
import './App.css'

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [showBinding, setShowBinding] = useState(false)
  const [drivers, setDrivers] = useState([])
  const [tasks, setTasks] = useState([])
  const [sourceDriver, setSourceDriver] = useState('quark')
  const [targetDriver, setTargetDriver] = useState('mobile')

  const getHeaders = () => ({
    'Content-Type': 'application/json',
    'X-Session-ID': localStorage.getItem('session_id') || '',
  })

  useEffect(() => {
    // 检查本地是否有 session
    const sessionId = localStorage.getItem('session_id')
    if (sessionId) {
      // 验证 session 是否有效
      fetch('/api/settings', {
        headers: getHeaders(),
        credentials: 'include',
      })
        .then(res => {
          if (res.ok) {
            setIsAuthenticated(true)
          } else {
            localStorage.removeItem('session_id')
          }
        })
        .catch(() => {
          localStorage.removeItem('session_id')
        })
    }
  }, [])

  useEffect(() => {
    if (!isAuthenticated) return

    fetch('/api/drivers', {
      headers: getHeaders(),
      credentials: 'include',
    })
      .then(res => res.json())
      .then(data => setDrivers(data.drivers || []))
      .catch(err => console.error('Failed to load drivers:', err))
  }, [isAuthenticated])

  useEffect(() => {
    if (!isAuthenticated) return

    fetch('/api/tasks', {
      headers: getHeaders(),
      credentials: 'include',
    })
      .then(res => res.json())
      .then(data => setTasks(data.tasks || []))
      .catch(err => console.error('Failed to load tasks:', err))

    const interval = setInterval(() => {
      fetch('/api/tasks', {
        headers: getHeaders(),
        credentials: 'include',
      })
        .then(res => res.json())
        .then(data => setTasks(data.tasks || []))
    }, 3000)

    return () => clearInterval(interval)
  }, [isAuthenticated])

  const handleTransfer = useCallback((sourcePath, targetPath, fileName) => {
    fetch('/api/transfer', {
      method: 'POST',
      headers: getHeaders(),
      credentials: 'include',
      body: JSON.stringify({
        source_driver: sourceDriver,
        source_path: sourcePath,
        target_driver: targetDriver,
        target_path: targetPath,
        file_name: fileName,
      }),
    })
      .then(res => res.json())
      .then(() => {
        fetch('/api/tasks', {
          headers: getHeaders(),
          credentials: 'include',
        })
          .then(res => res.json())
          .then(data => setTasks(data.tasks || []))
      })
      .catch(err => console.error('Transfer failed:', err))
  }, [sourceDriver, targetDriver])

  const handleCancelTask = useCallback((taskId) => {
    fetch(`/api/tasks/${taskId}`, {
      method: 'DELETE',
      headers: getHeaders(),
      credentials: 'include',
    })
      .then(() => {
        fetch('/api/tasks', {
          headers: getHeaders(),
          credentials: 'include',
        })
          .then(res => res.json())
          .then(data => setTasks(data.tasks || []))
      })
      .catch(err => console.error('Cancel failed:', err))
  }, [])

  const handleLogin = () => {
    setIsAuthenticated(true)
  }

  const handleLogout = () => {
    setIsAuthenticated(false)
    setShowBinding(false)
  }

  // 未登录显示登录页面
  if (!isAuthenticated) {
    return <LoginPage onLogin={handleLogin} />
  }

  // 显示绑定配置页面
  if (showBinding) {
    return <BindingPage onLogout={handleLogout} onBack={() => setShowBinding(false)} />
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1>☁️ 夸克 ↔ 移动云盘传输工具</h1>
        <div className="header-actions">
          <div className="driver-info">
            <span>已注册驱动:</span>
            {drivers.map(d => (
              <span key={d.type} className="driver-tag">{d.type}</span>
            ))}
          </div>
          <div className="action-buttons-header">
            <button className="btn-config" onClick={() => setShowBinding(true)}>
              ⚙️ 配置
            </button>
            <button className="btn-logout" onClick={handleLogout}>
              🚪 退出
            </button>
          </div>
        </div>
      </header>

      <main className="app-main">
        <div className="panels-row">
          <FileBrowser
            title="源网盘"
            driver={sourceDriver}
            onDriverChange={setSourceDriver}
            drivers={drivers}
          />
          <div className="arrow">→</div>
          <FileBrowser
            title="目标网盘"
            driver={targetDriver}
            onDriverChange={setTargetDriver}
            drivers={drivers}
          />
        </div>

        <TransferPanel
          sourceDriver={sourceDriver}
          targetDriver={targetDriver}
          onTransfer={handleTransfer}
        />

        <TaskList tasks={tasks} onCancel={handleCancelTask} />
      </main>
    </div>
  )
}

export default App
