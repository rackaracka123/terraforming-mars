/**
 * Icon type mappings for milestones and awards.
 * Used by GameIcon component for displaying achievement-related icons.
 */

/**
 * Get the icon type for a milestone.
 * @param milestoneType - The milestone type string (e.g., "terraformer", "mayor")
 * @returns The icon type to use with GameIcon
 */
export function getMilestoneIconType(milestoneType: string): string {
  switch (milestoneType) {
    case "terraformer":
      return "tr";
    case "mayor":
      return "city-tile";
    case "gardener":
      return "greenery-tile";
    case "builder":
      return "building";
    case "planner":
      return "card";
    default:
      return "milestone";
  }
}

/**
 * Get the icon type for an award.
 * @param awardType - The award type string (e.g., "landlord", "banker")
 * @returns The icon type to use with GameIcon
 */
export function getAwardIconType(awardType: string): string {
  switch (awardType) {
    case "landlord":
      return "city-tile";
    case "banker":
      return "credit";
    case "scientist":
      return "science";
    case "thermalist":
      return "heat";
    case "miner":
      return "steel";
    default:
      return "award";
  }
}
