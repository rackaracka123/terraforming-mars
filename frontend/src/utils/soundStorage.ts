/**
 * Sound settings persistence using localStorage
 */

const STORAGE_KEY = "terraforming-mars-sound";

export interface SoundSettings {
  enabled: boolean;
  volume: number;
}

const DEFAULT_SETTINGS: SoundSettings = {
  enabled: true,
  volume: 0.5,
};

/**
 * Get sound settings from localStorage, falling back to defaults
 */
export function getSoundSettings(): SoundSettings {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (!stored) {
      return DEFAULT_SETTINGS;
    }

    const parsed = JSON.parse(stored) as Partial<SoundSettings>;

    return {
      enabled: typeof parsed.enabled === "boolean" ? parsed.enabled : DEFAULT_SETTINGS.enabled,
      volume: typeof parsed.volume === "number" ? parsed.volume : DEFAULT_SETTINGS.volume,
    };
  } catch {
    return DEFAULT_SETTINGS;
  }
}

/**
 * Save sound settings to localStorage
 */
export function saveSoundSettings(settings: SoundSettings): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
  } catch {
    console.warn("Failed to save sound settings to localStorage");
  }
}
