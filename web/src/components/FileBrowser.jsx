import React, { useState, useEffect, useCallback } from 'react'

function FileBrowser({ title, driver, onDriverChange, drivers, onFileSelect, selectedInfo, selectMode }) {
  const [path, setPath] = useState('/')
  const [files, setFiles] = useState([])
  const [loading, setLoading] = useState(false)
  const [selectedFile, setSelectedFile] = useState(null)

  const loadFiles = useCallback(() => {
    if (!driver) return
    setLoading(true)
    fetch(`/api/files/${driver}?path=${encodeURIComponent(path)}`, {
      headers: {
        'X-Session-ID': localStorage.getItem('session_id') || '',
      },
      credentials: 'include',
    })
      .then(res => res.json())
      .then(data => setFiles(data.files || []))
      .catch(err => console.error('Failed to load files:', err))
      .finally(() => setLoading(false))
  }, [driver, path])

  useEffect(() => {
    loadFiles()
  }, [loadFiles])

  const getFullPath = (file) => {
    if (file.is_dir) {
      const sep = path.endsWith('/') ? '' : '/'
      return `${path}${sep}${file.name}`
    }
    return file.path || `${path}/${file.name}`
  }

  const handleFileClick = (file) => {
    if (file.is_dir) {
      setPath(getFullPath(file))
    } else {
      selectFile(file)
    }
  }

  const selectFile = (file) => {
    const fullPath = getFullPath(file)
    setSelectedFile(file)
    if (onFileSelect) {
      onFileSelect({
        name: file.name,
        path: fullPath,
        size: file.size,
        isDir: file.is_dir,
      })
    }
  }

  const handleSelectCurrentDir = () => {
    if (onFileSelect) {
      onFileSelect({
        name: path.split('/').pop() || '/',
        path: path,
        isDir: true,
      })
    }
  }

  const handleGoUp = () => {
    const parts = path.split('/').filter(Boolean)
    parts.pop()
    setPath('/' + parts.join('/'))
    setSelectedFile(null)
  }

  const formatSize = (bytes) => {
    if (!bytes || bytes === 0) return ''
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
    return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
  }

  const isSelected = (file) => {
    if (!selectedInfo) return false
    return selectedInfo.path === getFullPath(file)
  }

  return (
    <div className="file-browser">
      <div className="browser-header">
        <h3>{title}</h3>
        {selectedInfo && (
          <div className="selected-indicator">
            <span className="selected-badge">
              {selectMode === 'source' ? '📎 已选源' : '📍 已选目标'}
            </span>
            <span className="selected-name" title={selectedInfo.path}>
              {selectedInfo.name}
            </span>
          </div>
        )}
      </div>

      <select
        className="driver-select"
        value={driver}
        onChange={(e) => {
          onDriverChange(e.target.value)
          setPath('/')
          setSelectedFile(null)
        }}
      >
        {drivers.map(d => (
          <option key={d.type} value={d.type}>{d.type}</option>
        ))}
      </select>

      <div style={{ display: 'flex', gap: '8px', marginBottom: '12px' }}>
        <button onClick={handleGoUp} disabled={path === '/'} className="btn btn-primary" style={{ padding: '8px 12px' }}>
          ↑ 上级
        </button>
        <input
          className="path-input"
          value={path}
          onChange={(e) => setPath(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && loadFiles()}
          placeholder="路径"
        />
      </div>

      {selectMode === 'target' && (
        <div className="current-dir-actions">
          <button className="btn-select-dir" onClick={handleSelectCurrentDir}>
            📌 选择此目录作为目标
          </button>
        </div>
      )}

      <div className="file-list">
        {loading ? (
          <div className="empty">加载中...</div>
        ) : files.length === 0 ? (
          <div className="empty">暂无文件</div>
        ) : (
          files.map((file, idx) => (
            <div
              key={idx}
              className={`file-item ${isSelected(file) ? 'selected' : ''}`}
              onClick={() => handleFileClick(file)}
            >
              <span className="file-icon">{file.is_dir ? '📁' : '📄'}</span>
              <span className="file-name">{file.name}</span>
              {!file.is_dir && <span className="file-size">{formatSize(file.size)}</span>}
              {!file.is_dir && (
                <span className="file-action">
                  {isSelected(file) ? '✓ 已选' : '选择'}
                </span>
              )}
            </div>
          ))
        )}
      </div>

      {selectedFile && !selectedFile.is_dir && (
        <div className="selected-file-bar">
          <span>✅ 已选择: <strong>{selectedFile.name}</strong> ({formatSize(selectedFile.size)})</span>
        </div>
      )}
    </div>
  )
}

export default FileBrowser
