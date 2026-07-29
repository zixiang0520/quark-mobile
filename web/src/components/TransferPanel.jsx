import React, { useState, useEffect } from 'react'

function TransferPanel({ sourceDriver, targetDriver, selectedSource, selectedTarget, onTransfer, onClearSelection }) {
  const [sourcePath, setSourcePath] = useState('/')
  const [targetPath, setTargetPath] = useState('/')
  const [fileName, setFileName] = useState('')
  const [useCustom, setUseCustom] = useState(false)

  useEffect(() => {
    if (selectedSource && !useCustom) {
      setSourcePath(selectedSource.path)
      setFileName(selectedSource.name)
    }
  }, [selectedSource, useCustom])

  useEffect(() => {
    if (selectedTarget && !useCustom) {
      setTargetPath(selectedTarget.path)
    }
  }, [selectedTarget, useCustom])

  const handleSubmit = (e) => {
    e.preventDefault()
    if (!sourcePath || !targetPath) {
      alert('请填写源路径和目标路径')
      return
    }
    onTransfer(sourcePath, targetPath, fileName || undefined)
    if (onClearSelection) {
      onClearSelection()
    }
  }

  const hasSelection = selectedSource && selectedTarget

  return (
    <div className="transfer-panel">
      <div className="panel-header">
        <h3>📤 创建传输任务</h3>
        {hasSelection && !useCustom && (
          <button className="btn btn-text" onClick={() => { setUseCustom(true) }}>
            ✏️ 手动输入路径
          </button>
        )}
        {useCustom && (
          <button className="btn btn-text" onClick={() => { setUseCustom(false) }}>
            🔄 使用选中文件
          </button>
        )}
      </div>

      {hasSelection && !useCustom && (
        <div className="selection-summary">
          <div className="selection-item">
            <span className="selection-label">源文件:</span>
            <span className="selection-value">
              <strong>{selectedSource.name}</strong> ({selectedSource.size > 0 ? formatSize(selectedSource.size) : ''})
            </span>
            <span className="selection-path">{selectedSource.path}</span>
          </div>
          <div className="selection-arrow">→</div>
          <div className="selection-item">
            <span className="selection-label">目标目录:</span>
            <span className="selection-value">
              <strong>{selectedTarget.path === '/' ? '/ (根目录)' : selectedTarget.path}</strong>
            </span>
          </div>
        </div>
      )}

      <form onSubmit={handleSubmit} className="transfer-form">
        <div style={{ flex: 1, minWidth: '200px' }}>
          <label style={{ fontSize: '12px', color: '#666', display: 'block', marginBottom: '4px' }}>
            源路径 ({sourceDriver})
          </label>
          <input
            value={sourcePath}
            onChange={(e) => { setSourcePath(e.target.value); setUseCustom(true) }}
            placeholder="/path/to/file"
          />
        </div>
        <div style={{ flex: 1, minWidth: '200px' }}>
          <label style={{ fontSize: '12px', color: '#666', display: 'block', marginBottom: '4px' }}>
            目标路径 ({targetDriver})
          </label>
          <input
            value={targetPath}
            onChange={(e) => { setTargetPath(e.target.value); setUseCustom(true) }}
            placeholder="/target/path"
          />
        </div>
        <div style={{ flex: 1, minWidth: '200px' }}>
          <label style={{ fontSize: '12px', color: '#666', display: 'block', marginBottom: '4px' }}>
            文件名 (可选)
          </label>
          <input
            value={fileName}
            onChange={(e) => setFileName(e.target.value)}
            placeholder="留空则使用原文件名"
          />
        </div>
        <button type="submit" className="btn btn-primary" style={{ marginTop: '22px' }}>
          🚀 开始传输
        </button>
      </form>
    </div>
  )
}

function formatSize(bytes) {
  if (!bytes) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
  return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
}

export default TransferPanel
