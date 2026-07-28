import React, { useState, useEffect } from 'react'
import './BindingPage.css'

function BindingPage({ onLogout }) {
  const [config, setConfig] = useState(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState(null)
  const [error, setError] = useState('')
  const [mountPoints, setMountPoints] = useState([])
  const [showPassword, setShowPassword] = useState(false)
  const [changingPassword, setChangingPassword] = useState(false)
  const [passwordForm, setPasswordForm] = useState({ old: '', new: '', confirm: '' })

  useEffect(() => {
    loadConfig()
  }, [])

  const getHeaders = () => ({
    'Content-Type': 'application/json',
    'X-Session-ID': localStorage.getItem('session_id') || '',
  })

  const loadConfig = async () => {
    setLoading(true)
    setError('')
    try {
      const response = await fetch('/api/settings', {
        headers: getHeaders(),
        credentials: 'include',
      })
      
      if (response.status === 401) {
        handleLogout()
        return
      }

      if (response.ok) {
        const data = await response.json()
        setConfig(data)
      } else {
        setError('加载配置失败')
      }
    } catch (err) {
      setError('网络错误')
    } finally {
      setLoading(false)
    }
  }

  const handleLogout = async () => {
    try {
      await fetch('/api/logout', {
        method: 'POST',
        headers: getHeaders(),
        credentials: 'include',
      })
    } catch (err) {
      // ignore
    }
    localStorage.removeItem('session_id')
    onLogout()
  }

  const updateField = (path, value) => {
    const newConfig = { ...config }
    const keys = path.split('.')
    let obj = newConfig
    for (let i = 0; i < keys.length - 1; i++) {
      obj = { ...obj[keys[i]] }
      obj[keys[i]] = obj[keys[i]]
    }
    obj[keys[keys.length - 1]] = value
    setConfig(newConfig)
  }

  const handleSave = async () => {
    setSaving(true)
    setError('')
    try {
      const response = await fetch('/api/settings', {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify(config),
        credentials: 'include',
      })

      if (response.ok) {
        alert('配置保存成功')
        loadConfig()
      } else {
        const data = await response.json()
        setError(data.error || '保存失败')
      }
    } catch (err) {
      setError('网络错误')
    } finally {
      setSaving(false)
    }
  }

  const handleTestConnection = async () => {
    if (!config?.openlist) return

    setTesting(true)
    setTestResult(null)
    setError('')

    try {
      const response = await fetch('/api/settings/test', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          base_url: config.openlist.base_url,
          username: config.openlist.username,
          password: showPassword ? config.openlist.password : '',
        }),
      })

      const data = await response.json()
      setTestResult(data)

      if (data.connected && data.mount_points) {
        setMountPoints(data.mount_points)
      }
    } catch (err) {
      setTestResult({ connected: false, error: '网络错误' })
    } finally {
      setTesting(false)
    }
  }

  const handleChangePassword = async (e) => {
    e.preventDefault()
    if (passwordForm.new !== passwordForm.confirm) {
      setError('两次输入的新密码不一致')
      return
    }
    if (passwordForm.new.length < 6) {
      setError('新密码至少6位')
      return
    }

    try {
      const response = await fetch('/api/settings/password', {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({
          old_password: passwordForm.old,
          new_password: passwordForm.new,
        }),
        credentials: 'include',
      })

      if (response.ok) {
        alert('密码修改成功')
        setChangingPassword(false)
        setPasswordForm({ old: '', new: '', confirm: '' })
        setError('')
      } else {
        const data = await response.json()
        setError(data.error || '修改失败')
      }
    } catch (err) {
      setError('网络错误')
    }
  }

  const autoFillMounts = (driver) => {
    if (mountPoints.length > 0) {
      updateField(`openlist.mounts.${driver}`, mountPoints[0])
    }
  }

  if (loading) {
    return <div className="loading">加载配置中...</div>
  }

  if (!config) {
    return <div className="loading">无配置数据</div>
  }

  return (
    <div className="binding-container">
      {/* 头部 */}
      <div className="binding-header">
        <h1>🔗 连接配置</h1>
        <div className="header-actions">
          <button className="btn-text" onClick={() => setChangingPassword(!changingPassword)}>
            {changingPassword ? '取消' : '修改密码'}
          </button>
          <button className="btn-text" onClick={handleLogout}>
            退出登录
          </button>
        </div>
      </div>

      {error && <div className="error-banner">{error}</div>}

      {/* 修改密码面板 */}
      {changingPassword && (
        <div className="password-panel">
          <h3>修改管理员密码</h3>
          <form onSubmit={handleChangePassword} className="password-form">
            <div className="form-group">
              <label>当前密码</label>
              <input
                type="password"
                value={passwordForm.old}
                onChange={(e) => setPasswordForm({ ...passwordForm, old: e.target.value })}
                required
              />
            </div>
            <div className="form-group">
              <label>新密码</label>
              <input
                type="password"
                value={passwordForm.new}
                onChange={(e) => setPasswordForm({ ...passwordForm, new: e.target.value })}
                required
                minLength={6}
              />
            </div>
            <div className="form-group">
              <label>确认新密码</label>
              <input
                type="password"
                value={passwordForm.confirm}
                onChange={(e) => setPasswordForm({ ...passwordForm, confirm: e.target.value })}
                required
                minLength={6}
              />
            </div>
            <button type="submit" className="btn-primary">确认修改</button>
          </form>
        </div>
      )}

      {/* OpenList 连接配置 */}
      <div className="config-card">
        <div className="config-card-header">
          <div className="config-icon">🔗</div>
          <div>
            <h2>OpenList 连接设置</h2>
            <p className="config-desc">配置 OpenList 服务器连接信息</p>
          </div>
        </div>

        <div className="form-section">
          <div className="form-group full">
            <label>OpenList 地址</label>
            <input
              type="text"
              value={config.openlist.base_url}
              onChange={(e) => updateField('openlist.base_url', e.target.value)}
              placeholder="http://your-openlist:5244"
            />
          </div>

          <div className="form-row">
            <div className="form-group">
              <label>登录用户名</label>
              <input
                type="text"
                value={config.openlist.username}
                onChange={(e) => updateField('openlist.username', e.target.value)}
                placeholder="admin"
              />
            </div>
            <div className="form-group">
              <label>登录密码</label>
              <div className="password-input-wrapper">
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={config.openlist.password === '******' ? '' : config.openlist.password}
                  onChange={(e) => updateField('openlist.password', e.target.value)}
                  placeholder="留空则不修改"
                />
                <button
                  type="button"
                  className="toggle-password"
                  onClick={() => setShowPassword(!showPassword)}
                >
                  {showPassword ? '🙈' : '👁️'}
                </button>
              </div>
              <small className="hint">密码将使用 AES-GCM 加密存储</small>
            </div>
          </div>

          <div className="form-row">
            <div className="form-group">
              <label>夸克网盘挂载路径</label>
              <div className="mount-input">
                <input
                  type="text"
                  value={config.openlist.mounts.quark}
                  onChange={(e) => updateField('openlist.mounts.quark', e.target.value)}
                  placeholder="/quark"
                />
                {mountPoints.length > 0 && (
                  <select
                    value={config.openlist.mounts.quark}
                    onChange={(e) => updateField('openlist.mounts.quark', e.target.value)}
                  >
                    <option value="">选择挂载点</option>
                    {mountPoints.map((mp) => (
                      <option key={mp} value={`/${mp}`}>{mp}</option>
                    ))}
                  </select>
                )}
              </div>
            </div>
            <div className="form-group">
              <label>移动云盘挂载路径</label>
              <div className="mount-input">
                <input
                  type="text"
                  value={config.openlist.mounts.mobile}
                  onChange={(e) => updateField('openlist.mounts.mobile', e.target.value)}
                  placeholder="/mobile"
                />
                {mountPoints.length > 0 && (
                  <select
                    value={config.openlist.mounts.mobile}
                    onChange={(e) => updateField('openlist.mounts.mobile', e.target.value)}
                  >
                    <option value="">选择挂载点</option>
                    {mountPoints.map((mp) => (
                      <option key={mp} value={`/${mp}`}>{mp}</option>
                    ))}
                  </select>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* 操作按钮 */}
        <div className="action-buttons">
          <button
            className="btn-secondary"
            onClick={handleTestConnection}
            disabled={testing}
          >
            {testing ? '测试中...' : '🔍 测试连接'}
          </button>
          <button
            className="btn-primary"
            onClick={handleSave}
            disabled={saving}
          >
            {saving ? '保存中...' : '💾 保存配置'}
          </button>
        </div>

        {/* 测试结果 */}
        {testResult && (
          <div className={`test-result ${testResult.connected ? 'success' : 'error'}`}>
            {testResult.connected ? (
              <>
                <div className="result-icon">✅</div>
                <div>
                  <strong>连接成功！</strong>
                  <p>{testResult.message}</p>
                  {testResult.mount_points && testResult.mount_points.length > 0 && (
                    <div className="mount-points">
                      <span>发现挂载点：</span>
                      {testResult.mount_points.map((mp) => (
                        <span key={mp} className="mount-tag">{mp}</span>
                      ))}
                    </div>
                  )}
                </div>
              </>
            ) : (
              <>
                <div className="result-icon">❌</div>
                <div>
                  <strong>连接失败</strong>
                  <p>{testResult.error}</p>
                </div>
              </>
            )}
          </div>
        )}
      </div>

      {/* 已绑定的网盘驱动 */}
      <div className="config-card">
        <h2>已绑定的网盘驱动</h2>
        
        <div className="driver-list">
          <div className="driver-item">
            <div className="driver-icon quark">☁️</div>
            <div className="driver-info">
              <h4>夸克网盘</h4>
              <p>挂载路径：<code>{config.openlist.mounts.quark}</code></p>
              <p>已授权</p>
            </div>
            <div className={`driver-status ${config.openlist.password ? 'connected' : 'disconnected'}`}>
              {config.openlist.password ? '● 已配置' : '○ 未配置'}
            </div>
          </div>

          <div className="driver-item">
            <div className="driver-icon mobile">📱</div>
            <div className="driver-info">
              <h4>移动云盘</h4>
              <p>挂载路径：<code>{config.openlist.mounts.mobile}</code></p>
              <p>已授权</p>
            </div>
            <div className={`driver-status ${config.openlist.password ? 'connected' : 'disconnected'}`}>
              {config.openlist.password ? '● 已配置' : '○ 未配置'}
            </div>
          </div>
        </div>
      </div>

      {/* 设计说明 */}
      <div className="info-card">
        <h3>💡 为什么这样设计？</h3>
        <ul>
          <li><strong>无需重复配置：</strong>夸克/移动云盘的账号、Cookie、AppID 等都只需在 OpenList 中配置一次</li>
          <li><strong>统一管理：</strong>通过 OpenList 的 REST API 操作所有网盘，本项目只负责传输调度</li>
          <li><strong>自动秒传：</strong>调用 OpenList 的 Copy 接口时，OpenList 内部会自动判断是否可秒传</li>
          <li><strong>易于扩展：</strong>未来支持更多网盘时，只需在 OpenList 中添加挂载点，无需修改本项目代码</li>
          <li><strong>安全加密：</strong>密码使用 AES-GCM 加密存储，会话管理防止未授权访问</li>
        </ul>
      </div>
    </div>
  )
}

export default BindingPage
