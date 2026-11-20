package emoji

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed data/github.json data/gitmoji.json
var dataFS embed.FS

// GitHub returns the built-in GitHub emoji pack initialized lazily.
var GitHub = sync.OnceValue(func() *Pack {
	pack, err := loadGitHubPack()
	if err != nil {
		// Embedded data is static and validated in unit tests; failure indicates a corrupt binary build.
		panic(fmt.Sprintf("emoji: loading github dataset: %v", err))
	}

	return pack
})

// Gitmoji returns the built-in Gitmoji pack initialized lazily.
var Gitmoji = sync.OnceValue(func() *Pack {
	pack, err := loadGitmojiPack()
	if err != nil {
		// Embedded data is static and validated in unit tests; failure indicates a corrupt binary build.
		panic(fmt.Sprintf("emoji: loading gitmoji dataset: %v", err))
	}

	return pack
})

func loadGitHubPack() (*Pack, error) {
	raw, err := dataFS.ReadFile("data/github.json")
	if err != nil {
		return nil, fmt.Errorf("reading embedded github data: %w", err)
	}

	var pairs [][2]string
	if err := json.Unmarshal(raw, &pairs); err != nil {
		return nil, fmt.Errorf("parsing embedded github data: %w", err)
	}

	pack := NewPack()
	for _, pair := range pairs {
		pack.Add(Emoji{
			Code:  pair[0],
			Emoji: pair[1],
		})
	}

	return pack, nil
}

func loadGitmojiPack() (*Pack, error) {
	raw, err := dataFS.ReadFile("data/gitmoji.json")
	if err != nil {
		return nil, fmt.Errorf("reading embedded gitmoji data: %w", err)
	}

	var items []Emoji
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("parsing embedded gitmoji data: %w", err)
	}

	pack := NewPack()
	for _, item := range items {
		pack.Add(item)
	}

	return pack, nil
}
