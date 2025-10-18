import { Sound, getSoundPath } from "../utils/soundStore.ts";

/**
 * Audio service for managing game sound effects
 * Handles loading, caching, and playing audio files for game events
 */
class AudioService {
  private audioCache: Map<Sound, HTMLAudioElement> = new Map();
  private isEnabled: boolean = true;
  private volume: number = 0.5;

  constructor() {
    this.preloadAudioFiles();
  }

  /**
   * Preload audio files for better performance
   */
  private preloadAudioFiles() {
    // Preload all sounds defined in soundStore
    Object.values(Sound).forEach((sound) => {
      try {
        const path = getSoundPath(sound);
        const audio = new Audio(path);
        audio.preload = "auto";
        audio.volume = this.volume;

        // Handle loading events
        audio.addEventListener("canplaythrough", () => {
          // Audio preloaded successfully
        });

        audio.addEventListener("error", (e) => {
          console.warn(`⚠️ Failed to preload audio: ${sound}`, e);
        });

        this.audioCache.set(sound, audio);
      } catch (error) {
        console.warn(`⚠️ Error creating audio element for ${sound}:`, error);
      }
    });
  }

  /**
   * Play a sound effect using Sound enum (type-safe, no strings!)
   */
  public async playSound(sound: Sound): Promise<void> {
    if (!this.isEnabled) {
      return;
    }

    const audio = this.audioCache.get(sound);
    if (!audio) {
      console.warn(`⚠️ Sound not found: ${sound}`);
      return;
    }

    try {
      // Clone the audio element to allow overlapping sounds
      const audioClone = audio.cloneNode() as HTMLAudioElement;
      audioClone.volume = this.volume;

      await audioClone.play();
    } catch (error) {
      console.warn(`⚠️ Failed to play sound ${sound}:`, error);
    }
  }

  /**
   * Play production phase sound effect
   */
  public async playProductionSound(): Promise<void> {
    return this.playSound(Sound.Production);
  }

  /**
   * Set audio enabled/disabled
   */
  public setEnabled(enabled: boolean): void {
    this.isEnabled = enabled;
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
