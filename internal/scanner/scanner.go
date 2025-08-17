package scanner

import (
	"os"
	"path/filepath"
)

type VerkountFolder struct {
	Path   string
	Name   string
	Series string // Name of the series folder (direct child of ~/Documents)
}

func ScanForVerkountFolders(rootPath string) ([]VerkountFolder, error) {
	var folders []VerkountFolder

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			verkountPath := filepath.Join(path, ".verkount")
			if _, err := os.Stat(verkountPath); err == nil {
				// Determine the series name (direct child of rootPath)
				seriesName := getSeriesName(path, rootPath)

				folders = append(folders, VerkountFolder{
					Path:   path,
					Name:   filepath.Base(path),
					Series: seriesName,
				})
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return folders, nil
}

// getSeriesName extracts the name of the series folder (direct child of rootPath)
func getSeriesName(folderPath, rootPath string) string {
	relPath, err := filepath.Rel(rootPath, folderPath)
	if err != nil {
		return ""
	}

	// If the folder is directly in rootPath, it has no series
	if relPath == "." || relPath == "" {
		return ""
	}

	// Split by path separator and get the first component
	// This will be the series folder name
	dir := relPath
	for {
		parent := filepath.Dir(dir)
		if parent == "." || parent == "/" || parent == dir {
			return dir
		}
		dir = parent
	}
}
