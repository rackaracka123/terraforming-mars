import { ClassifiedBehavior, TileScaleInfo } from "../types.ts";
import { TILE_PLACEMENT_TYPES } from "../constants.ts";

/**
 * Detects if behaviors contain tile placements and calculates appropriate scaling
 * @param classifiedBehaviors - Array of classified behaviors
 * @returns Scale information for tile placement icons
 *
 * Edge cases handled:
 * - 1 tile placement (alone) → 1.5x scale (50% bigger)
 * - 2 tile placements (alone) → 1.25x scale (25% bigger)
 * - Otherwise → 1x scale (normal size)
 */
export const detectTilePlacementScale = (
  classifiedBehaviors: ClassifiedBehavior[],
): TileScaleInfo => {
  // Find all tile placements across all behaviors
  const tilePlacements: string[] = [];
  const tileBehaviorIndices: number[] = [];

  for (let i = 0; i < classifiedBehaviors.length; i++) {
    const behavior = classifiedBehaviors[i].behavior;
    if (behavior.outputs && behavior.outputs.length > 0) {
      for (const output of behavior.outputs) {
        const cleanType = output.type?.toLowerCase().replace(/[_\s]/g, "-");
        if (TILE_PLACEMENT_TYPES.includes(cleanType as string)) {
          tilePlacements.push(cleanType as string);
          if (!tileBehaviorIndices.includes(i)) {
            tileBehaviorIndices.push(i);
          }
        }
      }
    }
  }

  // No tile placements found
  if (tilePlacements.length === 0) {
    return { scale: 1, tileType: null, tileCount: 0 };
  }

  // Check if tile placements are in behaviors with other outputs
  for (const index of tileBehaviorIndices) {
    const behavior = classifiedBehaviors[index].behavior;
    const hasOtherOutputs =
      behavior.outputs &&
      behavior.outputs.filter((output: any) => {
        const cleanType = output.type?.toLowerCase().replace(/[_\s]/g, "-");
        return !TILE_PLACEMENT_TYPES.includes(cleanType as string);
      }).length > 0;

    if (hasOtherOutputs) {
      return { scale: 1, tileType: null, tileCount: tilePlacements.length };
    }
  }

  // Count other behaviors (production, actions, effects)
  const hasProductionBox = classifiedBehaviors.some(
    (cb) =>
      (cb.behavior as any).productionOutputs && (cb.behavior as any).productionOutputs.length > 0,
  );
  const hasActionBox = classifiedBehaviors.some(
    (cb) =>
      (cb.behavior.triggers && cb.behavior.triggers[0]?.type === "manual") || cb.behavior.choices,
  );
  const hasTriggeredEffect = classifiedBehaviors.some(
    (cb) =>
      cb.behavior.triggers &&
      cb.behavior.triggers[0]?.type === "auto" &&
      (cb.behavior as any).condition !== undefined &&
      (cb.behavior as any).condition !== null,
  );

  // Edge case 1: Exactly 2 tile placements, nothing else → 1.25x scale
  if (
    tilePlacements.length === 2 &&
    classifiedBehaviors.length === tileBehaviorIndices.length &&
    !hasProductionBox &&
    !hasActionBox &&
    !hasTriggeredEffect
  ) {
    return {
      scale: 1.25,
      tileType: tilePlacements[0],
      tileCount: tilePlacements.length,
    };
  }

  // Edge case 2: Exactly 1 tile placement, nothing else → 1.5x scale
  if (
    tilePlacements.length === 1 &&
    classifiedBehaviors.length === 1 &&
    !hasProductionBox &&
    !hasActionBox &&
    !hasTriggeredEffect
  ) {
    return { scale: 1.5, tileType: tilePlacements[0], tileCount: 1 };
  }

  // Tile placement(s) + production/action/effect → normal scale
  return { scale: 1, tileType: null, tileCount: tilePlacements.length };
};
