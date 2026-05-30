//go:build !ci

package sound

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
)

// soundFS 把音效文件编译进二进制，让客户端二进制自包含、可独立分发。
// 仅嵌入当前启用的音效：lobby（大厅/BGM）、gaming/effects（非人声音效）、
// gaming/voices/male（男声播报）。gaming/voices/female 为备用女声，暂未启用，
// 故不嵌入以免无谓增大二进制。
//
//go:embed lobby gaming/effects gaming/voices/male
var soundFS embed.FS

type SoundManager struct {
	buffers map[string]*beep.Buffer
	enabled bool // 扬声器是否初始化成功

	// bgmMu guards the mute flag and the background-music controller, which can
	// be touched both from the init goroutine and the UI goroutine.
	bgmMu   sync.Mutex
	muted   bool       // 是否静音（启动后默认静音，需手动开启）；同时控制事件音效与 BGM
	bgmCtrl *beep.Ctrl // 常驻 mixer 的 BGM 控制器，用于暂停/恢复/切换曲目
	bgmName string     // 当前 BGM 曲目名（避免重复切换同一首）
}

func NewSoundManager() *SoundManager {
	return &SoundManager{
		buffers: make(map[string]*beep.Buffer),
		enabled: false,
		muted:   true,
	}
}

func (sm *SoundManager) Init() error {
	sampleRate := beep.SampleRate(44100)
	// Init speaker with smaller buffer for lower latency
	if err := speaker.Init(sampleRate, sampleRate.N(time.Second/10)); err != nil {
		return fmt.Errorf("failed to initialize speaker: %w", err)
	}
	sm.enabled = true

	// Load sounds from the embedded filesystem
	if err := sm.loadSoundFiles(sampleRate); err != nil {
		return err
	}

	return nil
}

// loadSoundFiles walks the embedded filesystem recursively and loads every
// sound file, keyed by its base filename without extension (e.g.
// gaming/voices/male/single/single_A.mp3 -> "single_A"). The directory layout
// is purely for human maintainability; callers still address sounds by their
// flat key. Base filenames are therefore expected to be unique across the
// embedded set.
func (sm *SoundManager) loadSoundFiles(sampleRate beep.SampleRate) error {
	return fs.WalkDir(soundFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		key := strings.TrimSuffix(name, filepath.Ext(name))
		// Continue loading other files even if one fails.
		_ = sm.loadSoundFile(p, key, sampleRate)
		return nil
	})
}

// loadSoundFile loads a single sound file at path `p` into the buffer under
// key `key`.
func (sm *SoundManager) loadSoundFile(p, key string, sampleRate beep.SampleRate) error {
	ext := strings.ToLower(filepath.Ext(p))
	if ext != ".mp3" && ext != ".wav" {
		return nil
	}

	f, err := soundFS.Open(p)
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

	sm.buffers[key] = buffer
	return nil
}

func (sm *SoundManager) Play(name string) {
	if !sm.enabled {
		return
	}
	sm.bgmMu.Lock()
	muted := sm.muted
	sm.bgmMu.Unlock()
	if muted {
		return
	}

	buffer, ok := sm.buffers[name]
	if !ok {
		// Silent failure if sound not found
		return
	}

	speaker.Play(buffer.Streamer(0, buffer.Len()))
}

// PlayBGM switches the looping background track to the given sound. It is a
// no-op if that track is already playing or the sound isn't loaded. The track
// keeps looping and respects the current mute state (paused while muted).
func (sm *SoundManager) PlayBGM(name string) {
	if !sm.enabled {
		return
	}
	sm.bgmMu.Lock()
	defer sm.bgmMu.Unlock()

	if sm.bgmName == name {
		return
	}
	buffer, ok := sm.buffers[name]
	if !ok {
		return
	}
	looped, err := beep.Loop2(buffer.Streamer(0, buffer.Len()))
	if err != nil {
		return
	}
	sm.bgmName = name

	if sm.bgmCtrl == nil {
		// First track: register a single persistent controller in the mixer.
		sm.bgmCtrl = &beep.Ctrl{Streamer: looped, Paused: sm.muted}
		speaker.Play(sm.bgmCtrl)
		return
	}
	// Swap the looping streamer in place so we don't pile up mixer entries.
	speaker.Lock()
	sm.bgmCtrl.Streamer = looped
	sm.bgmCtrl.Paused = sm.muted
	speaker.Unlock()
}

// StopBGM stops the background music and drops it from the mixer, so a later
// PlayBGM starts fresh.
func (sm *SoundManager) StopBGM() {
	sm.bgmMu.Lock()
	defer sm.bgmMu.Unlock()

	if sm.bgmCtrl == nil {
		return
	}
	speaker.Lock()
	sm.bgmCtrl.Streamer = nil // drained -> mixer removes it
	speaker.Unlock()
	sm.bgmCtrl = nil
	sm.bgmName = ""
}

// Muted reports whether sound is currently muted.
func (sm *SoundManager) Muted() bool {
	sm.bgmMu.Lock()
	defer sm.bgmMu.Unlock()
	return sm.muted
}

// ToggleMute flips the global mute state and returns the new value (true =
// muted). It pauses/resumes the background music and gates event sounds.
func (sm *SoundManager) ToggleMute() bool {
	sm.bgmMu.Lock()
	defer sm.bgmMu.Unlock()

	sm.muted = !sm.muted
	if sm.bgmCtrl != nil {
		speaker.Lock()
		sm.bgmCtrl.Paused = sm.muted
		speaker.Unlock()
	}
	return sm.muted
}

func (sm *SoundManager) Close() {
	sm.enabled = false
}
