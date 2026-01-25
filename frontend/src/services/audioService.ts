import { getSoundSettings } from "../utils/soundStorage.ts";

/**
 * Audio service for managing game sound effects
 * Handles loading, caching, and playing audio files for game events
 */
class AudioService {
  private audioCache: Map<string, HTMLAudioElement> = new Map();
  private isEnabled: boolean = true;
  private volume: number = 0.5;

  constructor() {
    // Initialize from localStorage settings
    const settings = getSoundSettings();
    this.isEnabled = settings.enabled;
    this.volume = settings.volume;

    this.preloadAudioFiles();
  }

  /**
   * Preload audio files for better performance
   */
  private preloadAudioFiles() {
    const audioFiles = [
      { key: "production", path: "/assets/audio/production.mp3" },
      { key: "temperature-increase", path: "/sounds/temperature-increase.mp3" },
      { key: "water-placement", path: "/sounds/water-placement.mp3" },
      { key: "oxygen-increase", path: "/sounds/oxygen-increase.mp3" },
    ];

    audioFiles.forEach(({ key, path }) => {
      try {
        const audio = new Audio(path);
        audio.preload = "auto";
        audio.volume = this.volume;

        // Handle loading events
        audio.addEventListener("canplaythrough", () => {
          // Audio preloaded successfully
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
   * Play temperature increase sound effect
   */
  public async playTemperatureSound(): Promise<void> {
    return this.playSound("temperature-increase");
  }

  /**
   * Play water/ocean placement sound effect
   */
  public async playWaterPlacementSound(): Promise<void> {
    return this.playSound("water-placement");
  }

  /**
   * Play oxygen increase sound effect
   */
  public async playOxygenSound(): Promise<void> {
    return this.playSound("oxygen-increase");
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
