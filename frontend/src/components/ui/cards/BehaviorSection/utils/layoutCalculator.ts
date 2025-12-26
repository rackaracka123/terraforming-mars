import { LayoutRequirement, LayoutPlan, IconDisplayInfo } from "../types.ts";
import { MAX_HORIZONTAL_ICONS } from "../constants.ts";
import { analyzeResourceDisplayWithConstraints } from "./displayAnalysis.ts";

/**
 * Calculates icon space requirements for a behavior.
 * Counts total icons needed including inputs, outputs, and separators.
 *
 * @param behavior - Behavior to analyze
 * @param behaviorType - Type of behavior (manual-action, triggered-effect, etc.)
 * @returns Layout requirements including total icons and separator positions
 */
export const calculateIconRequirements = (
  behavior: any,
  behaviorType: string,
): LayoutRequirement => {
  let totalIcons = 0;
  let separatorCount = 0;
  let separatorPositions: number[] = [];

  // Handle choice-based behaviors
  if (behavior.choices && behavior.choices.length > 0) {
    // For choices, we need to count the total icons across all choice options
    const choiceIcons = behavior.choices.reduce((sum: number, choice: any) => {
      let choiceSum = 0;

      if (choice.inputs && choice.inputs.length > 0) {
        choiceSum += choice.inputs.reduce((inputSum: number, input: any) => {
          const analysis = analyzeResourceDisplayWithConstraints(
            input,
            MAX_HORIZONTAL_ICONS,
            false,
          );
          return inputSum + analysis.iconCount;
        }, 0);
      }

      if (choice.outputs && choice.outputs.length > 0) {
        choiceSum += choice.outputs.reduce((outputSum: number, output: any) => {
          const analysis = analyzeResourceDisplayWithConstraints(
            output,
            MAX_HORIZONTAL_ICONS,
            false,
          );
          return outputSum + analysis.iconCount;
        }, 0);
      }

      // Add separator for choice (if it has both inputs and outputs)
      if (choice.inputs?.length > 0 && choice.outputs?.length > 0) {
        choiceSum += 1;
      }

      return sum + choiceSum;
    }, 0);

    totalIcons += choiceIcons;

    // Add separators between choices (but not after the last one)
    if (behavior.choices.length > 1) {
      separatorCount += behavior.choices.length - 1;
      totalIcons += behavior.choices.length - 1;
    }
  } else {
    // Regular behavior handling
    if (behavior.inputs && behavior.inputs.length > 0) {
      const inputIcons = behavior.inputs.reduce((sum: number, input: any) => {
        const analysis = analyzeResourceDisplayWithConstraints(
          input,
          MAX_HORIZONTAL_ICONS,
          false,
        );
        return sum + analysis.iconCount;
      }, 0);
      totalIcons += inputIcons;
    }

    if (
      behaviorType === "manual-action" &&
      behavior.inputs?.length > 0 &&
      behavior.outputs?.length > 0
    ) {
      separatorCount = 1;
      separatorPositions = [totalIcons];
      totalIcons += 1;
    } else if (
      behaviorType === "triggered-effect" ||
      behaviorType === "discount"
    ) {
      separatorCount = 1;
      separatorPositions = [totalIcons];
      totalIcons += 1;
    }

    if (behavior.outputs && behavior.outputs.length > 0) {
      const outputIcons = behavior.outputs.reduce(
        (sum: number, output: any) => {
          const analysis = analyzeResourceDisplayWithConstraints(
            output,
            MAX_HORIZONTAL_ICONS,
            false,
          );
          return sum + analysis.iconCount;
        },
        0,
      );
      totalIcons += outputIcons;
    }
  }

  return {
    totalIcons,
    separatorCount,
    separatorPositions,
    behaviorType,
    needsMultipleRows: totalIcons > MAX_HORIZONTAL_ICONS,
    maxHorizontalIcons: MAX_HORIZONTAL_ICONS,
  };
};

/**
 * Creates a layout plan for displaying a behavior's icons.
 * Determines if single or multi-row layout is needed.
 *
 * @param behavior - Behavior to layout
 * @param behaviorType - Type of behavior
 * @returns Layout plan with rows and separators
 */
export const createLayoutPlan = (
  behavior: any,
  behaviorType: string,
): LayoutPlan => {
  const requirements = calculateIconRequirements(behavior, behaviorType);

  if (!requirements.needsMultipleRows) {
    const row: IconDisplayInfo[] = [];

    // Handle choice-based behaviors
    if (behavior.choices && behavior.choices.length > 0) {
      behavior.choices.forEach((choice: any) => {
        if (choice.inputs) {
          choice.inputs.forEach((input: any) => {
            row.push(
              analyzeResourceDisplayWithConstraints(
                input,
                MAX_HORIZONTAL_ICONS,
                false,
              ),
            );
          });
        }

        // Add separator between inputs and outputs within each choice
        if (choice.inputs?.length > 0 && choice.outputs?.length > 0) {
          // This separator will be handled in rendering
        }

        if (choice.outputs) {
          choice.outputs.forEach((output: any) => {
            row.push(
              analyzeResourceDisplayWithConstraints(
                output,
                MAX_HORIZONTAL_ICONS,
                false,
              ),
            );
          });
        }
      });
    } else {
      // Regular behavior handling
      if (behavior.inputs) {
        behavior.inputs.forEach((input: any) => {
          row.push(
            analyzeResourceDisplayWithConstraints(
              input,
              MAX_HORIZONTAL_ICONS,
              false,
            ),
          );
        });
      }

      if (behavior.outputs) {
        behavior.outputs.forEach((output: any) => {
          row.push(
            analyzeResourceDisplayWithConstraints(
              output,
              MAX_HORIZONTAL_ICONS,
              false,
            ),
          );
        });
      }
    }

    const separators = requirements.separatorPositions.map((_pos) => ({
      position: 0,
      type: behaviorType === "manual-action" ? "→" : ":",
    }));

    return { rows: [row], separators, totalRows: 1 };
  } else {
    return createMultiRowLayout(behavior, behaviorType, requirements);
  }
};

/**
 * Creates a balanced multi-row layout when icons exceed horizontal space.
 * Distributes icons intelligently across rows.
 *
 * For manual actions: Balances inputs and outputs around separator.
 * For other types: Simple distribution across available space.
 *
 * @param behavior - Behavior to layout
 * @param behaviorType - Type of behavior
 * @param _requirements - Calculated layout requirements (unused but kept for signature)
 * @returns Multi-row layout plan
 */
export const createMultiRowLayout = (
  behavior: any,
  behaviorType: string,
  _requirements: LayoutRequirement,
): LayoutPlan => {
  const rows: IconDisplayInfo[][] = [];

  if (behaviorType === "manual-action") {
    // Special handling for actions: balance inputs and outputs around separator
    const inputs = behavior.inputs
      ? behavior.inputs.map((input: any) =>
          analyzeResourceDisplayWithConstraints(input, 3, false),
        )
      : [];
    const outputs = behavior.outputs
      ? behavior.outputs.map((output: any) =>
          analyzeResourceDisplayWithConstraints(output, 3, false),
        )
      : [];

    const inputRows = distributeIconsAcrossRows(
      inputs,
      Math.floor((MAX_HORIZONTAL_ICONS - 1) / 2),
    ); // Reserve space for separator
    const outputRows = distributeIconsAcrossRows(
      outputs,
      Math.floor((MAX_HORIZONTAL_ICONS - 1) / 2),
    );

    // Combine rows, ensuring equal number of rows on each side
    const maxRows = Math.max(inputRows.length, outputRows.length);
    for (let i = 0; i < maxRows; i++) {
      rows.push([...(inputRows[i] || []), ...(outputRows[i] || [])]);
    }

    const separators = [{ position: 0, type: "→" }]; // Separator in each row
    return { rows, separators, totalRows: rows.length };
  } else {
    // For other types, simple distribution
    const allItems: IconDisplayInfo[] = [];

    if (behavior.outputs) {
      behavior.outputs.forEach((output: any) => {
        allItems.push(
          analyzeResourceDisplayWithConstraints(
            output,
            MAX_HORIZONTAL_ICONS,
            false,
          ),
        );
      });
    }

    const distributedRows = distributeIconsAcrossRows(
      allItems,
      MAX_HORIZONTAL_ICONS,
    );
    const separators =
      behaviorType !== "auto-no-background" ? [{ position: 0, type: ":" }] : [];

    return {
      rows: distributedRows,
      separators,
      totalRows: distributedRows.length,
    };
  }
};

/**
 * Distributes icons across multiple rows with intelligent balancing.
 * Fills rows up to maxPerRow before starting a new row.
 *
 * @param items - Display info items to distribute
 * @param maxPerRow - Maximum icon count per row
 * @returns Array of rows, each containing display info items
 */
export const distributeIconsAcrossRows = (
  items: IconDisplayInfo[],
  maxPerRow: number,
): IconDisplayInfo[][] => {
  if (items.length === 0) return [];

  const totalIconCount = items.reduce((sum, item) => sum + item.iconCount, 0);

  if (totalIconCount <= maxPerRow) {
    return [items];
  }

  const rows: IconDisplayInfo[][] = [];
  let currentRow: IconDisplayInfo[] = [];
  let currentRowIconCount = 0;

  for (const item of items) {
    if (
      currentRowIconCount + item.iconCount > maxPerRow &&
      currentRow.length > 0
    ) {
      // Start new row
      rows.push(currentRow);
      currentRow = [item];
      currentRowIconCount = item.iconCount;
    } else {
      currentRow.push(item);
      currentRowIconCount += item.iconCount;
    }
  }

  if (currentRow.length > 0) {
    rows.push(currentRow);
  }

  return rows;
};
