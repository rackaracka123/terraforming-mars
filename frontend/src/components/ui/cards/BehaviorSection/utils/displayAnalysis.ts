import { IconDisplayInfo } from "../types.ts";

/**
 * Enhanced resource display analysis that considers space constraints.
 * Determines whether to display resources individually or with a number indicator.
 *
 * @param resource - Resource object with type and amount
 * @param availableSpace - Maximum number of icons that can fit horizontally
 * @param forceCompact - Whether to force compact display mode
 * @returns Display information including mode and icon count
 */
export const analyzeResourceDisplayWithConstraints = (
  resource: any,
  availableSpace: number,
  forceCompact: boolean = false,
): IconDisplayInfo => {
  const resourceType = resource.resourceType || resource.type || "unknown";
  const amount = resource.amount ?? 1;
  const hasPer = resource.per;
  const isProduction = resourceType?.includes("-production");

  // Per conditions count as 2 icons (production icon + per icon)
  if (isProduction && hasPer) {
    return {
      resourceType,
      amount,
      displayMode: "number", // Always use number format for per conditions
      iconCount: 2, // Production icon + per icon
    };
  }

  // Use individual display for amounts ≤3 (unless compact mode forces earlier threshold)
  const maxIndividualIcons = forceCompact || resource.forceNumberFormat ? 2 : 3;
  const absoluteAmount = Math.abs(amount);
  const useIndividual =
    absoluteAmount > 0 &&
    absoluteAmount <= maxIndividualIcons &&
    absoluteAmount <= availableSpace;

  return {
    resourceType,
    amount,
    displayMode: useIndividual ? "individual" : "number",
    iconCount: useIndividual ? absoluteAmount : 1,
  };
};

/**
 * Coordinates display modes across multiple resources for consistency.
 * If ANY resource uses "number + icon" format, ALL should use it (except amount=1).
 *
 * @param resources - Array of resources to coordinate
 * @returns Map of resources to their display information
 */
export const coordinateDisplayModes = (
  resources: any[],
): Map<any, IconDisplayInfo> => {
  // First pass: analyze each resource independently
  const displayInfos = resources.map((r) => ({
    resource: r,
    info: analyzeResourceDisplayWithConstraints(r, 7, false),
  }));

  // Check if ANY resource uses "number" mode
  const hasNumberMode = displayInfos.some(
    (d) => d.info.displayMode === "number",
  );

  // Second pass: if any uses number mode, force all to use it (except amount=1)
  if (hasNumberMode) {
    return new Map(
      displayInfos.map(({ resource, info }) => {
        const amount = Math.abs(resource.amount ?? 1);
        if (amount === 1) {
          // Keep individual mode for amount=1 (redundant to show "1")
          return [resource, info];
        } else {
          // Force number mode for consistency
          return [resource, { ...info, displayMode: "number", iconCount: 1 }];
        }
      }),
    );
  }

  // Otherwise, keep original display modes
  return new Map(displayInfos.map(({ resource, info }) => [resource, info]));
};
