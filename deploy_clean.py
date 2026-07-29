import paramiko

SERVER_HOST = "36.140.147.210"
SERVER_USER = "root"
SERVER_PASS = "ZXliukai1."

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(SERVER_HOST, username=SERVER_USER, password=SERVER_PASS, timeout=30)

def run_cmd(cmd):
    stdin, stdout, stderr = client.exec_command(cmd)
    stdout.channel.recv_exit_status()
    return stdout.read().decode("utf-8", errors="replace")

# 彻底清理容器内的旧文件
print("清理容器内所有旧前端文件...")
run_cmd("docker exec quark-mobile rm -rf /app/web/dist")
run_cmd("docker exec quark-mobile mkdir -p /app/web/dist/assets")

# 用临时目录重新上传
import os
sftp = client.open_sftp()
local_dist = r"d:\quark-mobile\web\dist"
sftp.put(os.path.join(local_dist, "index.html"), "/tmp/web_new/index.html")
sftp.put(os.path.join(local_dist, "assets", "index-Da1cdxsL.js"), "/tmp/web_new/assets/index-Da1cdxsL.js")
sftp.put(os.path.join(local_dist, "assets", "index-fZp6Y3F6.css"), "/tmp/web_new/assets/index-fZp6Y3F6.css")
sftp.close()

run_cmd("docker cp /tmp/web_new/index.html quark-mobile:/app/web/dist/index.html")
run_cmd("docker cp /tmp/web_new/assets/index-Da1cdxsL.js quark-mobile:/app/web/dist/assets/index-Da1cdxsL.js")
run_cmd("docker cp /tmp/web_new/assets/index-fZp6Y3F6.css quark-mobile:/app/web/dist/assets/index-fZp6Y3F6.css")

# 验证
print("容器内文件列表:")
print(run_cmd("docker exec quark-mobile find /app/web/dist/ -type f"))

print("\nHTTP 验证:")
print(run_cmd("curl -s http://localhost:18900/"))

client.close()
print("\n✅ 部署完成！")
