import { useSoundEffects } from "./useSoundEffects.ts";

export function useHoverSound(disabled?: boolean) {
  const { playButtonHoverSound } = useSoundEffects();

  return {
    onMouseEnter: disabled ? undefined : () => void playButtonHoverSound(),
  };
}
