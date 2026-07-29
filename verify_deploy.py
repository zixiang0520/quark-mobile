import paramiko

SERVER_HOST = "36.140.147.210"
SERVER_USER = "root"
SERVER_PASS = "ZXliukai1."

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(SERVER_HOST, username=SERVER_USER, password=SERVER_PASS, timeout=30)

# 1. 查看容器状态
print("=== 容器状态 ===")
stdin, stdout, stderr = client.exec_command("docker ps -a | grep quark-mobile")
print(stdout.read().decode("utf-8", errors="replace"))

# 2. 健康检查
print("=== 健康检查 ===")
stdin, stdout, stderr = client.exec_command("curl -s http://localhost:18900/api/health")
print(stdout.read().decode("utf-8", errors="replace"))

# 3. 登录获取 token
print("=== 登录测试 ===")
stdin, stdout, stderr = client.exec_command("curl -s -X POST http://localhost:18900/api/login -H 'Content-Type: application/json' -d '{\"username\":\"admin\",\"password\":\"admin123\"}'")
login_result = stdout.read().decode("utf-8", errors="replace")
print(login_result)

# 提取 token
import json
try:
    data = json.loads(login_result)
    token = data.get("session_id", "")
    
    # 4. 获取驱动列表
    print("=== 驱动列表 ===")
    stdin, stdout, stderr = client.exec_command(f"curl -s http://localhost:18900/api/drivers -H 'X-Session-ID: {token}'")
    print(stdout.read().decode("utf-8", errors="replace"))
    
    # 5. 获取文件列表
    print("=== Quark 文件列表 ===")
    stdin, stdout, stderr = client.exec_command(f"curl -s 'http://localhost:18900/api/files/quark?path=%2F' -H 'X-Session-ID: {token}'")
    print(stdout.read().decode("utf-8", errors="replace"))
    
    print("=== Mobile 文件列表 ===")
    stdin, stdout, stderr = client.exec_command(f"curl -s 'http://localhost:18900/api/files/mobile?path=%2F' -H 'X-Session-ID: {token}'")
    print(stdout.read().decode("utf-8", errors="replace"))
    
except Exception as e:
    print(f"Error: {e}")

# 6. 查看容器日志
print("=== 最近日志 ===")
stdin, stdout, stderr = client.exec_command("docker logs --tail 10 quark-mobile")
print(stdout.read().decode("utf-8", errors="replace"))

client.close()
print("\n[OK] 验证完成")
