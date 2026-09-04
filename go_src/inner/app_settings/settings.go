package app_settings

import "obsi-conf-sync/go_src/infa/file_store"

const (
	softwareName     = "obsi-conf-sync"
	settingsFileName = "settings.toml"
	lastVaultRootKey = "last_vault_root"
)

// GetLastVaultRoot 返回最近一次加载成功的 Vault 扫描根目录，尚未记录时返回空字符串。
func GetLastVaultRoot() (string, error) {
	store, err := file_store.New(softwareName, settingsFileName)
	if err != nil {
		return "", err
	}

	var root string
	_, err = store.GetOptional(lastVaultRootKey, &root)
	return root, err
}

// SaveLastVaultRoot 保存最近一次加载成功的 Vault 扫描根目录。
func SaveLastVaultRoot(root string) error {
	store, err := file_store.New(softwareName, settingsFileName)
	if err != nil {
		return err
	}
	return store.Save(lastVaultRootKey, root)
}
