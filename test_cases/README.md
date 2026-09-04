# 测试场地说明

本目录用于验证 Obsidian 配置同步 MVP，提供一套可重复执行的“核心测试”流程。
每次开发完成后，默认执行一次核心测试，用一次同步覆盖大部分常规功能和关键边界。

## Vault 角色

- `vault1`：主库，提供待同步的配置文件和目录。
- `vault2`：最小目标库，用于测试新增配置和已有文件覆盖。
- `vault3`：混合目标库，用于测试文件覆盖、目录递归复制、目标库额外文件保留，以及 `themes` 路径类型冲突。

## 核心测试流程

1. 以 `vault1` 作为主库，选择 `vault2`、`vault3` 作为从库。
2. 固定选择以下同步项：

   ```text
   app.json
   snippets/
   themes
   plugins/open-in-new-tab/
   plugins/open-tab-settings/
   ```

3. 生成同步计划，检查新增、覆盖和失败项分类。
4. 确认执行，检查两个目标库的结果是否分别符合预期。
5. 检查目标库文件差异，确认 `vault3/.obsidian/snippets/target-only.css` 未被删除。

执行前应保留本目录中的基准内容；如果同步直接修改了这些目录，应先恢复基准内容再开始下一轮测试。

## 核心测试标准结果

### vault2

同步计划应包含：

```text
覆盖1：app.json
新增4：snippets、themes、plugins/open-in-new-tab、plugins/open-tab-settings
```

执行结果应为：

```text
成功覆盖 app.json
成功新增其余选中配置
无失败项
```

### vault3

同步计划应包含：

```text
覆盖：app.json、snippets、themes、plugins/open-in-new-tab
新增：plugins/open-tab-settings
```

执行结果应为：

```text
成功覆盖 app.json、snippets、plugins/open-in-new-tab
成功新增 plugins/open-tab-settings
themes 产生失败项：主库是目录，目标库是文件
```

执行后必须确认：

- `vault3/.obsidian/snippets/table-spacing-fix.css` 已被主库版本覆盖。
- `vault3/.obsidian/snippets/vscode_light.css` 已新增。
- `vault3/.obsidian/snippets/target-only.css` 仍然保留。
- `vault3/.obsidian/themes` 仍是文件，未被删除或替换。
- `vault2/.obsidian/community-plugins.json` 已创建，并按选择顺序写入两个插件 ID。
- `vault3/.obsidian/community-plugins.json` 保留原有插件，只追加缺少的插件 ID。

以上结果构成核心测试的标准预期；结果不一致时，应先排查同步逻辑，再修改测试场地。

## 后端回归测试入口

后端核心回归测试入口为：

```text
go_src/inner/svc/api_test.go
```

该测试使用本目录 vault 的临时副本，按固定核心测试流程调用后端接口并进行断言；当前已接入核心测试标准结果。

## 预期覆盖范围

一次固定流程应覆盖：

- 文件新增与覆盖
- 目录新增与递归复制
- 目录部分覆盖
- 目标库额外文件保留
- `themes` 目录与目标文件的类型冲突
- 多目标库结果分开记录
- 前端重复点击时的重复执行防护
