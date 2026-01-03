//go:build !ci

package sound

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
)

type SoundManager struct {
	buffers map[string]*beep.Buffer
	enabled bool
}

func NewSoundManager() *SoundManager {
	return &SoundManager{
		buffers: make(map[string]*beep.Buffer),
		enabled: false,
	}
}

func (sm *SoundManager) Init() error {
	sampleRate := beep.SampleRate(44100)
	// Init speaker with smaller buffer for lower latency
	if err := speaker.Init(sampleRate, sampleRate.N(time.Second/10)); err != nil {
		return fmt.Errorf("failed to initialize speaker: %w", err)
	}
	sm.enabled = true

	// Load sounds from assets directory
	if err := sm.loadSoundFiles(sampleRate); err != nil {
		return err
	}

	return nil
}

// loadSoundFiles loads all sound files from the assets/sounds directory
func (sm *SoundManager) loadSoundFiles(sampleRate beep.SampleRate) error {
	soundDir := "assets/sounds"
	files, err := os.ReadDir(soundDir)
	if err != nil {
		// It's okay if directory doesn't exist, just no sounds
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read sound directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		ext := strings.ToLower(filepath.Ext(name))
		baseName := strings.TrimSuffix(name, filepath.Ext(name))

		if ext != ".mp3" && ext != ".wav" {
			continue
		}

		if err := sm.loadSoundFile(soundDir, name, baseName, ext, sampleRate); err != nil {
			// Continue loading other files even if one fails
			continue
		}
	}

	return nil
}

// loadSoundFile loads a single sound file into the buffer
func (sm *SoundManager) loadSoundFile(soundDir, name, baseName, ext string, sampleRate beep.SampleRate) error {
	path := filepath.Join(soundDir, name)
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	var streamer beep.StreamSeekCloser
	var format beep.Format

	switch ext {
	case ".mp3":
		streamer, format, err = mp3.Decode(f)
	case ".wav":
		streamer, format, err = wav.Decode(f)
	}

	if err != nil {
		return err
	}
	defer func() { _ = streamer.Close() }()

	// Resample if necessary
	var resampled beep.Streamer = streamer
	if format.SampleRate != sampleRate {
		resampled = beep.Resample(4, format.SampleRate, sampleRate, streamer)
	}

	// Use standard stereo format
	standardFormat := beep.Format{
		SampleRate:  sampleRate,
		NumChannels: 2,
		Precision:   4,
	}

	buffer := beep.NewBuffer(standardFormat)
	buffer.Append(resampled)

	sm.buffers[baseName] = buffer
	return nil
}

func (sm *SoundManager) Play(name string) {
	if !sm.enabled {
		return
	}

	buffer, ok := sm.buffers[name]
	if !ok {
		// Silent failure if sound not found
		return
	}

	speaker.Play(buffer.Streamer(0, buffer.Len()))
}

func (sm *SoundManager) Close() {
	sm.enabled = false
}
