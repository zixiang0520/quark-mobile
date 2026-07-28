import React, { useState } from 'react'

function TransferPanel({ sourceDriver, targetDriver, onTransfer }) {
  const [sourcePath, setSourcePath] = useState('/')
  const [targetPath, setTargetPath] = useState('/')
  const [fileName, setFileName] = useState('')

  const handleSubmit = (e) => {
    e.preventDefault()
    if (!sourcePath || !targetPath) {
      alert('请填写源路径和目标路径')
      return
    }
    onTransfer(sourcePath, targetPath, fileName || undefined)
    setFileName('')
  }

  return (
    <div className="transfer-panel">
      <h3>📤 创建传输任务</h3>
      <form onSubmit={handleSubmit} className="transfer-form">
        <div style={{ flex: 1, minWidth: '200px' }}>
          <label style={{ fontSize: '12px', color: '#666', display: 'block', marginBottom: '4px' }}>
            源路径 ({sourceDriver})
          </label>
          <input
            value={sourcePath}
            onChange={(e) => setSourcePath(e.target.value)}
            placeholder="/path/to/file"
          />
        </div>
        <div style={{ flex: 1, minWidth: '200px' }}>
          <label style={{ fontSize: '12px', color: '#666', display: 'block', marginBottom: '4px' }}>
            目标路径 ({targetDriver})
          </label>
          <input
            value={targetPath}
            onChange={(e) => setTargetPath(e.target.value)}
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
          开始传输
        </button>
      </form>
    </div>
  )
}

export default TransferPanel
