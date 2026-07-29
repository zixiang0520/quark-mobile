import paramiko
import os

SERVER_HOST = "36.140.147.210"
SERVER_USER = "root"
SERVER_PASS = "ZXliukai1."
REMOTE_DIR = "/opt/quark-mobile"

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(SERVER_HOST, username=SERVER_USER, password=SERVER_PASS, timeout=30)

def run_cmd(cmd):
    stdin, stdout, stderr = client.exec_command(cmd)
    exit_code = stdout.channel.recv_exit_status()
    out = stdout.read().decode("utf-8", errors="replace")
    err = stderr.read().decode("utf-8", errors="replace")
    return out, err, exit_code

# 创建临时目录存放新文件
run_cmd("rm -rf /tmp/web_new && mkdir -p /tmp/web_new/assets")

# 上传新文件
local_dist = r"d:\quark-mobile\web\dist"
sftp = client.open_sftp()

sftp.put(os.path.join(local_dist, "index.html"), "/tmp/web_new/index.html")
sftp.put(os.path.join(local_dist, "assets", "index-Da1cdxsL.js"), "/tmp/web_new/assets/index-Da1cdxsL.js")
sftp.put(os.path.join(local_dist, "assets", "index-fZp6Y3F6.css"), "/tmp/web_new/assets/index-fZp6Y3F6.css")
sftp.close()

# 复制到容器
print("复制文件到容器...")
run_cmd("docker cp /tmp/web_new/. quark-mobile:/app/web/dist/")
run_cmd("docker cp /tmp/web_new/assets/. quark-mobile:/app/web/dist/assets/")

# 清理旧的资源文件
print("清理旧资源...")
for old_file in ["index-CeefqwZ2.js", "index-CqiFqdZg.js", "index-I4wjT48d.css", "index-D8kej4QG.js", "index-tTa6sWcR.css"]:
    run_cmd(f"docker exec quark-mobile rm -f /app/web/dist/assets/{old_file}")

# 验证
print("\n验证容器内文件:")
out, _, _ = run_cmd("docker exec quark-mobile ls /app/web/dist/assets/")
print(out)

# HTTP 验证
print("\nHTTP 验证:")
out, _, _ = run_cmd("curl -s http://localhost:18900/ | head -15")
print(out)

# 验证新 JS
out, _, _ = run_cmd("curl -s -o /dev/null -w '%{http_code}' http://localhost:18900/assets/index-Da1cdxsL.js")
print(f"JS 文件 HTTP: {out}")

out, _, _ = run_cmd("curl -s -o /dev/null -w '%{http_code}' http://localhost:18900/assets/index-fZp6Y3F6.css")
print(f"CSS 文件 HTTP: {out}")

client.close()
print("\n✅ 部署完成！")
