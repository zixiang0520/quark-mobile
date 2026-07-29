import paramiko

SERVER_HOST = "36.140.147.210"
SERVER_USER = "root"
SERVER_PASS = "ZXliukai1."
REMOTE_DIR = "/opt/quark-mobile"

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(SERVER_HOST, username=SERVER_USER, password=SERVER_PASS, timeout=30)

# 只上传前端构建产物
local_dist = r"d:\quark-mobile\web\dist"
remote_dist = f"{REMOTE_DIR}/web/dist"

# 清理旧的 dist 文件
client.exec_command(f"rm -rf {remote_dist}")
client.exec_command(f"mkdir -p {remote_dist}")

# 上传文件
sftp = client.open_sftp()
import os

for root, dirs, files in os.walk(local_dist):
    rel_path = os.path.relpath(root, local_dist)
    remote_dir_path = f"{remote_dist}/{rel_path}" if rel_path != "." else remote_dist
    client.exec_command(f"mkdir -p {remote_dir_path}")
    
    for file in files:
        local_file = os.path.join(root, file)
        remote_file = f"{remote_dir_path}/{file}"
        print(f"Uploading: {rel_path}/{file}")
        sftp.put(local_file, remote_file)

# 也更新 index.html
local_index = r"d:\quark-mobile\web\dist\index.html"
sftp.put(local_index, f"{REMOTE_DIR}/web/dist/index.html")

sftp.close()

# 重启容器使新前端生效
print("\n重启容器...")
stdin, stdout, stderr = client.exec_command(f"cd {REMOTE_DIR} && docker compose up -d")
print(stdout.read().decode("utf-8", errors="replace"))

import time
time.sleep(3)

# 验证
print("\n验证...")
stdin, stdout, stderr = client.exec_command("curl -s http://localhost:18900/ | head -15")
print(stdout.read().decode("utf-8", errors="replace"))

# 检查容器状态
stdin, stdout, stderr = client.exec_command("docker ps | grep quark-mobile")
print(stdout.read().decode("utf-8", errors="replace"))

client.close()
print("\n✅ 部署完成！")
