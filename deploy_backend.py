import paramiko
import os
import subprocess

# 先在本地编译 Go 后端
print("编译 Go 后端...")
result = subprocess.run(
    ["go", "build", "-o", "server", "./cmd/server"],
    cwd=r"d:\quark-mobile",
    capture_output=True,
    text=True,
    env={**os.environ, "CGO_ENABLED": "0", "GOOS": "linux"}
)
if result.returncode != 0:
    print(f"编译失败: {result.stderr}")
    exit(1)
print("✅ 编译成功")

# 部署到服务器
SERVER_HOST = "36.140.147.210"
SERVER_USER = "root"
SERVER_PASS = "ZXliukai1."
REMOTE_DIR = "/opt/quark-mobile"

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(SERVER_HOST, username=SERVER_USER, password=SERVER_PASS, timeout=30)

def run_cmd(cmd):
    stdin, stdout, stderr = client.exec_command(cmd)
    stdout.channel.recv_exit_status()
    return stdout.read().decode("utf-8", errors="replace")

# 上传编译好的 server 二进制
print("上传 server 二进制...")
sftp = client.open_sftp()
sftp.put(r"d:\quark-mobile\server", f"{REMOTE_DIR}/server")
sftp.close()
run_cmd(f"chmod +x {REMOTE_DIR}/server")

# 停止容器，更新容器内的二进制，然后重启
print("重启容器...")
run_cmd("docker stop quark-mobile")
run_cmd("docker cp /opt/quark-mobile/server quark-mobile:/app/server")
run_cmd("docker start quark-mobile")

import time
time.sleep(3)

# 验证
print("\n容器状态:")
print(run_cmd("docker ps --filter name=quark-mobile --format '{{.Status}}'"))

print("\n健康检查:")
print(run_cmd("curl -s http://localhost:18900/api/health"))

print("\n最近日志:")
print(run_cmd("docker logs --tail 10 quark-mobile 2>&1"))

client.close()
print("\n✅ 部署完成！")
