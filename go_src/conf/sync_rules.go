package conf

// ConfigItemRules 是已知配置项的同步规则。
var ConfigItemRules = map[string]ConfigItemRule{
	// 默认同步
	"app.json": {
		Description:     "编辑器的通用配置",
		DefaultSelected: true,
	},
	"appearance.json": {
		Description:     "外观配置和启用的 CSS 片段，应与 snippets/ 一起同步",
		DefaultSelected: true,
	},
	"core-plugins.json": {
		Description:     "Obsidian 核心插件开关",
		DefaultSelected: true,
	},
	"hotkeys.json": {
		Description:     "自定义快捷键",
		DefaultSelected: true,
	},
	"snippets/": {
		Description:     "自定义编辑器样式",
		DefaultSelected: true,
	},
	"themes/": {
		Description:     "已安装的主题，应与 appearance.json 一起同步",
		DefaultSelected: true,
	},

	// 默认跳过
	"graph.json": {
		Description:     "关系图谱的显示和布局设置，通常不需要跨库同步",
		DefaultSelected: false,
	},
	"bookmarks.json": {
		Description:     "当前库的书签列表，通常包含仅在本库存在的文件",
		DefaultSelected: false,
	},
	"types.json": {
		Description:     "当前库的属性类型定义，可能依赖本库的笔记结构",
		DefaultSelected: false,
	},
	"workspace.json": {
		Description:     "桌面端当前打开文件和界面布局，变化频繁且仅适用于当前库",
		DefaultSelected: false,
	},
	"workspace-mobile.json": {
		Description:     "移动端当前打开文件和界面布局，变化频繁且仅适用于当前库",
		DefaultSelected: false,
	},
	"workspaces.json": {
		Description:     "保存的工作区布局，可能引用仅在当前库存在的文件",
		DefaultSelected: false,
	},
}

// ConfigItemRule 定义已知配置项的展示和默认选择规则。
type ConfigItemRule struct {
	Description     string
	DefaultSelected bool
	IsPlugin        bool
}
