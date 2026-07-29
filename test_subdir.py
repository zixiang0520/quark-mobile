import paramiko
import json

SERVER_HOST = "36.140.147.210"
SERVER_USER = "root"
SERVER_PASS = "ZXliukai1."

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(SERVER_HOST, username=SERVER_USER, password=SERVER_PASS, timeout=30)

# 登录
login_data = json.dumps({"username": "admin", "password": "admin123"})
cmd = f"""curl -s -X POST http://localhost:18900/api/login \
  -H 'Content-Type: application/json' \
  -d '{login_data}'"""
stdin, stdout, stderr = client.exec_command(cmd)
login_resp = json.loads(stdout.read().decode("utf-8", errors="replace"))
token = login_resp["session_id"]

# 获取当前配置
print("=" * 50)
print("当前配置:")
cmd = f"curl -s http://localhost:18900/api/settings -H 'X-Session-ID: {token}'"
stdin, stdout, stderr = client.exec_command(cmd)
settings = json.loads(stdout.read().decode("utf-8", errors="replace"))
print(f"  Quark mount: {settings['openlist']['mounts']['quark']}")
print(f"  Mobile mount: {settings['openlist']['mounts']['mobile']}")

# 列出 Quark 根目录，然后尝试进入第一个子目录
print("=" * 50)
print("Quark 根目录:")
cmd = f"curl -s 'http://localhost:18900/api/files/quark?path=%2F' -H 'X-Session-ID: {token}'"
stdin, stdout, stderr = client.exec_command(cmd)
quark_files = json.loads(stdout.read().decode("utf-8", errors="replace"))
for f in quark_files["files"]:
    print(f"  {'📁' if f['is_dir'] else '📄'} {f['name']}")

# 尝试进入第一个子目录
if quark_files["files"]:
    first_dir = quark_files["files"][0]["name"]
    print(f"\n尝试进入子目录: {first_dir}")
    
    import urllib.parse
    encoded_path = urllib.parse.quote(f"/{first_dir}")
    cmd = f"curl -s 'http://localhost:18900/api/files/quark?path={encoded_path}' -H 'X-Session-ID: {token}'"
    stdin, stdout, stderr = client.exec_command(cmd)
    sub_result = stdout.read().decode("utf-8", errors="replace")
    print(f"  响应: {sub_result}")
    
    try:
        sub = json.loads(sub_result)
        if "files" in sub:
            print(f"  ✅ 子目录正常，共 {len(sub['files'])} 项")
            for f in sub["files"][:5]:
                print(f"    {'📁' if f['is_dir'] else '📄'} {f['name']}")
        elif "error" in sub:
            print(f"  ❌ 错误: {sub['error']}")
    except:
        print(f"  解析失败")

# 同样测试 Mobile
print("=" * 50)
print("Mobile 根目录:")
cmd = f"curl -s 'http://localhost:18900/api/files/mobile?path=%2F' -H 'X-Session-ID: {token}'"
stdin, stdout, stderr = client.exec_command(cmd)
mobile_files = json.loads(stdout.read().decode("utf-8", errors="replace"))
for f in mobile_files["files"]:
    print(f"  {'📁' if f['is_dir'] else '📄'} {f['name']}")

if mobile_files["files"]:
    first_dir = mobile_files["files"][0]["name"]
    print(f"\n尝试进入子目录: {first_dir}")
    
    encoded_path = urllib.parse.quote(f"/{first_dir}")
    cmd = f"curl -s 'http://localhost:18900/api/files/mobile?path={encoded_path}' -H 'X-Session-ID: {token}'"
    stdin, stdout, stderr = client.exec_command(cmd)
    sub_result = stdout.read().decode("utf-8", errors="replace")
    print(f"  响应: {sub_result}")
    
    try:
        sub = json.loads(sub_result)
        if "files" in sub:
            print(f"  ✅ 子目录正常，共 {len(sub['files'])} 项")
            for f in sub["files"][:5]:
                print(f"    {'📁' if f['is_dir'] else '📄'} {f['name']}")
        elif "error" in sub:
            print(f"  ❌ 错误: {sub['error']}")
    except:
        print(f"  解析失败")

client.close()
print("\n测试完成!")
