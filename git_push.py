import subprocess
import sys

# 执行 git 提交和推送
commands = [
    "git add -A",
    'git commit -m "feat: 配置页面添加返回按钮"',
    "git push origin main",
]

for cmd in commands:
    print(f"> {cmd}")
    result = subprocess.run(cmd, shell=True, cwd=r"d:\quark-mobile", capture_output=True, text=True)
    print(result.stdout)
    if result.stderr:
        print(result.stderr, file=sys.stderr)
    if result.returncode != 0 and "nothing to commit" not in result.stdout.lower():
        print(f"Error: {result.stderr}")

print("\n✅ Git 操作完成")
