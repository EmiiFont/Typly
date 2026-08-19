package typly

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"strings"
	"sync"

	"github.com/rivo/uniseg"
	"golang.org/x/image/draw"
)

//go:embed assets/twemoji/*.png
var twemojiAssets embed.FS

type emojiLibrary struct {
	once  sync.Once
	paths map[string]string
	mu    sync.Mutex
	cache map[string]*image.RGBA
}

var emojis emojiLibrary

func (e *emojiLibrary) init() {
	e.paths = make(map[string]string)
	e.cache = make(map[string]*image.RGBA)
	_ = fs.WalkDir(twemojiAssets, "assets/twemoji", func(path string, entry fs.DirEntry, err error) error {
		if err == nil && !entry.IsDir() && strings.HasSuffix(path, ".png") {
			name := strings.TrimSuffix(entry.Name(), ".png")
			e.paths[name] = path
		}
		return nil
	})
}

// emojiForCluster returns a scaled Twemoji image for one grapheme cluster.
func emojiForCluster(cluster string, size int) *image.RGBA {
	emojis.once.Do(emojis.init)
	key := emojiKey(cluster)
	if key == "" {
		return nil
	}
	emojis.mu.Lock()
	defer emojis.mu.Unlock()
	cacheKey := fmt.Sprintf("%s@%d", key, size)
	if cached := emojis.cache[cacheKey]; cached != nil {
		return cached
	}
	path, ok := emojis.paths[key]
	if !ok {
		// Text input often omits the variation selector that is present in
		// the Twemoji filename.
		path, ok = emojis.paths[strings.ReplaceAll(key, "-fe0f", "")]
	}
	if !ok {
		return nil
	}
	data, err := fs.ReadFile(twemojiAssets, path)
	if err != nil {
		return nil
	}
	src, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	emojis.cache[cacheKey] = dst
	return dst
}

func emojiKey(cluster string) string {
	var parts []string
	graphemes := uniseg.NewGraphemes(cluster)
	for graphemes.Next() {
		for _, r := range graphemes.Runes() {
			parts = append(parts, fmt.Sprintf("%x", r))
		}
	}
	return strings.Join(parts, "-")
}
