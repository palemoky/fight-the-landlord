//go:build ci

package sound

type SoundManager struct{}

func NewSoundManager() *SoundManager {
	return &SoundManager{}
}

func (sm *SoundManager) Init() error {
	return nil
}

func (sm *SoundManager) Play(name string) {
	// No-op
}

func (sm *SoundManager) Close() {
	// No-op
}
