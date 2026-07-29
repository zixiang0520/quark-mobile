#!/usr/bin/env python3
"""
部署脚本：将 quark-mobile 项目上传到服务器并部署
"""
import os
import sys
import tarfile
import tempfile
import paramiko

# 服务器配置
SERVER_HOST = "36.140.147.210"
SERVER_USER = "root"
SERVER_PASS = "ZXliukai1."
REMOTE_DIR = "/opt/quark-mobile"

# 本地项目目录
LOCAL_DIR = r"d:\quark-mobile"

# 排除的目录和文件
EXCLUDE_PATTERNS = [
    ".git", "node_modules", "dist", "data", "__pycache__",
    "*.exe", ".env", ".env.*", "*.swp",
    "Thumbs.db", ".DS_Store", ".vscode", ".idea",
    "server.exe", "web/package-lock.json"
]


def create_deployment_archive():
    """创建部署压缩包"""
    archive_path = os.path.join(tempfile.gettempdir(), "quark-mobile-deploy.tar.gz")
    
    with tarfile.open(archive_path, "w:gz") as tar:
        for root, dirs, files in os.walk(LOCAL_DIR):
            # 排除目录
            dirs[:] = [d for d in dirs if d not in EXCLUDE_PATTERNS]
            
            for file in files:
                # 排除文件
                skip = False
                for pattern in EXCLUDE_PATTERNS:
                    if pattern.startswith("*."):
                        if file.endswith(pattern[1:]):
                            skip = True
                            break
                    elif file == pattern:
                        skip = True
                        break
                if skip:
                    continue
                
                file_path = os.path.join(root, file)
                arcname = os.path.relpath(file_path, LOCAL_DIR).replace(os.sep, '/')
                tar.add(file_path, arcname=arcname)
    
    print(f"[OK] 部署包已创建: {archive_path}")
    return archive_path


def deploy():
    """部署到服务器"""
    print(f"[*] 正在连接服务器 {SERVER_HOST}...")
    
    # 创建 SSH 客户端
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    
    try:
        client.connect(SERVER_HOST, username=SERVER_USER, password=SERVER_PASS, timeout=30)
        print(f"[OK] 已连接到服务器")
        
        # 创建远程目录
        client.exec_command(f"mkdir -p {REMOTE_DIR}")
        
        # 上传部署包
        archive_path = create_deployment_archive()
        remote_archive = "/tmp/quark-mobile-deploy.tar.gz"
        
        print(f"[*] 正在上传文件...")
        sftp = client.open_sftp()
        sftp.put(archive_path, remote_archive)
        sftp.close()
        print(f"[OK] 文件已上传")
        
        # 解压并部署
        print(f"[*] 正在部署...")
        commands = [
            f"docker rm -f quark-mobile 2>/dev/null || true",
            f"cd {REMOTE_DIR} && rm -rf *",
            f"cd {REMOTE_DIR} && tar xzf {remote_archive}",
            f"rm -f {remote_archive}",
            f"cd {REMOTE_DIR} && docker compose build --no-cache",
            f"cd {REMOTE_DIR} && docker compose up -d",
            f"sleep 3 && docker ps | grep quark-mobile",
        ]
        
        for cmd in commands:
            print(f"[CMD] {cmd}")
            stdin, stdout, stderr = client.exec_command(cmd, timeout=300)
            out = stdout.read().decode("utf-8", errors="replace")
            err = stderr.read().decode("utf-8", errors="replace")
            if out.strip():
                print(out.strip())
            if err.strip():
                print(f"[WARN] {err.strip()}")
        
        # 验证
        print(f"\n[*] 验证部署...")
        stdin, stdout, stderr = client.exec_command(f"curl -s http://localhost:18900/api/health")
        health = stdout.read().decode("utf-8", errors="replace")
        print(f"健康检查: {health}")
        
        print(f"\n[OK] 部署完成！访问 http://{SERVER_HOST}:18900")
        
    except Exception as e:
        print(f"[ERROR] 部署失败: {e}")
        sys.exit(1)
    finally:
        client.close()


if __name__ == "__main__":
    deploy()
