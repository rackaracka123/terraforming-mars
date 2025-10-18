import { useCallback } from "react";
import audioService from "../services/audioService.ts";
import { Sound } from "../utils/soundStore.ts";

/**
 * React hook for playing sound effects in components.
 * Provides a type-safe interface to the audio service using Sound enum.
 *
 * @returns A function to play sounds using Sound enum values
 *
 * @example
 * ```tsx
 * function MyComponent() {
 *   const playSound = useSound();
 *
 *   const handleClick = () => {
 *     playSound(Sound.Button);  // ✅ Type-safe, autocomplete works
 *     // playSound('button');   // ❌ Won't compile
 *   };
 *
 *   return <button onClick={handleClick}>Click me</button>;
 * }
 * ```
 */
export function useSound() {
  return useCallback((sound: Sound) => {
    // Use void to explicitly discard the promise (prevents IDE warnings)
    void audioService.playSound(sound);
  }, []);
}
