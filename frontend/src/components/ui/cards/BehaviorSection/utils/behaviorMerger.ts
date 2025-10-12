import { ClassifiedBehavior } from "../types.ts";

/**
 * Merges multiple auto production behaviors into a single behavior.
 * This optimization reduces visual clutter by combining production resources
 * and immediate effects that occur automatically when the card is played.
 *
 * Behaviors with trigger conditions (e.g., "when greenery is placed") are NOT merged
 * to maintain clarity about when effects trigger.
 *
 * @param classifiedBehaviors - Array of behaviors to potentially merge
 * @returns Array with merged behaviors where applicable
 */
export const mergeAutoProductionBehaviors = (
  classifiedBehaviors: ClassifiedBehavior[],
): ClassifiedBehavior[] => {
  const autoProductionBehaviors: ClassifiedBehavior[] = [];
  const autoNoBackgroundBehaviors: ClassifiedBehavior[] = [];
  const otherBehaviors: ClassifiedBehavior[] = [];

  // Separate behaviors into categories
  classifiedBehaviors.forEach((classifiedBehavior) => {
    const { behavior, type } = classifiedBehavior;

    // Check if behavior has a trigger condition (e.g., greenery-placed for Herbivores)
    const hasCondition =
      behavior.triggers &&
      behavior.triggers.some((trigger: any) => trigger.condition);

    const isAutoProduction =
      type === "auto-no-background" &&
      behavior.outputs &&
      behavior.outputs.every(
        (output: any) =>
          output.type?.includes("production") ||
          output.resourceType?.includes("production") ||
          output.isProduction,
      );

    // Don't merge behaviors that have trigger conditions
    if (hasCondition) {
      otherBehaviors.push(classifiedBehavior);
    } else if (isAutoProduction) {
      autoProductionBehaviors.push(classifiedBehavior);
    } else if (type === "auto-no-background") {
      autoNoBackgroundBehaviors.push(classifiedBehavior);
    } else {
      otherBehaviors.push(classifiedBehavior);
    }
  });

  // Merge multiple auto production behaviors
  let mergedAutoProduction: ClassifiedBehavior | null = null;
  if (autoProductionBehaviors.length > 1) {
    const mergedOutputs: any[] = [];
    autoProductionBehaviors.forEach((classifiedBehavior) => {
      if (classifiedBehavior.behavior.outputs) {
        mergedOutputs.push(...classifiedBehavior.behavior.outputs);
      }
    });

    mergedAutoProduction = {
      behavior: {
        triggers: [{ type: "auto" }],
        outputs: mergedOutputs,
      },
      type: "auto-no-background",
    };
  } else if (autoProductionBehaviors.length === 1) {
    mergedAutoProduction = autoProductionBehaviors[0];
  }

  // Merge multiple auto-no-background behaviors (like Big Asteroid)
  let mergedAutoNoBackground: ClassifiedBehavior | null = null;
  if (autoNoBackgroundBehaviors.length > 1) {
    const mergedOutputs: any[] = [];
    autoNoBackgroundBehaviors.forEach((classifiedBehavior) => {
      if (classifiedBehavior.behavior.outputs) {
        mergedOutputs.push(...classifiedBehavior.behavior.outputs);
      }
    });

    mergedAutoNoBackground = {
      behavior: {
        triggers: [{ type: "auto" }],
        outputs: mergedOutputs,
      },
      type: "auto-no-background",
    };
  } else if (autoNoBackgroundBehaviors.length === 1) {
    mergedAutoNoBackground = autoNoBackgroundBehaviors[0];
  }

  // Combine results
  const result: ClassifiedBehavior[] = [];

  // If both production and immediate resources exist, merge them into a single behavior
  if (mergedAutoProduction && mergedAutoNoBackground) {
    const combinedBehavior: ClassifiedBehavior = {
      behavior: {
        triggers: [{ type: "auto" }],
        outputs: [
          ...(mergedAutoProduction.behavior.outputs || []),
          ...(mergedAutoNoBackground.behavior.outputs || []),
        ],
      },
      type: "auto-no-background",
    };
    result.push(combinedBehavior);
  } else {
    if (mergedAutoProduction) result.push(mergedAutoProduction);
    if (mergedAutoNoBackground) result.push(mergedAutoNoBackground);
  }

  result.push(...otherBehaviors);

  return result;
};
