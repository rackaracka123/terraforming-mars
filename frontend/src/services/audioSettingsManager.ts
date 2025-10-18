import audioService from "./audioService.ts";
import backgroundMusicService from "./backgroundMusicService.ts";

/**
 * Audio settings structure
 */
export interface AudioSettings {
  soundEffects: {
    enabled: boolean;
    volume: number;
  };
  backgroundMusic: {
    enabled: boolean;
    volume: number;
  };
}

/**
 * Default audio settings
 */
const DEFAULT_SETTINGS: AudioSettings = {
  soundEffects: {
    enabled: true,
    volume: 0.5,
  },
  backgroundMusic: {
    enabled: true,
    volume: 0.3,
  },
};

const STORAGE_KEY = "terraforming-mars-audio-settings";

/**
 * Audio Settings Manager
 * Manages audio settings persistence and application across audio services
 */
class AudioSettingsManager {
  private settings: AudioSettings = DEFAULT_SETTINGS;
  private initialized: boolean = false;

  /**
   * Initialize settings from localStorage and apply to audio services
   */
  public init(): void {
    if (this.initialized) return;

    this.loadSettings();
    this.applySettings();
    this.initialized = true;
  }

  /**
   * Load settings from localStorage
   */
  private loadSettings(): void {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        const parsed = JSON.parse(stored) as AudioSettings;
        this.settings = {
          soundEffects: {
            enabled:
              parsed.soundEffects?.enabled ??
              DEFAULT_SETTINGS.soundEffects.enabled,
            volume:
              parsed.soundEffects?.volume ??
              DEFAULT_SETTINGS.soundEffects.volume,
          },
          backgroundMusic: {
            enabled:
              parsed.backgroundMusic?.enabled ??
              DEFAULT_SETTINGS.backgroundMusic.enabled,
            volume:
              parsed.backgroundMusic?.volume ??
              DEFAULT_SETTINGS.backgroundMusic.volume,
          },
        };
      }
    } catch (error) {
      console.warn("⚠️ Failed to load audio settings:", error);
      this.settings = DEFAULT_SETTINGS;
    }
  }

  /**
   * Save settings to localStorage
   */
  private saveSettings(): void {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(this.settings));
    } catch (error) {
      console.warn("⚠️ Failed to save audio settings:", error);
    }
  }

  /**
   * Apply current settings to audio services
   */
  private applySettings(): void {
    // Apply sound effects settings
    audioService.setEnabled(this.settings.soundEffects.enabled);
    audioService.setVolume(this.settings.soundEffects.volume);

    // Apply background music settings
    backgroundMusicService.setEnabled(this.settings.backgroundMusic.enabled);
    backgroundMusicService.setVolume(this.settings.backgroundMusic.volume);
  }

  /**
   * Get current settings
   */
  public getSettings(): AudioSettings {
    return { ...this.settings };
  }

  /**
   * Update sound effects enabled state
   */
  public setSoundEffectsEnabled(enabled: boolean): void {
    this.settings.soundEffects.enabled = enabled;
    audioService.setEnabled(enabled);
    this.saveSettings();
  }

  /**
   * Update sound effects volume
   */
  public setSoundEffectsVolume(volume: number): void {
    this.settings.soundEffects.volume = Math.max(0, Math.min(1, volume));
    audioService.setVolume(this.settings.soundEffects.volume);
    this.saveSettings();
  }

  /**
   * Update background music enabled state
   */
  public setBackgroundMusicEnabled(enabled: boolean): void {
    this.settings.backgroundMusic.enabled = enabled;
    backgroundMusicService.setEnabled(enabled);
    this.saveSettings();
  }

  /**
   * Update background music volume
   */
  public setBackgroundMusicVolume(volume: number): void {
    this.settings.backgroundMusic.volume = Math.max(0, Math.min(1, volume));
    backgroundMusicService.setVolume(this.settings.backgroundMusic.volume);
    this.saveSettings();
  }

  /**
   * Reset to default settings
   */
  public resetToDefaults(): void {
    this.settings = { ...DEFAULT_SETTINGS };
    this.applySettings();
    this.saveSettings();
  }
}

// Singleton instance
export const audioSettingsManager = new AudioSettingsManager();
export default audioSettingsManager;
