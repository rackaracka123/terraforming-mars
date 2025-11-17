import {
  PlayerEffectDto,
  StandardProjectConvertPlantsToGreenery,
  StandardProjectConvertHeatToTemperature,
} from "../types/generated/api-types.ts";

const BASE_PLANTS_FOR_GREENERY = 8;
const BASE_HEAT_FOR_TEMPERATURE = 8;

/**
 * Calculates the required plants for greenery conversion considering player discounts
 * @param playerEffects - The player's active effects
 * @returns The required number of plants after discounts (minimum 1)
 */
export function calculatePlantsForGreenery(
  playerEffects?: PlayerEffectDto[],
): number {
  return calculateDiscountedCost(
    BASE_PLANTS_FOR_GREENERY,
    playerEffects,
    StandardProjectConvertPlantsToGreenery,
  );
}

/**
 * Calculates the required heat for temperature conversion considering player discounts
 * @param playerEffects - The player's active effects
 * @returns The required number of heat after discounts (minimum 1)
 */
export function calculateHeatForTemperature(
  playerEffects?: PlayerEffectDto[],
): number {
  return calculateDiscountedCost(
    BASE_HEAT_FOR_TEMPERATURE,
    playerEffects,
    StandardProjectConvertHeatToTemperature,
  );
}

/**
 * Generic discount calculator for resource conversions
 * @param baseCost - The base cost before discounts
 * @param playerEffects - The player's active effects
 * @param standardProject - The standard project type
 * @returns The final cost after discounts (minimum 1)
 */
function calculateDiscountedCost(
  baseCost: number,
  playerEffects: PlayerEffectDto[] | undefined,
  standardProject: string,
): number {
  if (!playerEffects || playerEffects.length === 0) {
    return baseCost;
  }

  let totalDiscount = 0;

  // Loop through all player effects
  for (const effect of playerEffects) {
    // Check each output in the effect's behavior
    for (const output of effect.behavior.outputs ?? []) {
      // Check if this is a discount effect
      if (output.type !== "discount") {
        continue;
      }

      // Check if this discount applies to our conversion type
      if (
        !output.affectedStandardProjects?.includes(
          standardProject as never /* Cast needed for TS type system */,
        )
      ) {
        continue;
      }

      // Add the discount amount
      totalDiscount += output.amount;
    }
  }

  // Calculate final cost (minimum 1)
  const finalCost = baseCost - totalDiscount;
  return Math.max(finalCost, 1);
}
