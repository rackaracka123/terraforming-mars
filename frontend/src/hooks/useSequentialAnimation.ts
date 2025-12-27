import { useState, useEffect, useCallback } from "react";

/**
 * Hook for managing sequential item reveal animations.
 * Returns the current animated index, indicating which items should be visible.
 *
 * @param itemCount - Total number of items to animate
 * @param delayMs - Delay between each item reveal in milliseconds
 * @param isAnimating - Whether the animation should be running
 * @param onComplete - Optional callback when animation completes
 * @returns Current animated index (-1 means no items shown yet)
 */
export function useSequentialAnimation(
  itemCount: number,
  delayMs: number,
  isAnimating: boolean,
  onComplete?: () => void,
): number {
  const [animatedIndex, setAnimatedIndex] = useState(-1);

  // Memoize onComplete to prevent unnecessary effect re-runs
  const onCompleteCallback = useCallback(() => {
    onComplete?.();
  }, [onComplete]);

  useEffect(() => {
    if (!isAnimating) {
      return;
    }

    // Reset to start
    setAnimatedIndex(-1);
    let currentIndex = 0;

    const interval = setInterval(() => {
      if (currentIndex < itemCount) {
        setAnimatedIndex(currentIndex);
        currentIndex++;
      } else {
        clearInterval(interval);
        onCompleteCallback();
      }
    }, delayMs);

    return () => clearInterval(interval);
  }, [isAnimating, itemCount, delayMs, onCompleteCallback]);

  return animatedIndex;
}

/**
 * Helper to check if an item at a given index should be visible.
 *
 * @param index - Item index to check
 * @param animatedIndex - Current animated index from useSequentialAnimation
 * @param isAnimating - Whether animation is running
 * @returns true if the item should be visible
 */
export function isItemVisible(index: number, animatedIndex: number, isAnimating: boolean): boolean {
  return !isAnimating || index <= animatedIndex;
}
