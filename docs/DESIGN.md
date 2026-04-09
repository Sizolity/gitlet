# Gitlet 设计文档

## 1. 项目结构

```
gitlet/
├── main.go                    # CLI 入口，命令路由
├── go.mod                     # Go module 定义 (go 1.22)
├── config/
│   └── config.go              # .gitlet 内部路径常量
├── gitlet/
│   ├── blob.go                # Blob 对象：文件内容的持久化单元
│   ├── tree.go                # Tree 对象：目录结构的递归表示
│   ├── commit.go              # Commit 对象：提交的持久化与查询
│   ├── stage.go               # Index（暂存区）：下一次提交的完整快照
│   ├── refs.go                # HEAD 与分支引用的读写（含 detached HEAD）
│   └── ignore.go              # .gitletignore 模式加载与匹配
├── instruction/
│   └── instruction.go         # 所有用户命令的实现逻辑
├── utils/
│   ├── io.go                  # 文件系统操作工具函数（自动创建父目录）
│   ├── utils.go               # SHA-1 哈希生成、路径规范化
│   ├── diff.go                # 基于 LCS 的行级 diff 与格式化输出
│   └── color.go               # 终端 ANSI 颜色输出
└── docs/
    ├── README.md              # 命令参考手册
    └── DESIGN.md              # 本文件
```

### 各包职责

| 包 | 职责 |
|---|------|
| `main` | 解析 `os.Args`，根据子命令分发到 `instruction` 包 |
| `config` | 定义 `.gitlet` 目录结构的路径常量 |
| `gitlet` | 核心数据模型：`Blob`、`Tree`、`Commit`、`Index`、引用操作、`.gitletignore` 匹配 |
| `instruction` | 命令实现层，协调 `gitlet` 包完成每条命令的业务逻辑 |
| `utils` | 与 Git 语义无关的底层工具：文件 I/O、哈希、行级 diff、终端颜色 |

## 2. 数据模型

### 2.1 Blob（文件对象）

```go
type Blob struct {
    Filename string   // 文件名（不含路径）
    FilePath string   // 相对路径（作为 Index 中的 key）
    Contents []byte   // 文件完整内容
    HashId   string   // SHA-1(Contents)，内容寻址
}
```

- 存储位置：`.gitlet/objects/blobs/<HashId>`
- 序列化格式：JSON
- **内容寻址**：相同内容的文件始终产生相同的 HashId，天然去重

### 2.2 Tree（目录对象）

```go
type TreeEntry struct {
    Name   string   // 单层名称（如 "main.go" 或 "src"）
    Type   string   // "blob" 或 "tree"
    HashId string   // 指向 Blob 或子 Tree 的 HashId
}

type Tree struct {
    Entries []TreeEntry   // 按 Name 排序的条目列表
    HashId  string        // SHA-1(JSON(排序后的 Entries))
}
```

- 存储位置：`.gitlet/objects/trees/<HashId>`
- 序列化格式：JSON
- 每个 Tree 表示一个目录层级，通过 `type: "tree"` 的条目递归引用子目录
- **内容寻址**：相同目录结构产生相同 HashId，天然去重

示例：对于文件 `foo.txt` 和 `src/main.go`，会生成：

```
Root Tree:
  {Name: "foo.txt", Type: "blob", HashId: "aaa..."}
  {Name: "src",     Type: "tree", HashId: "bbb..."}

src/ Tree (HashId: "bbb..."):
  {Name: "main.go", Type: "blob", HashId: "ccc..."}
```

### 2.3 Commit（提交对象）

```go
type Commit struct {
    Message  string      // 提交信息
    Parent   []string    // 父提交 HashId 列表（合并提交有两个父提交）
    CurrDate time.Time
    HashId   string      // SHA-1(Message + Time + Parents)
    TreeId   string      // 根 Tree 的 HashId
}
```

- 存储位置：`.gitlet/objects/commits/<HashId>`
- 序列化格式：JSON
- `TreeId` 指向根 Tree 对象，递归表示提交时的完整目录快照
- 加载时自动通过 `FlattenTree(TreeId)` 展平为 `BlobIds map[string]string` 供内部使用

### 2.4 Index（暂存区）

```go
type Index struct {
    Entries map[string]string   // filepath -> blob HashId
}
```

- 存储位置：`.gitlet/index`（单个 JSON 文件）
- **核心语义**：Index 始终表示"下一次提交的完整文件快照"
- Index 使用扁平的 `filepath -> blobId` 格式，Tree 仅在 commit 时构建
- `init` / `commit` / `checkout` / `reset` 之后，Index 与 HEAD 提交展平后的内容一致
- `add` 更新对应条目，`rm` 删除对应条目

## 3. 存储结构

`.gitlet` 目录的完整布局：

```
.gitlet/
├── HEAD                       # 分支引用路径或直接 commitId（detached HEAD）
├── index                      # 暂存区 JSON {"entries": {"foo.txt": "abc123..."}}
├── objects/
│   ├── blobs/                 # Blob 对象，文件名 = SHA-1(内容)
│   │   ├── a1b2c3d4...
│   │   └── e5f6a7b8...
│   ├── trees/                 # Tree 对象，文件名 = SHA-1(条目列表)
│   │   ├── f1e2d3c4...
│   │   └── b5a6c7d8...
│   └── commits/               # Commit 对象，文件名 = HashId
│       ├── 1a2b3c4d...
│       └── 5e6f7a8b...
└── refs/
    ├── heads/                 # 本地分支，每个文件内容是 commitId
    │   ├── master
    │   └── dev
    └── remotes/               # 远程分支（预留）
```

### 引用解析流程

```
HEAD -> commitId -> Commit 对象
  Commit.TreeId -> Root Tree 对象
    Tree.Entries -> Blob 对象（文件）
                 -> 子 Tree 对象（子目录）-> ...递归
```

## 4. 核心架构

### 4.1 三区模型

Gitlet 遵循 Git 的三区设计：

```
┌─────────────┐     add      ┌─────────────┐    commit    ┌─────────────┐
│  Working    │ ──────────→  │   Index      │ ──────────→ │  Commit     │
│  Tree       │              │  (暂存区)     │             │  (对象库)    │
│  工作区文件   │ ←────────── │  完整快照     │ ←────────── │  历史快照    │
└─────────────┘   checkout   └─────────────┘   checkout   └─────────────┘
                              ↑                  reset
                              └──────────────────────────────────┘
```

提交时，Index 的扁平 map 通过 `BuildTree` 构建为 Tree 层级结构；加载时，通过 `FlattenTree` 展平回 map。

三区之间的比较关系决定了 `status` 和 `diff` 的输出：

| 比较 | 结果 |
|------|------|
| Index vs HEAD Commit | Staged Files / Removed Files（`diff --staged`） |
| Working Tree vs Index | Modifications Not Staged / Untracked Files（`diff`） |

### 4.2 命令执行流程

#### `init`

```
1. 创建 .gitlet 目录结构（含 objects/trees/）
2. 创建空 Tree 对象并持久化
3. 创建初始 Commit（TreeId 指向空 Tree）
4. 写入 refs/heads/master = commitId
5. 写入 HEAD = ".gitlet/refs/heads/master"
6. 创建空 Index
```

#### `add <file>`

```
1. 规范化路径（NormalizePath）
2. 检查 .gitletignore，若匹配则拒绝添加
3. 读取工作区文件内容，计算 SHA-1 哈希 → blobId
4. 加载 Index 和 HEAD Commit
5. 如果 blobId == HEAD 中该文件的 blobId:
     → 将 Index 条目还原为 HEAD 的值（等效于 unstage）
6. 否则:
     → 将 Blob 持久化到 objects/blobs/
     → 更新 Index 中该文件的 blobId
7. 保存 Index
```

#### `commit <message>`

```
1. 加载 Index 和 HEAD Commit
2. 比较 Index.Entries 与 HEAD 展平后的 BlobIds
     → 完全相同则拒绝提交
3. 调用 BuildTree(Index.Entries) 构建 Tree 层级 → treeId
4. 创建新 Commit，TreeId = treeId
5. 持久化 Commit 到 objects/commits/
6. 更新当前分支指针（或 detached HEAD）→ 新 commitId
   （Index 保持不变，此时 Index == 新 HEAD 展平后的内容）
```

#### `rm <file>`

```
1. 加载 Index 和 HEAD Commit
2. 判断文件状态:
   a. 已暂存但 HEAD 未跟踪（新文件）→ 从 Index 删除条目
   b. HEAD 已跟踪 → 从 Index 删除条目 + 删除工作区文件
   c. 都不满足 → 报错
3. 保存 Index
```

#### `status`

```
1. 显示分支信息:
   - 正常模式: *当前分支名
   - Detached HEAD: *HEAD detached at <短哈希>
2. 遍历 Index vs HEAD:
   - Index 有但 HEAD 没有，或 blobId 不同 → Staged
   - HEAD 有但 Index 没有 → Removed
3. 遍历 Index vs WorkTree（尊重 .gitletignore，递归子目录）:
   - Index 有但工作区文件 SHA-1 不同 → Modified (not staged)
   - Index 有但工作区文件不存在 → Deleted (not staged)
4. 遍历工作区 vs Index:
   - 工作区有但 Index 没有 → Untracked
```

#### `checkout`

```
checkout - <file>          从 HEAD 恢复单个文件到工作区
checkout <commit> - <file> 从指定提交恢复单个文件到工作区
checkout <branch>          切换分支:
  1. 验证目标分支存在
  2. 删除当前 HEAD 提交对应的工作区文件
  3. 切换 HEAD 指向新分支
  4. 从新分支 HEAD 提交恢复工作区文件
  5. 同步 Index = 新提交的 BlobIds
checkout <commitId>        Detached HEAD（参数非分支名时）:
  1. 删除当前工作区文件
  2. 将 HEAD 直接写为 commitId（脱离分支）
  3. 从目标提交恢复工作区文件，同步 Index
```

#### `reset <commitId>`

```
1. 删除当前 HEAD 提交对应的工作区文件
2. 移动当前分支指针（或 detached HEAD）到目标 commitId
3. 从目标提交恢复工作区文件
4. 同步 Index = 目标提交的 BlobIds
```

#### `merge <branch>`

```
1. 前置检查:
   - 不能在 detached HEAD 状态下合并
   - 目标分支存在
   - 不能与自身合并
   - 暂存区无未提交的修改（Index == HEAD.BlobIds）
2. BFS 求分裂点（Split Point）:
   - 收集当前分支所有祖先 commit
   - 从目标分支 BFS，首个命中祖先集合的 commit 即为分裂点
3. 快进合并（Fast-forward）:
   - 若分裂点 == 目标分支 HEAD → 已是最新，无需操作
   - 若分裂点 == 当前 HEAD → 直接将当前分支指针移至目标分支 HEAD
4. 三路合并（Three-way merge）:
   - 遍历 split/current/target 三个提交的所有文件
   - 对每个文件按修改情况决定保留哪一版:
     · 仅一侧修改 → 采用修改方
     · 两侧相同修改 → 采用任一
     · 两侧不同修改 → 写入冲突标记（<<<<<<< / ======= / >>>>>>>）
   - 清理旧工作区文件，写入合并后文件
   - 调用 BuildTree 从合并结果构建 Tree
   - 创建 merge commit（两个父提交，TreeId 指向合并后的 Tree）
```

#### `diff`

```
diff           工作区 vs 暂存区:
  1. 遍历 Index 中所有文件
  2. 对比工作区文件内容的 SHA-1 与 Index 记录的 blobId
  3. 若不同，使用 LCS 算法计算行级差异
  4. 以 unified-diff 风格 + ANSI 颜色输出

diff --staged  暂存区 vs HEAD:
  1. 遍历 Index，找出 blobId 与 HEAD 不同的文件（staged 修改/新增）
  2. 遍历 HEAD，找出 Index 中不存在的文件（staged 删除）
  3. 使用 LCS 算法计算行级差异
  4. 以 unified-diff 风格 + ANSI 颜色输出
```

## 5. 内容寻址与 SHA-1

### ID 生成

```go
func GenerateID(data []byte) string {
    hasher := sha1.New()
    hasher.Write(data)
    return fmt.Sprintf("%x", hasher.Sum(nil))
}
```

| 对象 | 哈希输入 | 含义 |
|------|----------|------|
| Blob | 文件原始字节 | 相同内容 = 相同 HashId，天然去重 |
| Tree | 排序后的 Entries JSON | 相同目录结构 = 相同 HashId |
| Commit | message + time + parentIds | 确保每个提交有唯一标识 |

## 6. 包依赖关系

```
main
 └── instruction
      ├── gitlet
      │    ├── config
      │    └── utils
      ├── config
      └── utils
           （仅依赖标准库）
```

所有包单向依赖，无循环引用。`gitlet` 包提供数据模型，`instruction` 包编排业务流程，`utils` 包提供底层 I/O 和 diff。

## 7. 与真实 Git 的对比

| 特性 | Git | Gitlet |
|------|-----|--------|
| 对象存储 | 二进制 packfile + loose objects | JSON 文件 |
| 对象类型 | blob, tree, commit, tag | blob, tree, commit |
| 暂存区 | 二进制 index 文件 | JSON index 文件（扁平 map） |
| 内容寻址 | SHA-1（带类型前缀 `blob <size>\0`） | SHA-1（纯内容） |
| 分支 | refs/heads/ 下的文本文件 | 相同 |
| HEAD | 支持 detached HEAD | 支持 detached HEAD |
| Merge | 三路合并 + 冲突处理 | 三路合并 + 冲突标记 |
| Diff | 多种 diff 算法（Myers、patience 等） | 基于 LCS 的行级 diff |
| 忽略文件 | `.gitignore`（支持嵌套、否定模式） | `.gitletignore`（顶层 glob 模式） |
| 目录跟踪 | tree 对象递归表示 | tree 对象递归表示 |
| 网络协议 | push/pull/fetch | 未实现 |

## 8. 已知限制

1. **路径一致性** — `NormalizePath` 已处理 `./` 前缀，但更复杂的相对路径（如 `../`）未处理
2. **无网络功能** — 不支持 push / pull / fetch / remote