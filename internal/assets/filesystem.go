package assets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"amr-disp/internal/domain"
)

type FilesystemAssets struct {
	Root string
}

func NewFilesystemAssets(root string) *FilesystemAssets {
	return &FilesystemAssets{Root: root}
}

func (f *FilesystemAssets) ListMaps() ([]string, error) {
	return f.listJSONFiles(filepath.Join(f.Root, "fixtures", "maps"))
}

func (f *FilesystemAssets) ListScenarios() ([]string, error) {
	return f.listJSONFiles(filepath.Join(f.Root, "fixtures", "scenarios"))
}

func (f *FilesystemAssets) LoadMap(path string) (domain.MapData, error) {
	data, err := os.ReadFile(f.resolve(path))
	if err != nil {
		return domain.MapData{}, err
	}
	var item domain.MapData
	if err := json.Unmarshal(data, &item); err != nil {
		return domain.MapData{}, err
	}
	return item, nil
}

func (f *FilesystemAssets) LoadScenario(path string) (domain.Scenario, error) {
	data, err := os.ReadFile(f.resolve(path))
	if err != nil {
		return domain.Scenario{}, err
	}
	var item domain.Scenario
	if err := json.Unmarshal(data, &item); err != nil {
		return domain.Scenario{}, err
	}
	return item, nil
}

func (f *FilesystemAssets) resolve(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(f.Root, path)
}

func (f *FilesystemAssets) listJSONFiles(root string) ([]string, error) {
	items, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0)
	for _, item := range items {
		if item.IsDir() || !strings.HasSuffix(strings.ToLower(item.Name()), ".json") {
			continue
		}
		fullPath := filepath.Join(root, item.Name())
		relPath, err := filepath.Rel(f.Root, fullPath)
		if err != nil {
			relPath = fullPath
		}
		result = append(result, filepath.ToSlash(relPath))
	}
	sortStrings(result)
	return result, nil
}

func sortStrings(items []string) {
	if len(items) < 2 {
		return
	}
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j] < items[i] {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}
