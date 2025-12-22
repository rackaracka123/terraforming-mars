import React from "react";
import GameIcon from "../../../display/GameIcon.tsx";
import ResourceDisplay from "./ResourceDisplay.tsx";
import BehaviorIcon from "./BehaviorIcon.tsx";
import CardIcon from "./CardIcon.tsx";
import { getIconPath, getTagIconPath } from "@/utils/iconStore.ts";
import { analyzeCardOutputs } from "../utils/displayAnalysis.ts";

interface IconDisplayInfo {
  resourceType: string;
  amount: number;
  displayMode: "individual" | "number";
  iconCount: number;
}

interface LayoutPlan {
  rows: IconDisplayInfo[][];
  separators: Array<{ position: number; type: string }>;
  totalRows: number;
}

interface TileScaleInfo {
  scale: 1 | 1.25 | 1.5 | 2;
  tileType: string | null;
}

interface ImmediateResourceLayoutProps {
  behavior: any;
  layoutPlan: LayoutPlan;
  isResourceAffordable: (resource: any, isInput: boolean) => boolean;
  analyzeResourceDisplayWithConstraints: (
    resource: any,
    availableSpace: number,
    forceCompact: boolean,
  ) => IconDisplayInfo;
  tileScaleInfo: TileScaleInfo;
  renderIcon: (
    resourceType: string,
    isProduction: boolean,
    isAttack: boolean,
    context: "standalone" | "action" | "production" | "default",
    isAffordable: boolean,
  ) => React.ReactNode;
}

const ImmediateResourceLayout: React.FC<ImmediateResourceLayoutProps> = ({
  behavior,
  layoutPlan: _layoutPlan,
  isResourceAffordable,
  analyzeResourceDisplayWithConstraints,
  tileScaleInfo,
  renderIcon,
}) => {
  // Helper function to check if an output is a global parameter or tile placement
  const isGlobalParamOrTile = (output: any): boolean => {
    const type = output.resourceType || output.type || "";
    return (
      type === "temperature" ||
      type === "oxygen" ||
      type === "ocean" ||
      type === "venus" ||
      type === "tr" ||
      type === "city-tile" ||
      type === "city-placement" ||
      type === "ocean-tile" ||
      type === "ocean-placement" ||
      type === "greenery-tile" ||
      type === "greenery-placement"
    );
  };

  // Helper function to coordinate display modes across resources for consistency
  // If ANY resource uses "number + icon" format, ALL should use it (except amount=1)
  const coordinateDisplayModes = (resources: any[]): Map<any, IconDisplayInfo> => {
    // First pass: analyze each resource independently
    const displayInfos = resources.map((r) => ({
      resource: r,
      info: analyzeResourceDisplayWithConstraints(r, 7, false),
    }));

    // Check if ANY resource uses "number" mode
    const hasNumberMode = displayInfos.some((d) => d.info.displayMode === "number");

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

  // Helper function to render production group content
  const renderProductionGroup = (negative: any[], positive: any[]): React.ReactNode => {
    return (
      <div
        className={`flex flex-col gap-[3px] justify-center ${negative.length > 0 ? "items-start" : "items-center"}`}
      >
        {/* Negative production on first row */}
        {negative.length > 0 && (
          <div className="flex gap-[3px] items-center justify-start">
            <span className="text-xl font-bold text-[#ffcdd2] w-[20px] h-[26px] flex items-center justify-center [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)]">
              -
            </span>
            {negative.map((output: any, index: number) => {
              const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
              return (
                <React.Fragment key={`neg-prod-${index}`}>
                  <ResourceDisplay
                    displayInfo={displayInfo}
                    isInput={false}
                    resource={output}
                    isGroupedWithOtherNegatives={true}
                    context="standalone"
                    isAffordable={isResourceAffordable(output, false)}
                    tileScaleInfo={tileScaleInfo}
                  />
                </React.Fragment>
              );
            })}
          </div>
        )}

        {/* Positive production on second row */}
        {positive.length > 0 && (
          <>
            {negative.length === 0 && positive.length === 2 ? (
              positive.map((output: any, index: number) => {
                const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
                return (
                  <div
                    key={`pos-prod-row-${index}`}
                    className="flex gap-[3px] items-center justify-center"
                  >
                    <ResourceDisplay
                      displayInfo={displayInfo}
                      isInput={false}
                      resource={output}
                      isGroupedWithOtherNegatives={false}
                      context="standalone"
                      isAffordable={isResourceAffordable(output, false)}
                      tileScaleInfo={tileScaleInfo}
                    />
                  </div>
                );
              })
            ) : (
              <div className="flex gap-[3px] items-center justify-start">
                {negative.length > 0 && (
                  <span className="text-xl font-bold text-[#c8e6c9] w-[20px] h-[26px] flex items-center justify-center [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)]">
                    +
                  </span>
                )}
                {positive.map((output: any, index: number) => {
                  const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
                  return (
                    <React.Fragment key={`pos-prod-${index}`}>
                      <ResourceDisplay
                        displayInfo={displayInfo}
                        isInput={false}
                        resource={output}
                        isGroupedWithOtherNegatives={false}
                        context="standalone"
                        isAffordable={isResourceAffordable(output, false)}
                        tileScaleInfo={tileScaleInfo}
                      />
                    </React.Fragment>
                  );
                })}
              </div>
            )}
          </>
        )}
      </div>
    );
  };

  // Helper function to render non-production group content
  const renderNonProductionGroup = (negative: any[], positive: any[]): React.ReactNode => {
    return (
      <div
        className={`flex flex-col gap-[3px] justify-center ${negative.length > 0 && positive.length > 0 ? "items-start" : "items-center"}`}
      >
        {/* Negative resources on first row */}
        {negative.length > 0 && (
          <div className="flex gap-[3px] items-center justify-start">
            {negative.length > 1 && (
              <span className="text-xl font-bold text-[#ffcdd2] w-[20px] h-[26px] flex items-center justify-center [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)]">
                -
              </span>
            )}
            {negative.map((output: any, index: number) => {
              const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
              const isGrouped = negative.length > 1;
              return (
                <React.Fragment key={`neg-${index}`}>
                  <ResourceDisplay
                    displayInfo={displayInfo}
                    isInput={false}
                    resource={output}
                    isGroupedWithOtherNegatives={isGrouped}
                    context="standalone"
                    isAffordable={isResourceAffordable(output, false)}
                    tileScaleInfo={tileScaleInfo}
                  />
                </React.Fragment>
              );
            })}
          </div>
        )}

        {/* Positive resources on second row */}
        {positive.length > 0 && (
          <>
            {negative.length === 0 && positive.length === 2 ? (
              positive.map((output: any, index: number) => {
                const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
                return (
                  <div
                    key={`pos-row-${index}`}
                    className="flex gap-[3px] items-center justify-center"
                  >
                    <ResourceDisplay
                      displayInfo={displayInfo}
                      isInput={false}
                      resource={output}
                      isGroupedWithOtherNegatives={false}
                      context="standalone"
                      isAffordable={isResourceAffordable(output, false)}
                      tileScaleInfo={tileScaleInfo}
                    />
                  </div>
                );
              })
            ) : (
              <div className="flex gap-[3px] items-center justify-start">
                {negative.length > 0 && (
                  <span className="text-xl font-bold text-[#c8e6c9] w-[20px] h-[26px] flex items-center justify-center [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)]">
                    +
                  </span>
                )}
                {positive.map((output: any, index: number) => {
                  const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
                  return (
                    <React.Fragment key={`pos-${index}`}>
                      <ResourceDisplay
                        displayInfo={displayInfo}
                        isInput={false}
                        resource={output}
                        isGroupedWithOtherNegatives={false}
                        context="standalone"
                        isAffordable={isResourceAffordable(output, false)}
                        tileScaleInfo={tileScaleInfo}
                      />
                    </React.Fragment>
                  );
                })}
              </div>
            )}
          </>
        )}
      </div>
    );
  };

  // Check if this is a global-parameter-lenience effect (special case - no trigger display)
  const isGlobalParameterLenience =
    behavior.outputs?.some(
      (output: any) =>
        output.type === "global-parameter-lenience" ||
        output.resourceType === "global-parameter-lenience",
    ) ?? false;

  // Special case: if there are trigger conditions (e.g., Herbivores card), render condition icon : output icons
  // BUT skip this for global-parameter-lenience (it has its own display format)
  if (
    !isGlobalParameterLenience &&
    behavior.triggers &&
    behavior.triggers.length > 0 &&
    behavior.triggers.some((trigger: any) => trigger.condition) &&
    behavior.outputs &&
    behavior.outputs.length > 0
  ) {
    return (
      <div className="flex gap-[3px] items-center justify-center">
        {/* Render trigger condition icon(s) */}
        {behavior.triggers
          .filter((trigger: any) => trigger.condition)
          .map((trigger: any, triggerIndex: number) => {
            // For card-played triggers with affectedTags, render the tags instead of the condition type
            const condition = trigger.condition;
            if (
              condition.type === "card-played" &&
              condition.affectedTags &&
              condition.affectedTags.length > 0
            ) {
              return (
                <React.Fragment key={`trigger-condition-${triggerIndex}`}>
                  {condition.affectedTags.map((tag: string, tagIndex: number) => (
                    <React.Fragment key={`trigger-tag-${triggerIndex}-${tagIndex}`}>
                      {tagIndex > 0 && (
                        <span className="text-white font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                          /
                        </span>
                      )}
                      <BehaviorIcon
                        resourceType={tag}
                        isProduction={false}
                        isAttack={false}
                        context="standalone"
                        isAffordable={true}
                        tileScaleInfo={tileScaleInfo}
                      />
                    </React.Fragment>
                  ))}
                </React.Fragment>
              );
            } else {
              // Default: render the condition type as icon
              return (
                <React.Fragment key={`trigger-condition-${triggerIndex}`}>
                  <BehaviorIcon
                    resourceType={condition.type}
                    isProduction={false}
                    isAttack={false}
                    context="standalone"
                    isAffordable={true}
                    tileScaleInfo={tileScaleInfo}
                  />
                </React.Fragment>
              );
            }
          })}
        {/* Colon separator */}
        <span className="flex items-center justify-center text-white text-base font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] min-w-[20px] z-[1]">
          :
        </span>
        {/* Render output icons */}
        {behavior.outputs.map((output: any, outputIndex: number) => {
          const displayInfo = analyzeResourceDisplayWithConstraints(output, 6, false);
          return (
            <React.Fragment key={`trigger-output-${outputIndex}`}>
              <ResourceDisplay
                displayInfo={displayInfo}
                isInput={false}
                resource={output}
                isGroupedWithOtherNegatives={false}
                context="standalone"
                isAffordable={isResourceAffordable(output, false)}
                tileScaleInfo={tileScaleInfo}
              />
            </React.Fragment>
          );
        })}
      </div>
    );
  }

  // Special case: if there are choices AND outputs, render them on separate rows
  if (
    behavior.choices &&
    behavior.choices.length > 0 &&
    behavior.outputs &&
    behavior.outputs.length > 0
  ) {
    return (
      <div className="flex flex-col gap-[6px] items-center justify-start w-full py-1">
        {/* Render choices in compact format: amount <icon> / amount <icon> / ... */}
        <div className="flex flex-wrap gap-1 items-center justify-center">
          {behavior.choices.map((choice: any, choiceIndex: number) => (
            <React.Fragment key={`choice-compact-${choiceIndex}`}>
              {choiceIndex > 0 && (
                <span className="text-white font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                  /
                </span>
              )}
              {choice.outputs &&
                choice.outputs.map((output: any, outputIndex: number) => {
                  const amount = output.amount || 1;
                  const resourceType = output.resourceType || output.type;
                  const isAffordable = isResourceAffordable(output, false);

                  return (
                    <React.Fragment key={`choice-${choiceIndex}-output-${outputIndex}`}>
                      <div className="flex gap-[3px] items-center">
                        {amount > 0 && (
                          <span className="text-white font-bold text-base [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                            {amount}
                          </span>
                        )}
                        <BehaviorIcon
                          resourceType={resourceType}
                          isProduction={false}
                          isAttack={false}
                          context="standalone"
                          isAffordable={isAffordable}
                          tileScaleInfo={tileScaleInfo}
                        />
                      </div>
                    </React.Fragment>
                  );
                })}
            </React.Fragment>
          ))}
        </div>

        {/* Render regular outputs on a new row */}
        {(() => {
          // Check if outputs are production outputs
          const productionOutputs = behavior.outputs.filter(
            (output: any) =>
              output.resourceType?.includes("production") ||
              output.type?.includes("production") ||
              output.isProduction === true,
          );
          const nonProductionOutputs = behavior.outputs.filter(
            (output: any) =>
              !(
                output.resourceType?.includes("production") ||
                output.type?.includes("production") ||
                output.isProduction === true
              ),
          );

          // If all outputs are production, wrap them in brown box with row separation
          if (productionOutputs.length > 0 && nonProductionOutputs.length === 0) {
            const negativeProduction = productionOutputs.filter(
              (output: any) => (output.amount || 0) < 0,
            );
            const positiveProduction = productionOutputs.filter(
              (output: any) => (output.amount || 0) > 0,
            );

            return (
              <div className="flex flex-wrap gap-[3px] items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-1.5 py-[3px] shadow-[0_1px_3px_rgba(0,0,0,0.2)]">
                <div
                  className={`flex flex-col gap-[3px] justify-center ${negativeProduction.length > 0 ? "items-start" : "items-center"}`}
                >
                  {/* Negative production on first row */}
                  {negativeProduction.length > 0 && (
                    <div className="flex gap-[3px] items-center">
                      <span className="text-white font-bold text-base [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                        -
                      </span>
                      {negativeProduction.map((output: any, index: number) => {
                        const displayInfo = analyzeResourceDisplayWithConstraints(
                          { ...output, amount: Math.abs(output.amount) },
                          7,
                          false,
                        );
                        return (
                          <React.Fragment key={`neg-prod-${index}`}>
                            <ResourceDisplay
                              displayInfo={displayInfo}
                              isInput={false}
                              resource={{
                                ...output,
                                amount: Math.abs(output.amount),
                              }}
                              isGroupedWithOtherNegatives={false}
                              context="standalone"
                              isAffordable={isResourceAffordable(output, false)}
                              tileScaleInfo={tileScaleInfo}
                            />
                          </React.Fragment>
                        );
                      })}
                    </div>
                  )}
                  {/* Positive production on second row */}
                  {positiveProduction.length > 0 && (
                    <div className="flex gap-[3px] items-center">
                      {negativeProduction.length > 0 && (
                        <span className="text-white font-bold text-base [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                          +
                        </span>
                      )}
                      {positiveProduction.map((output: any, index: number) => {
                        const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
                        return (
                          <React.Fragment key={`pos-prod-${index}`}>
                            <ResourceDisplay
                              displayInfo={displayInfo}
                              isInput={false}
                              resource={output}
                              isGroupedWithOtherNegatives={false}
                              context="standalone"
                              isAffordable={isResourceAffordable(output, false)}
                              tileScaleInfo={tileScaleInfo}
                            />
                          </React.Fragment>
                        );
                      })}
                    </div>
                  )}
                </div>
              </div>
            );
          }

          // Otherwise render outputs normally
          return (
            <div className="flex flex-wrap gap-[3px] items-center justify-center">
              {behavior.outputs.map((output: any, index: number) => {
                const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
                return (
                  <React.Fragment key={`output-${index}`}>
                    <ResourceDisplay
                      displayInfo={displayInfo}
                      isInput={false}
                      resource={output}
                      isGroupedWithOtherNegatives={false}
                      context="standalone"
                      isAffordable={isResourceAffordable(output, false)}
                      tileScaleInfo={tileScaleInfo}
                    />
                  </React.Fragment>
                );
              })}
            </div>
          );
        })()}
      </div>
    );
  }

  // Special case: choices with only production outputs and no behavior-level outputs
  // Display in a single brown production box with OR separators
  if (
    (!behavior.outputs || behavior.outputs.length === 0) &&
    behavior.choices &&
    behavior.choices.length > 0
  ) {
    // Check if all choices contain only production outputs
    const allChoicesAreProduction = behavior.choices.every((choice: any) => {
      if (!choice.outputs || choice.outputs.length === 0) return false;
      return choice.outputs.every(
        (output: any) =>
          output.type?.includes("production") || output.resourceType?.includes("production"),
      );
    });

    if (allChoicesAreProduction) {
      return (
        <div className="flex flex-wrap gap-[3px] items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-1.5 py-[3px] shadow-[0_1px_3px_rgba(0,0,0,0.2)]">
          <div className="flex items-center gap-2">
            {behavior.choices.map((choice: any, choiceIndex: number) => (
              <React.Fragment key={`prod-choice-${choiceIndex}`}>
                {choiceIndex > 0 && (
                  <div className="text-xs font-semibold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] my-0.5 mx-1 bg-[rgba(139,89,42,0.6)] py-0.5 px-1.5 rounded-[2px] backdrop-blur-[2px]">
                    OR
                  </div>
                )}
                <div className="flex gap-[3px] items-center">
                  {choice.outputs.map((output: any, outputIndex: number) => {
                    const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
                    return (
                      <React.Fragment key={`prod-choice-${choiceIndex}-output-${outputIndex}`}>
                        <ResourceDisplay
                          displayInfo={displayInfo}
                          isInput={false}
                          resource={output}
                          isGroupedWithOtherNegatives={false}
                          context="standalone"
                          isAffordable={isResourceAffordable(output, false)}
                          tileScaleInfo={tileScaleInfo}
                        />
                      </React.Fragment>
                    );
                  })}
                </div>
              </React.Fragment>
            ))}
          </div>
        </div>
      );
    }
  }

  if (!behavior.outputs || behavior.outputs.length === 0) return null;

  // Analyze and consolidate card outputs (card-draw, card-peek, card-take, card-buy)
  const consolidatedCards = analyzeCardOutputs(behavior.outputs);

  // Helper to check if an output is a card resource
  const isCardResource = (output: any): boolean => {
    const type = output.resourceType || output.type || "";
    return (
      type === "card-draw" || type === "card-peek" || type === "card-take" || type === "card-buy"
    );
  };

  // Separate production and non-production outputs
  const productionOutputs = behavior.outputs.filter(
    (output: any) =>
      output.resourceType?.includes("production") ||
      output.type?.includes("production") ||
      output.isProduction === true,
  );
  // Filter out card resources from non-production outputs (they'll be rendered separately)
  const nonProductionOutputs = behavior.outputs.filter(
    (output: any) =>
      !(
        output.resourceType?.includes("production") ||
        output.type?.includes("production") ||
        output.isProduction === true
      ) && !isCardResource(output),
  );

  // Separate per-condition production (which already has its own wrapper) from regular production
  const perConditionProduction = productionOutputs.filter((output: any) => output.per);
  const regularProduction = productionOutputs.filter((output: any) => !output.per);

  // Separate negative and positive production outputs (only for regular production, not per-condition)
  const negativeProduction = regularProduction.filter((output: any) => (output.amount ?? 1) < 0);
  const positiveProduction = regularProduction.filter((output: any) => (output.amount ?? 1) >= 0);

  // Separate negative and positive non-production outputs
  const negativeOutputs = nonProductionOutputs.filter((output: any) => (output.amount ?? 1) < 0);
  const positiveOutputs = nonProductionOutputs.filter((output: any) => (output.amount ?? 1) >= 0);

  // Special handling: if nonProductionOutputs has both regular resources AND global params/tiles,
  // and there are at least 3 outputs total, use special layouts
  const globalParamOutputs = nonProductionOutputs.filter(isGlobalParamOrTile);
  const regularResourceOutputs = nonProductionOutputs.filter(
    (output: any) => !isGlobalParamOrTile(output),
  );

  const hasGlobalParamsOrTiles = globalParamOutputs.length > 0;
  const hasRegularResources = regularResourceOutputs.length > 0;
  const shouldUseTwoColumnLayout =
    nonProductionOutputs.length >= 3 &&
    hasGlobalParamsOrTiles &&
    hasRegularResources &&
    regularProduction.length === 0 &&
    perConditionProduction.length === 0;

  // Special case: if there are 1+ global params/tiles and 1+ regular resources,
  // stack them vertically (resources on top, global params/tiles on bottom)
  const shouldUseTwoRowLayout =
    shouldUseTwoColumnLayout &&
    globalParamOutputs.length >= 1 &&
    regularResourceOutputs.length >= 1;

  if (shouldUseTwoRowLayout) {
    // Split regular resources into attacks and non-attacks
    const attackResources = regularResourceOutputs.filter(
      (output: any) => output.target === "any-player",
    );
    const positiveRegular = regularResourceOutputs.filter(
      (output: any) => output.target !== "any-player" && (output.amount ?? 1) >= 0,
    );
    const negativeRegular = regularResourceOutputs.filter(
      (output: any) => output.target !== "any-player" && (output.amount ?? 1) < 0,
    );

    // Coordinate display modes for consistency across all regular resources
    const regularDisplayModes = coordinateDisplayModes([
      ...attackResources,
      ...negativeRegular,
      ...positiveRegular,
    ]);

    return (
      <div className="flex flex-col gap-[9px] items-center justify-center max-w-full">
        {/* Top row: regular resources */}
        <div className="flex gap-[3px] items-center justify-center">
          {attackResources.map((output: any, index: number) => {
            const displayInfo = regularDisplayModes.get(output)!;
            return (
              <React.Fragment key={`attack-${index}`}>
                <ResourceDisplay
                  displayInfo={displayInfo}
                  isInput={false}
                  resource={output}
                  isGroupedWithOtherNegatives={false}
                  context="standalone"
                  isAffordable={isResourceAffordable(output, false)}
                  tileScaleInfo={tileScaleInfo}
                />
              </React.Fragment>
            );
          })}
          {negativeRegular.length > 0 && (
            <>
              {negativeRegular.length > 1 && (
                <span className="text-xl font-bold text-[#ffcdd2] w-[20px] h-[26px] flex items-center justify-center [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)]">
                  -
                </span>
              )}
              {negativeRegular.map((output: any, index: number) => {
                const displayInfo = regularDisplayModes.get(output)!;
                const isGrouped = negativeRegular.length > 1;
                return (
                  <React.Fragment key={`neg-reg-${index}`}>
                    <ResourceDisplay
                      displayInfo={displayInfo}
                      isInput={false}
                      resource={output}
                      isGroupedWithOtherNegatives={isGrouped}
                      context="standalone"
                      isAffordable={isResourceAffordable(output, false)}
                      tileScaleInfo={tileScaleInfo}
                    />
                  </React.Fragment>
                );
              })}
            </>
          )}
          {positiveRegular.map((output: any, index: number) => {
            const displayInfo = regularDisplayModes.get(output)!;
            return (
              <React.Fragment key={`pos-reg-${index}`}>
                <ResourceDisplay
                  displayInfo={displayInfo}
                  isInput={false}
                  resource={output}
                  isGroupedWithOtherNegatives={false}
                  context="standalone"
                  isAffordable={isResourceAffordable(output, false)}
                  tileScaleInfo={tileScaleInfo}
                />
              </React.Fragment>
            );
          })}
        </div>

        {/* Bottom row: global parameters and tiles */}
        <div className="flex gap-[3px] items-center justify-center">
          {[...globalParamOutputs]
            .sort((a, b) => {
              // TR should appear last in global param group
              const typeA = a.resourceType || a.type || "";
              const typeB = b.resourceType || b.type || "";
              if (typeA === "tr") return 1;
              if (typeB === "tr") return -1;
              return 0;
            })
            .map((output: any, index: number) => {
              const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
              return (
                <React.Fragment key={`global-${index}`}>
                  <ResourceDisplay
                    displayInfo={displayInfo}
                    isInput={false}
                    resource={output}
                    isGroupedWithOtherNegatives={false}
                    context="standalone"
                    isAffordable={isResourceAffordable(output, false)}
                    tileScaleInfo={tileScaleInfo}
                  />
                </React.Fragment>
              );
            })}
        </div>
      </div>
    );
  }

  if (shouldUseTwoColumnLayout) {
    // Further split regular resources into attacks and non-attacks
    // Attacks (any-player target) are displayed first, then regular positive resources
    const attackResources = regularResourceOutputs.filter(
      (output: any) => output.target === "any-player",
    );
    const positiveRegular = regularResourceOutputs.filter(
      (output: any) => output.target !== "any-player" && (output.amount ?? 1) >= 0,
    );
    const negativeRegular = regularResourceOutputs.filter(
      (output: any) => output.target !== "any-player" && (output.amount ?? 1) < 0,
    );

    // Coordinate display modes for consistency across all regular resources
    const regularDisplayModes = coordinateDisplayModes([
      ...attackResources,
      ...negativeRegular,
      ...positiveRegular,
    ]);

    return (
      <div className="flex gap-2 items-center justify-center max-w-full">
        {/* Left side: regular resources in rows */}
        <div className="flex flex-col gap-[6px] items-center justify-center">
          {/* Attack resources (any-player) on first row */}
          {attackResources.length > 0 && (
            <div className="flex gap-[3px] items-center justify-center">
              {attackResources.map((output: any, index: number) => {
                const displayInfo = regularDisplayModes.get(output)!;
                return (
                  <React.Fragment key={`attack-${index}`}>
                    <ResourceDisplay
                      displayInfo={displayInfo}
                      isInput={false}
                      resource={output}
                      isGroupedWithOtherNegatives={false}
                      context="standalone"
                      isAffordable={isResourceAffordable(output, false)}
                      tileScaleInfo={tileScaleInfo}
                    />
                  </React.Fragment>
                );
              })}
            </div>
          )}
          {/* Negative resources */}
          {negativeRegular.length > 0 && (
            <div className="flex gap-[3px] items-center justify-center">
              {negativeRegular.length > 1 && (
                <span className="text-xl font-bold text-[#ffcdd2] w-[20px] h-[26px] flex items-center justify-center [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)]">
                  -
                </span>
              )}
              {negativeRegular.map((output: any, index: number) => {
                const displayInfo = regularDisplayModes.get(output)!;
                const isGrouped = negativeRegular.length > 1;
                return (
                  <React.Fragment key={`neg-reg-${index}`}>
                    <ResourceDisplay
                      displayInfo={displayInfo}
                      isInput={false}
                      resource={output}
                      isGroupedWithOtherNegatives={isGrouped}
                      context="standalone"
                      isAffordable={isResourceAffordable(output, false)}
                      tileScaleInfo={tileScaleInfo}
                    />
                  </React.Fragment>
                );
              })}
            </div>
          )}
          {/* Positive resources */}
          {positiveRegular.length > 0 && (
            <div className="flex gap-[3px] items-center justify-center">
              {positiveRegular.map((output: any, index: number) => {
                const displayInfo = regularDisplayModes.get(output)!;
                return (
                  <React.Fragment key={`pos-reg-${index}`}>
                    <ResourceDisplay
                      displayInfo={displayInfo}
                      isInput={false}
                      resource={output}
                      isGroupedWithOtherNegatives={false}
                      context="standalone"
                      isAffordable={isResourceAffordable(output, false)}
                      tileScaleInfo={tileScaleInfo}
                    />
                  </React.Fragment>
                );
              })}
            </div>
          )}
        </div>

        {/* Right side: global parameters and tiles in a single row */}
        <div className="flex gap-[3px] items-center justify-center">
          {[...globalParamOutputs]
            .sort((a, b) => {
              // TR should appear last in global param group
              const typeA = a.resourceType || a.type || "";
              const typeB = b.resourceType || b.type || "";
              if (typeA === "tr") return 1;
              if (typeB === "tr") return -1;
              return 0;
            })
            .map((output: any, index: number) => {
              const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
              return (
                <React.Fragment key={`global-${index}`}>
                  <ResourceDisplay
                    displayInfo={displayInfo}
                    isInput={false}
                    resource={output}
                    isGroupedWithOtherNegatives={false}
                    context="standalone"
                    isAffordable={isResourceAffordable(output, false)}
                    tileScaleInfo={tileScaleInfo}
                  />
                </React.Fragment>
              );
            })}
        </div>
      </div>
    );
  }

  // Count groups that have content
  const groups = [
    { content: regularProduction, hasGlobalParamOrTile: false },
    {
      content: perConditionProduction,
      hasGlobalParamOrTile: perConditionProduction.some(isGlobalParamOrTile),
    },
    {
      content: nonProductionOutputs,
      hasGlobalParamOrTile: nonProductionOutputs.some(isGlobalParamOrTile),
    },
  ].filter((group) => group.content.length > 0);

  // Special layout for 3 groups: 2 left (vertical), 1 right (prioritize global params/tiles on right)
  if (groups.length === 3) {
    // Determine which group goes on the right (prioritize global parameters and tiles)
    let rightGroupIndex = groups.findIndex((g) => g.hasGlobalParamOrTile);
    if (rightGroupIndex === -1) rightGroupIndex = 2; // Default to last group

    const leftGroups = groups.filter((_, i) => i !== rightGroupIndex);
    const rightGroup = groups[rightGroupIndex];

    return (
      <div className="flex gap-2 items-center justify-center max-w-full">
        {/* Left side: 2 groups vertically stacked */}
        <div className="flex flex-col gap-[3px] items-center justify-center">
          {leftGroups.map((group, index) => {
            if (group.content === regularProduction) {
              return (
                <div
                  key={`left-prod-${index}`}
                  className="flex flex-wrap gap-[3px] items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-1.5 py-[3px] shadow-[0_1px_3px_rgba(0,0,0,0.2)]"
                >
                  {renderProductionGroup(negativeProduction, positiveProduction)}
                </div>
              );
            } else if (group.content === perConditionProduction) {
              return (
                <div
                  key={`left-per-${index}`}
                  className="flex flex-wrap gap-[3px] items-center justify-center"
                >
                  {perConditionProduction.map((output: any, idx: number) => {
                    const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
                    return (
                      <React.Fragment key={`per-prod-left-${idx}`}>
                        <ResourceDisplay
                          displayInfo={displayInfo}
                          isInput={false}
                          resource={output}
                          isGroupedWithOtherNegatives={false}
                          context="standalone"
                          isAffordable={isResourceAffordable(output, false)}
                          tileScaleInfo={tileScaleInfo}
                        />
                      </React.Fragment>
                    );
                  })}
                </div>
              );
            } else {
              return (
                <div
                  key={`left-nonprod-${index}`}
                  className="flex flex-wrap gap-[3px] items-center justify-center"
                >
                  {renderNonProductionGroup(negativeOutputs, positiveOutputs)}
                </div>
              );
            }
          })}
        </div>

        {/* Right side: 1 group */}
        <div className="flex items-center justify-center">
          {rightGroup.content === regularProduction ? (
            <div className="flex flex-wrap gap-[3px] items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-1.5 py-[3px] shadow-[0_1px_3px_rgba(0,0,0,0.2)]">
              {renderProductionGroup(negativeProduction, positiveProduction)}
            </div>
          ) : rightGroup.content === perConditionProduction ? (
            <div className="flex flex-wrap gap-[3px] items-center justify-center">
              {perConditionProduction.map((output: any, idx: number) => {
                const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
                return (
                  <React.Fragment key={`per-prod-right-${idx}`}>
                    <ResourceDisplay
                      displayInfo={displayInfo}
                      isInput={false}
                      resource={output}
                      isGroupedWithOtherNegatives={false}
                      context="standalone"
                      isAffordable={isResourceAffordable(output, false)}
                      tileScaleInfo={tileScaleInfo}
                    />
                  </React.Fragment>
                );
              })}
            </div>
          ) : (
            <div className="flex flex-wrap gap-[3px] items-center justify-center">
              {renderNonProductionGroup(negativeOutputs, positiveOutputs)}
            </div>
          )}
        </div>
      </div>
    );
  }

  // Check if we have both regular and per-condition production - if so, combine them in ONE brown box
  const hasAllProductionTypes = regularProduction.length > 0 && perConditionProduction.length > 0;

  return (
    <div className="flex flex-wrap gap-2 items-center justify-center max-w-full">
      {/* If we have both types, wrap them together in ONE brown box */}
      {hasAllProductionTypes ? (
        <div className="flex flex-wrap gap-[3px] items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-1.5 py-[3px] shadow-[0_1px_3px_rgba(0,0,0,0.2)]">
          <div
            className={`flex flex-col gap-[3px] justify-center ${negativeProduction.length > 0 ? "items-start" : "items-center"}`}
          >
            {/* Negative production on first row */}
            {negativeProduction.length > 0 && (
              <div className="flex gap-[3px] items-center justify-start">
                <span className="text-xl font-bold text-[#ffcdd2] w-[20px] h-[26px] flex items-center justify-center [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)]">
                  -
                </span>
                {negativeProduction.map((output: any, index: number) => {
                  const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
                  const isGrouped = true;
                  return (
                    <React.Fragment key={`neg-prod-${index}`}>
                      <ResourceDisplay
                        displayInfo={displayInfo}
                        isInput={false}
                        resource={output}
                        isGroupedWithOtherNegatives={isGrouped}
                        context="standalone"
                        isAffordable={isResourceAffordable(output, false)}
                        tileScaleInfo={tileScaleInfo}
                      />
                    </React.Fragment>
                  );
                })}
              </div>
            )}

            {/* Positive regular production */}
            {positiveProduction.length > 0 &&
              positiveProduction.map((output: any, index: number) => {
                const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
                return (
                  <div
                    key={`pos-prod-row-${index}`}
                    className="flex gap-[3px] items-center justify-start"
                  >
                    {index === 0 && negativeProduction.length > 0 && (
                      <span className="text-xl font-bold text-[#c8e6c9] w-[20px] h-[26px] flex items-center justify-center [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)]">
                        +
                      </span>
                    )}
                    <ResourceDisplay
                      displayInfo={displayInfo}
                      isInput={false}
                      resource={output}
                      isGroupedWithOtherNegatives={false}
                      context="standalone"
                      isAffordable={isResourceAffordable(output, false)}
                      tileScaleInfo={tileScaleInfo}
                    />
                  </div>
                );
              })}

            {/* Per-condition production (render without wrapper since we're already in the brown box) */}
            {perConditionProduction.map((output: any, index: number) => {
              const baseResourceType = output.type.replace("-production", "");
              const hasPer = output.per;

              let perIcon = null;
              if (hasPer.tag) {
                perIcon = getTagIconPath(hasPer.tag);
              } else if (hasPer.type) {
                perIcon = getIconPath(hasPer.type);
              }

              if (baseResourceType === "credit") {
                return (
                  <div
                    key={`per-prod-${index}`}
                    className="flex gap-[3px] items-center justify-center"
                  >
                    <GameIcon
                      iconType="credit"
                      amount={Math.abs(output.amount ?? 1)}
                      size="small"
                    />
                    <span className="text-base font-bold text-white mx-[3px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
                      /
                    </span>
                    <img
                      src={perIcon!}
                      alt={hasPer.tag || hasPer.type}
                      className="w-[26px] h-[26px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] max-md:w-[22px] max-md:h-[22px]"
                    />
                  </div>
                );
              } else {
                const productionIcon = renderIcon(
                  baseResourceType,
                  false,
                  false,
                  "production",
                  true,
                );
                return (
                  <div
                    key={`per-prod-${index}`}
                    className="flex gap-[3px] items-center justify-center"
                  >
                    <div className="flex items-center gap-px relative">
                      {(output.amount ?? 1) > 1 && (
                        <span className="text-lg font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)] leading-none flex items-center ml-0.5 max-md:text-xs">
                          {output.amount}
                        </span>
                      )}
                      {productionIcon}
                    </div>
                    <span className="text-base font-bold text-white mx-[3px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
                      /
                    </span>
                    <img
                      src={perIcon!}
                      alt={hasPer.tag || hasPer.type}
                      className="w-[26px] h-[26px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] max-md:w-[22px] max-md:h-[22px]"
                    />
                  </div>
                );
              }
            })}
          </div>
        </div>
      ) : (
        <>
          {/* Regular production only (original logic) */}
          {regularProduction.length > 0 && (
            <div className="flex flex-wrap gap-[3px] items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-1.5 py-[3px] shadow-[0_1px_3px_rgba(0,0,0,0.2)]">
              <div
                className={`flex flex-col gap-[3px] justify-center ${negativeProduction.length > 0 ? "items-start" : "items-center"}`}
              >
                {negativeProduction.length > 0 && (
                  <div className="flex gap-[3px] items-center justify-start">
                    <span className="text-xl font-bold text-[#ffcdd2] w-[20px] h-[26px] flex items-center justify-center [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)]">
                      -
                    </span>
                    {negativeProduction.map((output: any, index: number) => {
                      const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
                      const isGrouped = true;
                      return (
                        <React.Fragment key={`neg-prod-${index}`}>
                          <ResourceDisplay
                            displayInfo={displayInfo}
                            isInput={false}
                            resource={output}
                            isGroupedWithOtherNegatives={isGrouped}
                            context="standalone"
                            isAffordable={isResourceAffordable(output, false)}
                            tileScaleInfo={tileScaleInfo}
                          />
                        </React.Fragment>
                      );
                    })}
                  </div>
                )}

                {positiveProduction.length > 0 && (
                  <>
                    {negativeProduction.length === 0 && positiveProduction.length === 2 ? (
                      positiveProduction.map((output: any, index: number) => {
                        const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
                        return (
                          <div
                            key={`pos-prod-row-${index}`}
                            className="flex gap-[3px] items-center justify-center"
                          >
                            <ResourceDisplay
                              displayInfo={displayInfo}
                              isInput={false}
                              resource={output}
                              isGroupedWithOtherNegatives={false}
                              context="standalone"
                              isAffordable={isResourceAffordable(output, false)}
                              tileScaleInfo={tileScaleInfo}
                            />
                          </div>
                        );
                      })
                    ) : (
                      <div className="flex gap-[3px] items-center justify-start flex-wrap">
                        {negativeProduction.length > 0 && (
                          <span className="text-xl font-bold text-[#c8e6c9] w-[20px] h-[26px] flex items-center justify-center [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)]">
                            +
                          </span>
                        )}
                        {positiveProduction.map((output: any, index: number) => {
                          const displayInfo = analyzeResourceDisplayWithConstraints(
                            output,
                            7,
                            false,
                          );
                          return (
                            <React.Fragment key={`pos-prod-${index}`}>
                              <ResourceDisplay
                                displayInfo={displayInfo}
                                isInput={false}
                                resource={output}
                                isGroupedWithOtherNegatives={false}
                                context="standalone"
                                isAffordable={isResourceAffordable(output, false)}
                                tileScaleInfo={tileScaleInfo}
                              />
                            </React.Fragment>
                          );
                        })}
                      </div>
                    )}
                  </>
                )}
              </div>
            </div>
          )}

          {/* Per-condition production only (with wrapper from renderResourceFromDisplayInfo) */}
          {perConditionProduction.length > 0 && (
            <div className="flex flex-col gap-[3px] items-center justify-center">
              {perConditionProduction.map((output: any, index: number) => {
                const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
                return (
                  <React.Fragment key={`per-prod-${index}`}>
                    <ResourceDisplay
                      displayInfo={displayInfo}
                      isInput={false}
                      resource={output}
                      isGroupedWithOtherNegatives={false}
                      context="standalone"
                      isAffordable={isResourceAffordable(output, false)}
                      tileScaleInfo={tileScaleInfo}
                    />
                  </React.Fragment>
                );
              })}
            </div>
          )}
        </>
      )}

      {/* Non-production outputs */}
      {(nonProductionOutputs.length > 0 || consolidatedCards.length > 0) && (
        <div
          className={`flex flex-col gap-[3px] justify-center ${negativeOutputs.length > 0 && positiveOutputs.length > 0 ? "items-start" : "items-center"}`}
        >
          {/* Negative resources on first row */}
          {negativeOutputs.length > 0 && (
            <div className="flex gap-[3px] items-center justify-start">
              {negativeOutputs.length > 1 && (
                <span className="text-xl font-bold text-[#ffcdd2] w-[20px] h-[26px] flex items-center justify-center [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)]">
                  -
                </span>
              )}
              {negativeOutputs.map((output: any, index: number) => {
                const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
                const isGrouped = negativeOutputs.length > 1;
                return (
                  <React.Fragment key={`neg-${index}`}>
                    <ResourceDisplay
                      displayInfo={displayInfo}
                      isInput={false}
                      resource={output}
                      isGroupedWithOtherNegatives={isGrouped}
                      context="standalone"
                      isAffordable={isResourceAffordable(output, false)}
                      tileScaleInfo={tileScaleInfo}
                    />
                  </React.Fragment>
                );
              })}
            </div>
          )}

          {/* Positive resources on second row */}
          {positiveOutputs.length > 0 && (
            <>
              {negativeOutputs.length === 0 && positiveOutputs.length === 2 ? (
                // When there are exactly 2 positive outputs and no negatives, show them on separate rows
                positiveOutputs.map((output: any, index: number) => {
                  const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
                  return (
                    <div
                      key={`pos-row-${index}`}
                      className="flex gap-[3px] items-center justify-center"
                    >
                      <ResourceDisplay
                        displayInfo={displayInfo}
                        isInput={false}
                        resource={output}
                        isGroupedWithOtherNegatives={false}
                        context="standalone"
                        isAffordable={isResourceAffordable(output, false)}
                        tileScaleInfo={tileScaleInfo}
                      />
                    </div>
                  );
                })
              ) : (
                // Default: all positive outputs in one row
                <div className="flex gap-[3px] items-center justify-start">
                  {negativeOutputs.length > 0 && (
                    <span className="text-xl font-bold text-[#c8e6c9] w-[20px] h-[26px] flex items-center justify-center [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)]">
                      +
                    </span>
                  )}
                  {positiveOutputs.map((output: any, index: number) => {
                    const displayInfo = analyzeResourceDisplayWithConstraints(output, 7, false);
                    return (
                      <React.Fragment key={`pos-${index}`}>
                        <ResourceDisplay
                          displayInfo={displayInfo}
                          isInput={false}
                          resource={output}
                          isGroupedWithOtherNegatives={false}
                          context="standalone"
                          isAffordable={isResourceAffordable(output, false)}
                          tileScaleInfo={tileScaleInfo}
                        />
                      </React.Fragment>
                    );
                  })}
                </div>
              )}
            </>
          )}

          {/* Consolidated card icons (card-draw, card-peek, card-take, card-buy) */}
          {consolidatedCards.length > 0 && (
            <div className="flex gap-[3px] items-center justify-start">
              {consolidatedCards.map((cardItem, index) => (
                <React.Fragment key={`card-${index}`}>
                  <CardIcon
                    amount={cardItem.amount}
                    badgeType={cardItem.badgeType}
                    isAffordable={true}
                  />
                </React.Fragment>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default ImmediateResourceLayout;
