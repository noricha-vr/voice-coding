package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/noricha-vr/voicecode/internal/app"
	"github.com/noricha-vr/voicecode/internal/core/settings"
	"github.com/noricha-vr/voicecode/internal/core/transcriber"
	"github.com/noricha-vr/voicecode/internal/platform/clipboard"
	"github.com/noricha-vr/voicecode/internal/platform/hotkey"
	"github.com/noricha-vr/voicecode/internal/platform/overlay"
	"github.com/noricha-vr/voicecode/internal/platform/recorder"
	"github.com/noricha-vr/voicecode/internal/platform/sound"
	"github.com/noricha-vr/voicecode/internal/platform/tray"
)

func main() {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "transcribe":
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "Usage: voicecode transcribe <wav-file>")
				os.Exit(1)
			}
			runTranscribe(os.Args[2])
			return
		case "help", "-h", "--help":
			printUsage()
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
			printUsage()
			os.Exit(1)
		}
	}

	// GUI mode
	runGUI()
}

func printUsage() {
	fmt.Println("Usage: voicecode [command]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  transcribe <wav-file>   Transcribe a WAV file")
	fmt.Println("  help                    Show this help message")
	fmt.Println()
	fmt.Println("Without a command, starts in GUI mode with system tray.")
}

func runTranscribe(wavPath string) {
	if _, err := os.Stat(wavPath); os.IsNotExist(err) {
		log.Fatalf("File not found: %s", wavPath)
	}

	ctx := context.Background()
	t, err := transcriber.New(ctx, "")
	if err != nil {
		log.Fatalf("Failed to initialize transcriber: %v", err)
	}

	text, elapsed, err := t.Transcribe(ctx, wavPath)
	if err != nil {
		log.Fatalf("Transcription failed: %v", err)
	}

	fmt.Println(text)
	log.Printf("Elapsed: %.2fs, Model: %s", elapsed, t.ModelName())
}

func runGUI() {
	ctx := context.Background()

	// Load settings
	cfg, err := settings.Load()
	if err != nil {
		log.Printf("[Init] Settings load failed, using defaults: %v", err)
		cfg = settings.Default()
	}

	// Initialize transcriber
	tr, err := transcriber.New(ctx, "")
	if err != nil {
		log.Fatalf("[Init] Transcriber init failed: %v", err)
	}

	// Initialize platform adapters
	rec, err := recorder.NewRecorder()
	if err != nil {
		log.Fatalf("[Init] Recorder init failed: %v", err)
	}

	clip, err := clipboard.NewClipboard()
	if err != nil {
		log.Fatalf("[Init] Clipboard init failed: %v", err)
	}

	snd := sound.NewPlayer()
	if err := sound.WarmUp(); err != nil {
		log.Printf("[Init] Sound warmup failed: %v", err)
	}
	ov := overlay.NewOverlay()
	hk := hotkey.NewManager()
	tm := tray.NewManager()

	// Create and run app
	a := app.New(cfg, tr, rec, clip, snd, ov, hk, tm)
	a.Run()
}
