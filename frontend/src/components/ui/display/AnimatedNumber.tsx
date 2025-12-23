import { FC, useEffect, useRef, useState } from "react";

interface AnimatedNumberProps {
  /** Target value to animate to */
  value: number;
  /** Duration of animation in milliseconds */
  duration?: number;
  /** Starting value (defaults to 0) */
  startValue?: number;
  /** Callback when animation completes */
  onComplete?: () => void;
  /** Additional CSS classes */
  className?: string;
  /** Whether to animate on value change (vs only on mount) */
  animateOnChange?: boolean;
  /** Format function for the displayed number */
  format?: (value: number) => string;
}

/**
 * AnimatedNumber - Displays a number that counts up/down smoothly
 *
 * Uses requestAnimationFrame for smooth 60fps animation.
 * Applies vpCountUp animation class on each increment for visual pulse effect.
 */
const AnimatedNumber: FC<AnimatedNumberProps> = ({
  value,
  duration = 1500,
  startValue = 0,
  onComplete,
  className = "",
  animateOnChange = true,
  format = (v) => Math.round(v).toString(),
}) => {
  const [displayValue, setDisplayValue] = useState(startValue);
  const [isPulsing, setIsPulsing] = useState(false);
  const animationRef = useRef<number | null>(null);
  const startTimeRef = useRef<number | null>(null);
  const previousValueRef = useRef(startValue);

  useEffect(() => {
    const fromValue = animateOnChange ? previousValueRef.current : startValue;
    const toValue = value;

    // Don't animate if values are the same, but still call onComplete
    if (fromValue === toValue) {
      setDisplayValue(toValue);
      onComplete?.();
      return;
    }

    // Cancel any existing animation
    if (animationRef.current) {
      cancelAnimationFrame(animationRef.current);
    }

    startTimeRef.current = null;

    const animate = (timestamp: number) => {
      if (startTimeRef.current === null) {
        startTimeRef.current = timestamp;
      }

      const elapsed = timestamp - startTimeRef.current;
      const progress = Math.min(elapsed / duration, 1);

      // Easing function (ease-out cubic)
      const easedProgress = 1 - Math.pow(1 - progress, 3);

      const currentValue = fromValue + (toValue - fromValue) * easedProgress;
      setDisplayValue(currentValue);

      // Trigger pulse animation on significant changes
      const roundedCurrent = Math.round(currentValue);
      const roundedPrevious = Math.round(previousValueRef.current);
      if (roundedCurrent !== roundedPrevious && roundedCurrent <= toValue) {
        setIsPulsing(true);
        setTimeout(() => setIsPulsing(false), 300);
      }

      if (progress < 1) {
        animationRef.current = requestAnimationFrame(animate);
      } else {
        setDisplayValue(toValue);
        previousValueRef.current = toValue;
        onComplete?.();
      }
    };

    animationRef.current = requestAnimationFrame(animate);

    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
    };
  }, [value, duration, startValue, animateOnChange, onComplete]);

  // Update previous value when animation starts
  useEffect(() => {
    return () => {
      previousValueRef.current = value;
    };
  }, [value]);

  return (
    <span
      className={`inline-block tabular-nums ${isPulsing ? "vp-count-animate" : ""} ${className}`}
    >
      {format(displayValue)}
    </span>
  );
};

export default AnimatedNumber;
