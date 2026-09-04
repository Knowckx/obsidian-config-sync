package file_store

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/pelletier/go-toml/v2"
)

// newStore 创建 Store 并确保目标目录和文件存在。
func newStore(softName string, fileName string) (*Store, error) {
	if strings.TrimSpace(softName) == "" {
		return nil, errors.New("softName 不能为空")
	}
	if strings.TrimSpace(fileName) == "" {
		return nil, errors.New("fileName 不能为空")
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	// C:\Users\AA\AppData\Roaming\{softName}\
	storeDir := filepath.Join(configDir, softName)
	if err := os.MkdirAll(storeDir, 0700); err != nil {
		return nil, err
	}

	storePath := filepath.Join(storeDir, tomlFileName(fileName))
	file, err := os.OpenFile(storePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	if err := file.Close(); err != nil {
		return nil, err
	}
	return &Store{path: storePath}, nil
}

// tomlFileName 补齐 TOML 文件扩展名。
func tomlFileName(fileName string) string {
	if strings.HasSuffix(fileName, ".toml") {
		return fileName
	}
	return fileName + ".toml"
}

// Store 表示一个用户配置目录下的 TOML 文件存储。
type Store struct {
	path string
}

// Path 返回当前存储文件的完整路径。
func (s *Store) Path() string {
	return s.path
}

// Save 保存指定 key 的值。
func (s *Store) Save(key string, value any) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("key 不能为空")
	}

	// 先读
	data, err := s.readAll()
	if err != nil {
		return errors.WithStack(err)
	}

	// 修改
	data[key] = value
	content, err := toml.Marshal(data)
	if err != nil {
		return errors.WithStack(err)
	}

	// 写回
	if err := os.WriteFile(s.path, content, 0600); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Get 读取指定 key 的值并写入 out。
func (s *Store) Get(key string, out any) error {
	found, err := s.GetOptional(key, out)
	if err != nil {
		return errors.WithStack(err)
	}
	if !found {
		return errors.Errorf("key 不存在: %s", key)
	}
	return nil
}

// GetOptional 读取指定 key；key 不存在时返回 false。
func (s *Store) GetOptional(key string, out any) (bool, error) {
	data, err := s.readAll()
	if err != nil {
		return false, errors.WithStack(err)
	}

	value, ok := data[key]
	if !ok {
		return false, nil
	}

	if err := decodeValue(key, value, out); err != nil {
		return false, errors.WithStack(err)
	}
	return true, nil
}

// readAll 读取整个 TOML 文件。
func (s *Store) readAll() (map[string]any, error) {
	content, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(string(content)) == "" {
		return map[string]any{}, nil
	}

	data := map[string]any{}
	if err := toml.Unmarshal(content, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// decodeValue 将 TOML 值还原到调用方传入的目标类型。
func decodeValue(key string, value any, out any) error {
	outValue := reflect.ValueOf(out)
	if !outValue.IsValid() || outValue.Kind() != reflect.Ptr || outValue.IsNil() {
		return errors.New("out 必须是非空指针")
	}

	target := outValue.Elem()
	if !target.CanSet() {
		return errors.New("out 指向的值不可写")
	}

	valueType := reflect.TypeOf(value)
	if valueType == nil {
		return errors.Errorf("key 值为空: %s", key)
	}

	valueReflect := reflect.ValueOf(value)
	if valueType.AssignableTo(target.Type()) {
		target.Set(valueReflect)
		return nil
	}
	if valueType.ConvertibleTo(target.Type()) {
		target.Set(valueReflect.Convert(target.Type()))
		return nil
	}

	return decodeValueByToml(key, value, target)
}

// decodeValueByToml 通过 TOML 编解码还原复杂结构。
func decodeValueByToml(key string, value any, target reflect.Value) error {
	wrapperType := reflect.MapOf(reflect.TypeOf(""), target.Type())
	wrapperPtr := reflect.New(wrapperType)

	content, err := toml.Marshal(map[string]any{key: value})
	if err != nil {
		return err
	}
	if err := toml.Unmarshal(content, wrapperPtr.Interface()); err != nil {
		return err
	}

	result := wrapperPtr.Elem().MapIndex(reflect.ValueOf(key))
	if !result.IsValid() {
		return errors.Errorf("key 解析失败: %s", key)
	}

	target.Set(result)
	return nil
}
