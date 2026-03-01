package library

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type VideoFile struct {
	ID   string
	Path string
}

func BuildVideoPath(downloadDir, profile, dateYYYYMMDD, shortcode string) string {
	name := fmt.Sprintf("%s_%s-%s.mp4", profile, dateYYYYMMDD, shortcode)
	return filepath.Join(downloadDir, name)
}

func VideoIDFromPath(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

func ListVideoFiles(downloadDir string) ([]VideoFile, error) {
	entries, err := os.ReadDir(downloadDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []VideoFile{}, nil
		}
		return nil, err
	}

	out := make([]VideoFile, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || strings.ToLower(filepath.Ext(e.Name())) != ".mp4" {
			continue
		}
		p := filepath.Join(downloadDir, e.Name())
		out = append(out, VideoFile{ID: VideoIDFromPath(p), Path: p})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func PickRandomUnsent(files []VideoFile, sent map[string]struct{}) (VideoFile, bool) {
	unsent := make([]VideoFile, 0, len(files))
	for _, f := range files {
		if _, ok := sent[f.ID]; ok {
			continue
		}
		unsent = append(unsent, f)
	}
	if len(unsent) == 0 {
		return VideoFile{}, false
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return unsent[r.Intn(len(unsent))], true
}
