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

  const [selectedSource, setSelectedSource] = useState(null)
  const [selectedTarget, setSelectedTarget] = useState(null)

  const getHeaders = () => ({
    'Content-Type': 'application/json',
    'X-Session-ID': localStorage.getItem('session_id') || '',
  })

  useEffect(() => {
    const sessionId = localStorage.getItem('session_id')
    if (sessionId) {
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

    const loadTasks = () => {
      fetch('/api/tasks', {
        headers: getHeaders(),
        credentials: 'include',
      })
        .then(res => res.json())
        .then(data => setTasks(data.tasks || []))
    }

    loadTasks()
    const interval = setInterval(loadTasks, 3000)
    return () => clearInterval(interval)
  }, [isAuthenticated])

  const handleSourceSelect = (fileInfo) => {
    if (!fileInfo.isDir) {
      setSelectedSource(fileInfo)
    }
  }

  const handleTargetSelect = (dirInfo) => {
    setSelectedTarget(dirInfo)
  }

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
      .then(data => {
        if (data.task) {
          setSelectedSource(null)
        }
        fetch('/api/tasks', {
          headers: getHeaders(),
          credentials: 'include',
        })
          .then(res => res.json())
          .then(data => setTasks(data.tasks || []))
      })
      .catch(err => console.error('Transfer failed:', err))
  }, [sourceDriver, targetDriver])

  const handleQuickTransfer = useCallback(() => {
    if (!selectedSource || !selectedTarget) return
    handleTransfer(selectedSource.path, selectedTarget.path, selectedSource.name)
  }, [selectedSource, selectedTarget, handleTransfer])

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

  if (!isAuthenticated) {
    return <LoginPage onLogin={handleLogin} />
  }

  if (showBinding) {
    return <BindingPage onLogout={handleLogout} onBack={() => setShowBinding(false)} />
  }

  const canQuickTransfer = selectedSource && selectedTarget && !selectedSource.isDir && selectedTarget.isDir

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
            title="📤 源网盘 (选择文件)"
            driver={sourceDriver}
            onDriverChange={setSourceDriver}
            drivers={drivers}
            onFileSelect={handleSourceSelect}
            selectedInfo={selectedSource}
            selectMode="source"
          />
          <div className="arrow-area">
            <div className="arrow">→</div>
            {canQuickTransfer && (
              <button className="btn btn-quick-transfer" onClick={handleQuickTransfer}>
                ⚡ 一键传输
              </button>
            )}
          </div>
          <FileBrowser
            title="📥 目标网盘 (选择目录)"
            driver={targetDriver}
            onDriverChange={setTargetDriver}
            drivers={drivers}
            onFileSelect={handleTargetSelect}
            selectedInfo={selectedTarget}
            selectMode="target"
          />
        </div>

        <TransferPanel
          sourceDriver={sourceDriver}
          targetDriver={targetDriver}
          selectedSource={selectedSource}
          selectedTarget={selectedTarget}
          onTransfer={handleTransfer}
          onClearSelection={() => { setSelectedSource(null); setSelectedTarget(null); }}
        />

        <TaskList tasks={tasks} onCancel={handleCancelTask} />
      </main>
    </div>
  )
}

export default App
