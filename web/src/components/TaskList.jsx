import React from 'react'

function TaskList({ tasks, onCancel }) {
  const formatSize = (bytes) => {
    if (!bytes) return '-'
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
    return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
  }

  const getStatusClass = (status) => {
    switch (status) {
      case 'pending': return 'status-pending'
      case 'running': return 'status-running'
      case 'completed': return 'status-completed'
      case 'failed': return 'status-failed'
      case 'cancelled': return 'status-cancelled'
      default: return ''
    }
  }

  const getStatusText = (task) => {
    if (task.instant_done) return '⚡ 秒传完成'
    switch (task.status) {
      case 'pending': return '等待中'
      case 'running': return '传输中'
      case 'completed': return '✅ 完成'
      case 'failed': return '❌ 失败'
      case 'cancelled': return '已取消'
      default: return task.status
    }
  }

  return (
    <div className="task-list">
      <h3>📋 传输任务</h3>
      {tasks.length === 0 ? (
        <div className="empty">暂无任务</div>
      ) : (
        tasks.map(task => (
          <div key={task.id} className="task-item">
            <div className="task-info">
              <div className="task-path">
                {task.source_driver}: {task.source_path} → {task.target_driver}: {task.target_path}/{task.file_name}
              </div>
              <div className="task-detail">
                {formatSize(task.file_size)} | {task.sha256 ? task.sha256.slice(0, 16) + '...' : '-'}
                {task.error && <span style={{ color: '#ff4757', marginLeft: '8px' }}>| {task.error}</span>}
              </div>
            </div>
            <span className={`task-status ${getStatusClass(task.status)}`}>
              {getStatusText(task)}
            </span>
            {task.status === 'running' && (
              <div className="progress-bar">
                <div className="progress-fill" style={{ width: `${task.progress}%` }}></div>
              </div>
            )}
            {task.status === 'running' && (
              <button
                className="btn btn-danger"
                onClick={() => onCancel(task.id)}
              >
                取消
              </button>
            )}
          </div>
        ))
      )}
    </div>
  )
}

export default TaskList
