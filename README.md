# Obsidian Config Sync

Obsidian Config Sync 是一个桌面工具，用于将一个 Obsidian 主库的 `.obsidian` 配置同步到多个目标库。

本项目基于 Wails 构建，理论上支持 Windows、macOS 和 Linux 桌面平台。目前作者仅在 Windows 上完成自测，macOS 和 Linux 尚未验证，无法保证正常运行。

同步过程完全在本地运行。你可以选择需要同步的配置文件、主题、CSS 片段和社区插件，并在执行前查看同步计划。

![Obsidian Config Sync 同步范围界面](docs/images/ui%202026-09-04.png)

> **Beta 提示**
>
> 当前版本为 `v0.1.0-beta.1`，已经完成基本功能并经过个人使用测试，但仍可能存在未发现的问题。首次使用前，请务必手动备份 Vault，至少备份各个库的 `.obsidian` 目录。

## 当前功能

- 一个主库同步到多个目标库。
- 扫描并识别指定目录中的 Obsidian Vault。
- 按配置项选择同步范围。
- 在执行前展示新增和覆盖计划。
- 递归覆盖所选目录，但不删除目标库中的额外文件。
- 同步社区插件的程序和设置，并在目标库中启用插件。
- 完全离线运行，不依赖 Obsidian 插件 API。

## 下载与安装

当前 Beta 版本主要面向 Windows x64。

1. 打开 [GitHub Releases](https://github.com/Knowckx/obsidian-config-sync/releases)。
2. 找到 `Obsidian Config Sync v0.1.0 Beta 1`。
3. 下载 `obsidian-config-sync-v0.1.0-beta.1-windows-x64.exe`。
4. 运行安装程序并按提示完成安装。

当前安装程序尚未进行代码签名，Windows 可能显示 SmartScreen 警告。请确认安装包来自本项目的 GitHub Releases 页面后再决定是否继续运行。

## 使用方法

建议同步前关闭正在使用相关 Vault 的 Obsidian 窗口，避免 Obsidian 与本工具同时修改配置。

### 1. 扫描库

选择包含 Obsidian Vault 的目录。程序会扫描该目录本身及最多两层子目录，并列出其中包含 `.obsidian` 的 Vault。

不参与本轮同步的 Vault 可以从候选列表中移除；该操作只影响当前列表，不会删除磁盘目录。

### 2. 选择库

选择一个主库和至少一个目标库。

同步方向始终为：

```text
主库 → 目标库
```

主库是配置来源，目标库中的同名配置可能被覆盖，请在继续前确认选择无误。

### 3. 选择同步范围

选择需要同步的配置项和社区插件。可以恢复默认选择，也可以使用“全不选”后重新选择。

同步社区插件时，程序会：

```text
复制插件程序和设置 → 保留目标库原有启用列表 → 追加并启用所选插件
```

### 4. 检查同步计划

程序会按目标库列出本次新增和覆盖的配置项。此时尚未写入文件，请确认计划后再执行同步。

### 5. 查看执行结果

同步完成后，程序会分别显示每个目标库中的新增成功、覆盖成功和失败项。

## 同步规则

- 只处理用户选择的 `.obsidian` 配置项。
- 文件采用覆盖同步。
- 目录采用递归覆盖，但不会删除目标库中额外存在的文件。
- 社区插件的启用列表采用追加方式，不会移除或禁用目标库原有插件。
- 当前不提供镜像同步，也不会根据主库删除目标库配置。

## 数据安全

当前 Beta 版本尚未提供内置备份和回滚功能。

首次使用以及同步重要配置前，建议：

1. 退出正在使用相关 Vault 的 Obsidian。
2. 手动复制各个 Vault 的 `.obsidian` 目录。
3. 先使用非重要 Vault 完成一次同步验证。
4. 在同步计划中仔细检查主库、目标库和覆盖项。

## 已知限制

- 当前主要测试环境为 Windows 10。
- 当前仅提供覆盖同步，不支持镜像同步。
- 主库、目标库和同步范围不会持久化保存。
- 尚未经过大规模、多环境测试。
- 安装程序未进行代码签名，可能触发 Windows SmartScreen 警告。
- 当前版本没有内置备份和回滚能力。

## 问题反馈

如果遇到问题，请通过 [GitHub Issues](https://github.com/Knowckx/obsidian-config-sync/issues) 反馈，并尽量附上：

- Windows 版本。
- Obsidian 版本。
- 主库和目标库的大致目录结构，请注意移除隐私信息。
- 可以复现问题的操作步骤。
- 同步计划、错误提示或执行结果截图。

## 开发与构建

项目使用 Wails 3、Go、Svelte 5、TypeScript 和 pnpm。

本地开发：

```powershell
task dev
```

生成 Windows 安装包：

```powershell
task package
```

当前前端还依赖同一 pnpm workspace 中的本地包 `infa-s5`。直接克隆本仓库进行源码构建前，需要先提供该 workspace 依赖。

## 版本

当前公开测试版本：`v0.1.0-beta.1`
