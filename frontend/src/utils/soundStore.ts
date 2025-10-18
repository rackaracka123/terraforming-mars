/**
 * Centralized sound asset path lookup store.
 * All sound path mappings are defined here to avoid duplication across components.
 * Uses enum for maximum type safety - no raw strings allowed.
 */

/**
 * Sound enum - use these values instead of strings
 * @example
 * playSound(Sound.Button)  // ✅ Type-safe
 * playSound('button')      // ❌ Compile error
 */
export enum Sound {
  Button = "button",
  Production = "production",
}

/**
 * Sound asset paths mapped by Sound enum
 */
const SOUND_PATHS: Record<Sound, string> = {
  [Sound.Button]: "/assets/audio/button.mp3",
  [Sound.Production]: "/assets/audio/production.mp3",
};

/**
 * Get the sound path for a given Sound enum value.
 */
export function getSoundPath(sound: Sound): string {
  return SOUND_PATHS[sound];
}
