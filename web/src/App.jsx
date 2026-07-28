import React, { useState, useEffect, useCallback } from 'react'
import FileBrowser from './components/FileBrowser.jsx'
import TransferPanel from './components/TransferPanel.jsx'
import TaskList from './components/TaskList.jsx'
import './App.css'

function App() {
  const [drivers, setDrivers] = useState([])
  const [tasks, setTasks] = useState([])
  const [sourceDriver, setSourceDriver] = useState('quark')
  const [targetDriver, setTargetDriver] = useState('mobile')

  useEffect(() => {
    fetch('/api/drivers')
      .then(res => res.json())
      .then(data => setDrivers(data.drivers || []))
      .catch(err => console.error('Failed to load drivers:', err))
  }, [])

  useEffect(() => {
    fetch('/api/tasks')
      .then(res => res.json())
      .then(data => setTasks(data.tasks || []))
      .catch(err => console.error('Failed to load tasks:', err))

    const interval = setInterval(() => {
      fetch('/api/tasks')
        .then(res => res.json())
        .then(data => setTasks(data.tasks || []))
    }, 3000)

    return () => clearInterval(interval)
  }, [])

  const handleTransfer = useCallback((sourcePath, targetPath, fileName) => {
    fetch('/api/transfer', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        source_driver: sourceDriver,
        source_path: sourcePath,
        target_driver: targetDriver,
        target_path: targetPath,
        file_name: fileName
      })
    })
      .then(res => res.json())
      .then(() => {
        fetch('/api/tasks')
          .then(res => res.json())
          .then(data => setTasks(data.tasks || []))
      })
      .catch(err => console.error('Transfer failed:', err))
  }, [sourceDriver, targetDriver])

  const handleCancelTask = useCallback((taskId) => {
    fetch(`/api/tasks/${taskId}`, { method: 'DELETE' })
      .then(() => {
        fetch('/api/tasks')
          .then(res => res.json())
          .then(data => setTasks(data.tasks || []))
      })
      .catch(err => console.error('Cancel failed:', err))
  }, [])

  return (
    <div className="app">
      <header className="app-header">
        <h1>☁️ 夸克 ↔ 移动云盘传输工具</h1>
        <div className="driver-info">
          <span>已注册驱动:</span>
          {drivers.map(d => (
            <span key={d.type} className="driver-tag">{d.type}</span>
          ))}
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
