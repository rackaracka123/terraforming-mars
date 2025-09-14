/**
 * Audio service for managing game sound effects
 * Handles loading, caching, and playing audio files for game events
 */
class AudioService {
  private audioCache: Map<string, HTMLAudioElement> = new Map();
  private isEnabled: boolean = true;
  private volume: number = 0.5;

  constructor() {
    this.preloadAudioFiles();
  }

  /**
   * Preload audio files for better performance
   */
  private preloadAudioFiles() {
    const audioFiles = [
      { key: "production", path: "/assets/audio/production.mp3" },
    ];

    audioFiles.forEach(({ key, path }) => {
      try {
        const audio = new Audio(path);
        audio.preload = "auto";
        audio.volume = this.volume;

        // Handle loading events
        audio.addEventListener("canplaythrough", () => {
          console.log(`🔊 Audio preloaded: ${key}`);
        });

        audio.addEventListener("error", (e) => {
          console.warn(`⚠️ Failed to preload audio: ${key}`, e);
        });

        this.audioCache.set(key, audio);
      } catch (error) {
        console.warn(`⚠️ Error creating audio element for ${key}:`, error);
      }
    });
  }

  /**
   * Play a sound effect by key
   */
  public async playSound(soundKey: string): Promise<void> {
    if (!this.isEnabled) {
      return;
    }

    const audio = this.audioCache.get(soundKey);
    if (!audio) {
      console.warn(`⚠️ Sound not found: ${soundKey}`);
      return;
    }

    try {
      // Clone the audio element to allow overlapping sounds
      const audioClone = audio.cloneNode() as HTMLAudioElement;
      audioClone.volume = this.volume;

      await audioClone.play();
      console.log(`🔊 Playing sound: ${soundKey}`);
    } catch (error) {
      console.warn(`⚠️ Failed to play sound ${soundKey}:`, error);
    }
  }

  /**
   * Play production phase sound effect
   */
  public async playProductionSound(): Promise<void> {
    return this.playSound("production");
  }

  /**
   * Set audio enabled/disabled
   */
  public setEnabled(enabled: boolean): void {
    this.isEnabled = enabled;
    console.log(`🔊 Audio ${enabled ? "enabled" : "disabled"}`);
  }

  /**
   * Set audio volume (0.0 to 1.0)
   */
  public setVolume(volume: number): void {
    this.volume = Math.max(0, Math.min(1, volume));

    // Update volume on all cached audio elements
    this.audioCache.forEach((audio) => {
      audio.volume = this.volume;
    });

    console.log(`🔊 Audio volume set to ${(this.volume * 100).toFixed(0)}%`);
  }

  /**
   * Get current audio settings
   */
  public getSettings() {
    return {
      enabled: this.isEnabled,
      volume: this.volume,
    };
  }
}

// Singleton instance for application-wide use
export const audioService = new AudioService();
export default audioService;
