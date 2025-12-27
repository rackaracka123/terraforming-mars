/**
 * Placement styling utilities for end-game displays.
 * Provides consistent colors and styles for 1st, 2nd, 3rd place, etc.
 */

export interface PlacementStyle {
  /** Podium height class (e.g., "h-32") */
  height: string;
  /** Background gradient */
  bg: string;
  /** Border color class */
  border: string;
  /** Text color class */
  text: string;
  /** Medal background class */
  medalBg: string;
  /** Medal text color class */
  medalText: string;
  /** Placement label (e.g., "1st", "2nd") */
  label: string;
}

/**
 * Get the full styling for a placement position.
 * @param placement - The placement number (1, 2, 3, etc.)
 * @returns PlacementStyle object with all style classes
 */
export function getPlacementStyle(placement: number): PlacementStyle {
  switch (placement) {
    case 1:
      return {
        height: "h-32",
        bg: "bg-gradient-to-t from-amber-600 to-amber-400",
        border: "border-amber-300",
        text: "text-amber-100",
        medalBg: "bg-amber-400",
        medalText: "text-amber-900",
        label: "1st",
      };
    case 2:
      return {
        height: "h-24",
        bg: "bg-gradient-to-t from-gray-500 to-gray-300",
        border: "border-gray-200",
        text: "text-gray-100",
        medalBg: "bg-gray-300",
        medalText: "text-gray-800",
        label: "2nd",
      };
    case 3:
      return {
        height: "h-16",
        bg: "bg-gradient-to-t from-amber-800 to-amber-600",
        border: "border-amber-500",
        text: "text-amber-100",
        medalBg: "bg-amber-700",
        medalText: "text-amber-100",
        label: "3rd",
      };
    default:
      return {
        height: "h-12",
        bg: "bg-gradient-to-t from-gray-700 to-gray-600",
        border: "border-gray-500",
        text: "text-gray-300",
        medalBg: "bg-gray-600",
        medalText: "text-gray-300",
        label: `${placement}th`,
      };
  }
}
