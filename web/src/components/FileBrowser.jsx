import React, { useState, useEffect, useCallback } from 'react'

function FileBrowser({ title, driver, onDriverChange, drivers }) {
  const [path, setPath] = useState('/')
  const [files, setFiles] = useState([])
  const [loading, setLoading] = useState(false)
  const [selectedFile, setSelectedFile] = useState(null)

  const loadFiles = useCallback(() => {
    if (!driver) return
    setLoading(true)
    fetch(`/api/files/${driver}?path=${encodeURIComponent(path)}`)
      .then(res => res.json())
      .then(data => setFiles(data.files || []))
      .catch(err => console.error('Failed to load files:', err))
      .finally(() => setLoading(false))
  }, [driver, path])

  useEffect(() => {
    loadFiles()
  }, [loadFiles])

  const handleFileClick = (file) => {
    if (file.is_dir) {
      setPath(file.path)
    } else {
      setSelectedFile(file)
    }
  }

  const handleGoUp = () => {
    const parts = path.split('/').filter(Boolean)
    parts.pop()
    setPath('/' + parts.join('/'))
  }

  const formatSize = (bytes) => {
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
    return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
  }

  return (
    <div className="file-browser">
      <h3>{title}</h3>

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

      {selectedFile && (
        <div style={{ padding: '8px', background: '#e8f5e9', borderRadius: '8px', marginBottom: '12px' }}>
          <strong>已选文件:</strong> {selectedFile.name} ({formatSize(selectedFile.size)})
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
              className="file-item"
              onClick={() => handleFileClick(file)}
            >
              <span className="file-icon">{file.is_dir ? '📁' : '📄'}</span>
              <span className="file-name">{file.name}</span>
              {!file.is_dir && <span className="file-size">{formatSize(file.size)}</span>}
            </div>
          ))
        )}
      </div>
    </div>
  )
}

export default FileBrowser
