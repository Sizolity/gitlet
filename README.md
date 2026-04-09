# Gitlet

用 Go 语言实现的简易版 Git，灵感来源于 UC Berkeley CS61B 课程项目 **gitlet**。

Gitlet 复现了 Git 的核心工作流：暂存、提交、分支、合并、差异比较，采用 SHA-1 内容寻址和 JSON 序列化存储对象，便于学习和理解版本控制系统的底层原理。

## 功能

| 命令 | 说明 |
|------|------|
| `init` | 初始化 `.gitlet` 仓库 |
| `add <file>` | 将文件添加到暂存区（支持子目录文件） |
| `commit <message>` | 提交暂存区快照 |
| `rm <file>` | 取消暂存或删除已跟踪文件 |
| `log` | 显示当前分支的提交历史（沿第一父提交链） |
| `global-log` | 列出所有提交的短哈希 |
| `find <message>` | 按提交信息查找提交 |
| `status` | 显示分支、暂存、修改、未跟踪文件状态 |
| `checkout` | 恢复文件、切换分支或进入 detached HEAD |
| `branch <name>` | 在当前 HEAD 创建新分支 |
| `rm-branch <name>` | 删除分支 |
| `reset <commitId>` | 将当前分支重置到指定提交 |
| `merge <branch>` | 三路合并（支持快进和冲突标记） |
| `diff [--staged]` | 显示行级差异（工作区 vs 暂存区，或暂存区 vs HEAD） |

## 快速开始

```bash
# 构建
go build -o gitlet .

# 初始化仓库
./gitlet init

# 基本工作流
./gitlet add hello.txt
./gitlet commit "add hello.txt"
./gitlet log
```

## 用法示例

### 子目录文件

```bash
./gitlet add src/main.go
./gitlet add docs/README.md
./gitlet commit "add project files"
```

### 分支与合并

```bash
./gitlet branch feature
./gitlet checkout feature

# 在 feature 分支上修改并提交
./gitlet add hello.txt
./gitlet commit "update hello on feature"

# 切回 master 并合并
./gitlet checkout master
./gitlet merge feature
```

### Detached HEAD

```bash
# 直接 checkout 到某个 commitId（脱离分支）
./gitlet checkout a1b2c3d

# 在 detached HEAD 下创建新分支
./gitlet branch hotfix
./gitlet checkout hotfix
```

### 查看差异

```bash
# 工作区 vs 暂存区
./gitlet diff

# 暂存区 vs HEAD
./gitlet diff --staged
```

### 忽略文件

在项目根目录创建 `.gitletignore` 文件，每行一个 glob 模式：

```
*.exe
*.o
temp_*
```

## 项目结构

```
config/       路径常量
gitlet/       核心数据模型（Blob、Commit、Index、Refs、Ignore）
instruction/  命令实现
utils/        文件 I/O、SHA-1 哈希、行级 diff、ANSI 颜色
docs/         设计文档与命令参考
```

详细架构设计见 [docs/DESIGN.md](docs/DESIGN.md)，命令参考见 [docs/README.md](docs/README.md)。

## 技术细节

- **语言**: Go 1.22，仅使用标准库，零外部依赖
- **对象存储**: JSON 序列化，SHA-1 内容寻址
- **Diff 算法**: 基于 LCS（最长公共子序列）的行级比较
- **Merge**: BFS 求分裂点 + 三路合并 + 冲突标记
- **目录支持**: 递归遍历子目录，自动创建/清理父目录
- **Detached HEAD**: HEAD 可直接指向 commitId，脱离分支

## 参考

- [CS61B Gitlet Spec](https://sp21.datastructur.es/materials/proj/proj2/proj2)
- [实现参考博客](https://zhuanlan.zhihu.com/p/533852291)