import paramiko
import json

SERVER_HOST = "36.140.147.210"
SERVER_USER = "root"
SERVER_PASS = "ZXliukai1."

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(SERVER_HOST, username=SERVER_USER, password=SERVER_PASS, timeout=30)

# 清理临时文件
stdin, stdout, stderr = client.exec_command("rm -rf /tmp/web_new")
stdout.read()

# 验证完整功能
print("=" * 50)
print("完整功能验证")

# 登录
login_data = json.dumps({"username": "admin", "password": "admin123"})
cmd = f"""curl -s -X POST http://localhost:18900/api/login \
  -H 'Content-Type: application/json' \
  -d '{login_data}'"""
stdin, stdout, stderr = client.exec_command(cmd)
login_resp = json.loads(stdout.read().decode("utf-8", errors="replace"))
token = login_resp["session_id"]

# 测试文件列表
print("文件列表测试:")
cmd = f"curl -s 'http://localhost:18900/api/files/quark?path=%2F' -H 'X-Session-ID: {token}'"
stdin, stdout, stderr = client.exec_command(cmd)
result = json.loads(stdout.read().decode("utf-8", errors="replace"))
print(f"  Quark 根目录: {len(result.get('files', []))} 项")

cmd = f"curl -s 'http://localhost:18900/api/files/mobile?path=%2F' -H 'X-Session-ID: {token}'"
stdin, stdout, stderr = client.exec_command(cmd)
result = json.loads(stdout.read().decode("utf-8", errors="replace"))
print(f"  Mobile 根目录: {len(result.get('files', []))} 项")

# 测试健康状态
print("健康检查:")
stdin, stdout, stderr = client.exec_command("curl -s http://localhost:18900/api/health")
print(f"  {stdout.read().decode('utf-8', errors='replace')}")

# 容器状态
print("容器状态:")
stdin, stdout, stderr = client.exec_command("docker ps --filter name=quark-mobile --format '{{.Status}}'")
print(f"  {stdout.read().decode('utf-8', errors='replace').strip()}")

client.close()
print("=" * 50)
print("✅ 部署完成！")
