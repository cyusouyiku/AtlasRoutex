package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	scriptsDir := filepath.Join("cmd", "migration", "scripts")
	entries, err := os.ReadDir(scriptsDir)
	if err != nil {
		fmt.Println("migration scripts not found:", err)
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		fmt.Println("apply", e.Name())
	}
}
