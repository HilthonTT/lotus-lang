package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// watchFile runs filePath every time it changes on disk.
// It polls for changes (no external dependencies needed).
// Ctrl+C exits cleanly.
func watchFile(filePath, engine string) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		fatal("watch: " + err.Error())
	}

	fmt.Printf("\033[2m  watching %s — press Ctrl+C to stop\033[0m\n\n", filepath.Base(filePath))

	var lastMod time.Time
	for {
		info, err := os.Stat(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "watch: %s\n", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if !info.ModTime().Equal(lastMod) {
			if !lastMod.IsZero() {
				fmt.Printf("\n\033[2m─── %s ───\033[0m\n\n", time.Now().Format("15:04:05"))
			}
			lastMod = info.ModTime()
			runFile(absPath, engine)
		}
		time.Sleep(200 * time.Millisecond)
	}
}
