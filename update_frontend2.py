import paramiko
import os

SERVER_HOST = "36.140.147.210"
SERVER_USER = "root"
SERVER_PASS = "ZXliukai1."
REMOTE_DIR = "/opt/quark-mobile"

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(SERVER_HOST, username=SERVER_USER, password=SERVER_PASS, timeout=30)

# 等待命令完成的辅助函数
def run_cmd(cmd):
    stdin, stdout, stderr = client.exec_command(cmd)
    exit_code = stdout.channel.recv_exit_status()
    return stdout.read().decode("utf-8", errors="replace"), stderr.read().decode("utf-8", errors="replace"), exit_code

# 先清空远程 dist 目录
print("清空远程 dist 目录...")
run_cmd(f"rm -rf {REMOTE_DIR}/web/dist")
run_cmd(f"mkdir -p {REMOTE_DIR}/web/dist/assets")

# 上传新文件（只上传当前 dist 中的文件）
local_dist = r"d:\quark-mobile\web\dist"
sftp = client.open_sftp()

# 上传 index.html
sftp.put(os.path.join(local_dist, "index.html"), f"{REMOTE_DIR}/web/dist/index.html")
print("Uploaded: index.html")

# 只上传新的 assets（index-D8kej4QG.js, index-tTa6sWcR.css）
for f in os.listdir(os.path.join(local_dist, "assets")):
    # 只上传新构建的文件（跳过旧的）
    if f in ["index-D8kej4QG.js", "index-tTa6sWcR.css"]:
        sftp.put(os.path.join(local_dist, "assets", f), f"{REMOTE_DIR}/web/dist/assets/{f}")
        print(f"Uploaded: assets/{f}")

sftp.close()

# 验证上传结果
print("\n验证远程文件:")
out, _, _ = run_cmd(f"ls -la {REMOTE_DIR}/web/dist/")
print(out)
out, _, _ = run_cmd(f"ls -la {REMOTE_DIR}/web/dist/assets/")
print(out)

out, _, _ = run_cmd(f"cat {REMOTE_DIR}/web/dist/index.html")
print("index.html 内容:")
print(out)

# 验证通过 HTTP 访问
print("\nHTTP 验证:")
out, _, _ = run_cmd("curl -s http://localhost:18900/ | head -15")
print(out)

client.close()
print("\n✅ 部署完成！")
