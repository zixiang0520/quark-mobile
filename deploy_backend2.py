import paramiko
import os

# 使用已编译的 server 文件
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

# 检查本地 server 文件是否存在
local_server = r"d:\quark-mobile\server"
if not os.path.exists(local_server):
    print("错误: server 文件不存在！请先编译。")
    exit(1)

print("上传 server 二进制到服务器...")
sftp = client.open_sftp()
sftp.put(local_server, f"{REMOTE_DIR}/server")
sftp.close()
run_cmd(f"chmod +x {REMOTE_DIR}/server")

# 停止容器，更新容器内的二进制，然后重启
print("重启容器...")
print(run_cmd("docker stop quark-mobile"))
print(run_cmd("docker cp /opt/quark-mobile/server quark-mobile:/app/server"))
print(run_cmd("docker start quark-mobile"))

import time
time.sleep(5)

# 验证
print("\n容器状态:")
print(run_cmd("docker ps --filter name=quark-mobile --format '{{.Status}}'"))

print("\n健康检查:")
print(run_cmd("curl -s http://localhost:18900/api/health"))

print("\n最近日志:")
print(run_cmd("docker logs --tail 15 quark-mobile 2>&1"))

client.close()
print("\n✅ 部署完成！")
