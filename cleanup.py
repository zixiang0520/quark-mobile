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

# 删除所有旧的 assets 文件（保留新的）
print("删除旧文件...")
old_files = [
    "index-CeefqwZ2.js",
    "index-CqiFqdZg.js",
    "index-D8kej4QG.js", 
    "index-I4wjT48d.css",
    "index-tTa6sWcR.css",
]
for f in old_files:
    run_cmd(f"docker exec quark-mobile rm -f /app/web/dist/assets/{f}")

print("容器内文件:")
print(run_cmd("docker exec quark-mobile ls -la /app/web/dist/assets/"))

# 验证
print("\nHTTP 验证:")
print(run_cmd("curl -s http://localhost:18900/ | head -10"))

client.close()
print("\n✅ 清理完成！")
