import paramiko

# 测试远程服务器的前台功能
SERVER_HOST = "36.140.147.210"
SERVER_USER = "root"
SERVER_PASS = "ZXliukai1."

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(SERVER_HOST, username=SERVER_USER, password=SERVER_PASS, timeout=30)

import json

# 1. 检查容器状态
print("=" * 50)
print("【1】容器状态检查")
stdin, stdout, stderr = client.exec_command("docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'")
print(stdout.read().decode("utf-8", errors="replace"))

# 2. 检查前端静态文件是否存在
print("=" * 50)
print("【2】前端静态文件检查")
stdin, stdout, stderr = client.exec_command("docker exec quark-mobile ls -la /app/web/dist/ 2>/dev/null || echo 'NO DIST'")
print(stdout.read().decode("utf-8", errors="replace"))

# 3. 检查 index.html
print("=" * 50)
print("【3】检查 index.html")
stdin, stdout, stderr = client.exec_command("curl -s http://localhost:18900/ | head -20")
print(stdout.read().decode("utf-8", errors="replace"))

# 4. 完整的前台流程测试
print("=" * 50)
print("【4】完整前台流程测试")

# 4.1 登录
print("  [4.1] 登录...")
login_data = json.dumps({"username": "admin", "password": "admin123"})
cmd = f"""curl -s -X POST http://localhost:18900/api/login \
  -H 'Content-Type: application/json' \
  -d '{login_data}'"""
stdin, stdout, stderr = client.exec_command(cmd)
login_resp = stdout.read().decode("utf-8", errors="replace")
print(f"  登录响应: {login_resp}")

try:
    login_json = json.loads(login_resp)
    token = login_json["session_id"]
    print(f"  Token: {token[:30]}...")
    
    # 4.2 获取驱动列表
    print("\n  [4.2] 获取驱动列表...")
    cmd = f"curl -s http://localhost:18900/api/drivers -H 'X-Session-ID: {token}'"
    stdin, stdout, stderr = client.exec_command(cmd)
    drivers = stdout.read().decode("utf-8", errors="replace")
    print(f"  驱动列表: {drivers}")
    
    # 4.3 列出 Quark 根目录文件
    print("\n  [4.3] 列出 Quark 网盘根目录...")
    cmd = f"curl -s 'http://localhost:18900/api/files/quark?path=%2F' -H 'X-Session-ID: {token}'"
    stdin, stdout, stderr = client.exec_command(cmd)
    quark_files = stdout.read().decode("utf-8", errors="replace")
    print(f"  Quark 文件: {quark_files}")
    
    try:
        qf = json.loads(quark_files)
        print(f"  ✅ Quark 目录正常，共 {len(qf.get('files', []))} 项")
        for f in qf.get("files", []):
            icon = "📁" if f.get("is_dir") else "📄"
            print(f"     {icon} {f.get('name')} (is_dir={f.get('is_dir')}, size={f.get('size')})")
    except:
        print(f"  ❌ Quark 目录解析失败")
    
    # 4.4 列出 Mobile 根目录文件
    print("\n  [4.4] 列出 Mobile 网盘根目录...")
    cmd = f"curl -s 'http://localhost:18900/api/files/mobile?path=%2F' -H 'X-Session-ID: {token}'"
    stdin, stdout, stderr = client.exec_command(cmd)
    mobile_files = stdout.read().decode("utf-8", errors="replace")
    print(f"  Mobile 文件: {mobile_files}")
    
    try:
        mf = json.loads(mobile_files)
        print(f"  ✅ Mobile 目录正常，共 {len(mf.get('files', []))} 项")
        for f in mf.get("files", []):
            icon = "📁" if f.get("is_dir") else "📄"
            print(f"     {icon} {f.get('name')} (is_dir={f.get('is_dir')}, size={f.get('size')})")
    except:
        print(f"  ❌ Mobile 目录解析失败")
    
    # 4.5 进入子目录测试
    print("\n  [4.5] 进入子目录测试 (影视01)...")
    cmd = f"curl -s 'http://localhost:18900/api/files/quark?path=%2F%E5%BD%B1%E8%A7%8601' -H 'X-Session-ID: {token}'"
    stdin, stdout, stderr = client.exec_command(cmd)
    sub_files = stdout.read().decode("utf-8", errors="replace")
    print(f"  子目录: {sub_files}")
    
    try:
        sf = json.loads(sub_files)
        print(f"  ✅ 子目录正常，共 {len(sf.get('files', []))} 项")
    except:
        print(f"  ❌ 子目录解析失败")
    
    # 4.6 检查配置是否正确
    print("\n  [4.6] 获取配置信息...")
    cmd = f"curl -s http://localhost:18900/api/settings -H 'X-Session-ID: {token}'"
    stdin, stdout, stderr = client.exec_command(cmd)
    settings = stdout.read().decode("utf-8", errors="replace")
    print(f"  配置: {settings}")
    
except Exception as e:
    print(f"  ❌ 测试异常: {e}")

# 5. 查看容器日志
print("=" * 50)
print("【5】最近容器日志")
stdin, stdout, stderr = client.exec_command("docker logs --tail 20 quark-mobile 2>&1")
logs = stdout.read().decode("utf-8", errors="replace")
print(logs)

client.close()
print("=" * 50)
print("所有测试完成！")
