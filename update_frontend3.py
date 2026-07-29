import paramiko
import os

SERVER_HOST = "36.140.147.210"
SERVER_USER = "root"
SERVER_PASS = "ZXliukai1."

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(SERVER_HOST, username=SERVER_USER, password=SERVER_PASS, timeout=30)

def run_cmd(cmd):
    stdin, stdout, stderr = client.exec_command(cmd)
    exit_code = stdout.channel.recv_exit_status()
    out = stdout.read().decode("utf-8", errors="replace")
    err = stderr.read().decode("utf-8", errors="replace")
    return out, err, exit_code

# 直接复制文件到容器内部的 /app/web/dist
local_dist = r"d:\quark-mobile\web\dist"

# 创建临时目录存放新文件
run_cmd("rm -rf /tmp/web_new && mkdir -p /tmp/web_new/assets")

sftp = client.open_sftp()
sftp.put(os.path.join(local_dist, "index.html"), "/tmp/web_new/index.html")
sftp.put(os.path.join(local_dist, "assets", "index-D8kej4QG.js"), "/tmp/web_new/assets/index-D8kej4QG.js")
sftp.put(os.path.join(local_dist, "assets", "index-tTa6sWcR.css"), "/tmp/web_new/assets/index-tTa6sWcR.css")
sftp.close()

# 从宿主机复制到容器内
print("复制文件到容器...")
run_cmd("docker cp /tmp/web_new/. quark-mobile:/app/web/dist/")
run_cmd("docker cp /tmp/web_new/assets/. quark-mobile:/app/web/dist/assets/")

# 清理旧文件
print("清理容器内旧的资源文件...")
run_cmd("docker exec quark-mobile rm -f /app/web/dist/assets/index-CeefqwZ2.js")
run_cmd("docker exec quark-mobile rm -f /app/web/dist/assets/index-CqiFqdZg.js")
run_cmd("docker exec quark-mobile rm -f /app/web/dist/assets/index-I4wjT48d.css")

# 验证
print("\n验证容器内文件:")
out, _, _ = run_cmd("docker exec quark-mobile ls -la /app/web/dist/")
print(out)
out, _, _ = run_cmd("docker exec quark-mobile ls -la /app/web/dist/assets/")
print(out)

# HTTP 验证
print("HTTP 验证:")
out, _, _ = run_cmd("curl -s http://localhost:18900/ | head -15")
print(out)

# 验证新 JS 文件可访问
print("新 JS 文件验证:")
out, _, _ = run_cmd("curl -s -o /dev/null -w '%{http_code}' http://localhost:18900/assets/index-D8kej4QG.js")
print(f"HTTP {out}")

client.close()
print("\n✅ 部署完成！")
