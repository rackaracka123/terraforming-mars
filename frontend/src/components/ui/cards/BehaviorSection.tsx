import React from "react";
import { CardBehaviorDto, ResourcesDto } from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import { getIconPath, isTagIcon } from "@/utils/iconStore.ts";

interface BehaviorSectionProps {
  behaviors?: CardBehaviorDto[];
  playerResources?: ResourcesDto; // Optional: if provided, enables dynamic affordability highlighting
  resourceStorage?: { [cardId: string]: number }; // Optional: card storage data for validation
  cardId?: string; // Optional: ID of the card these behaviors belong to
  greyOutAll?: boolean; // Optional: if true, grey out all resources (e.g., for played actions)
}

interface ClassifiedBehavior {
  behavior: CardBehaviorDto;
  type:
    | "manual-action"
    | "immediate-production"
    | "immediate-effect"
    | "triggered-effect"
    | "auto-no-background"
    | "discount";
}

const BehaviorSection: React.FC<BehaviorSectionProps> = ({
  behaviors,
  playerResources,
  resourceStorage,
  cardId,
  greyOutAll = false,
}) => {
  if (!behaviors || behaviors.length === 0) {
    return null;
  }

  // Helper function to check if a resource is affordable
  const isResourceAffordable = (
    resource: any,
    isInput: boolean = true,
  ): boolean => {
    if (greyOutAll) return false; // Grey out everything if greyOutAll is true
    if (!playerResources) return true; // If no player resources provided, show normally
    if (!isInput) return true; // Outputs are always "affordable" unless greyOutAll

    const resourceType = resource.resourceType || resource.type;
    const amount = resource.amount || 1;
    const target = resource.target;

    switch (resourceType) {
      case "credits":
        return playerResources.credits >= amount;
      case "steel":
        return playerResources.steel >= amount;
      case "titanium":
        return playerResources.titanium >= amount;
      case "plants":
        return playerResources.plants >= amount;
      case "energy":
        return playerResources.energy >= amount;
      case "heat":
        return playerResources.heat >= amount;
    }

    // Check card storage resources
    const cardStorageTypes = [
      "animals",
      "microbes",
      "floaters",
      "science",
      "asteroid",
    ];
    if (cardStorageTypes.includes(resourceType) && target === "self-card") {
      if (!resourceStorage || !cardId) return true; // Can't validate, show normally
      const currentStorage = resourceStorage[cardId] || 0;
      return currentStorage >= amount;
    }

    return true; // For other resource types, show normally
  };

  const classifyBehaviors = (
    behaviors: CardBehaviorDto[],
  ): ClassifiedBehavior[] => {
    return behaviors.map((behavior) => {
      const hasTrigger = behavior.triggers && behavior.triggers.length > 0;
      const triggerType = hasTrigger ? behavior.triggers?.[0]?.type : null;
      const hasInputs = behavior.inputs && behavior.inputs.length > 0;
      const hasChoices = behavior.choices && behavior.choices.length > 0;
      const hasProduction =
        behavior.outputs &&
        behavior.outputs.some((output: any) =>
          output.type?.includes("production"),
        );

      const hasDiscount =
        behavior.outputs &&
        behavior.outputs.some((output: any) => output.type === "discount");

      if (hasDiscount) {
        return { behavior, type: "discount" };
      }

      if (triggerType === "manual" || hasChoices) {
        return { behavior, type: "manual-action" };
      }

      if (triggerType === "auto" && !hasInputs) {
        return { behavior, type: "auto-no-background" };
      }

      if (hasTrigger && hasInputs) {
        return { behavior, type: "triggered-effect" };
      }

      if (hasProduction && (!hasTrigger || triggerType === "auto")) {
        return { behavior, type: "immediate-production" };
      }

      return { behavior, type: "immediate-effect" };
    });
  };

  const classifiedBehaviors = classifyBehaviors(behaviors);

  const renderIcon = (
    resourceType: string,
    _isProduction: boolean = false,
    isAttack: boolean = false,
    context: "standalone" | "action" | "production" | "default" = "default",
    isAffordable: boolean = true,
  ): React.ReactNode => {
    const cleanType = resourceType?.toLowerCase().replace(/[_\s]/g, "-");
    const icon = getIconPath(resourceType);

    if (!icon) return null;

    let iconClass =
      "w-[26px] h-[26px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] max-md:w-[22px] max-md:h-[22px]";
    const isTag = isTagIcon(cleanType);
    const isPlacement =
      cleanType === "city-placement" ||
      cleanType === "greenery-placement" ||
      cleanType === "ocean-placement";
    const isTR = cleanType === "tr";
    const isCard =
      cleanType === "card-draw" ||
      cleanType === "card-take" ||
      cleanType === "card-peek";

    // Check if this should be a standalone larger icon
    const isStandaloneTile =
      cleanType === "city-tile" ||
      cleanType === "greenery-tile" ||
      cleanType === "ocean-tile";
    const isStandaloneCard = cleanType === "card-draw";
    const shouldUseStandaloneSize =
      context === "standalone" && (isStandaloneTile || isStandaloneCard);

    if (isAttack) {
      iconClass =
        "w-[26px] h-[26px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_1px_rgba(244,67,54,0.9))_drop-shadow(0_0_2px_rgba(244,67,54,0.7))] animate-[attackPulse_2s_ease-in-out_infinite] max-md:w-[22px] max-md:h-[22px]";
    } else if (shouldUseStandaloneSize) {
      iconClass =
        "w-9 h-9 object-contain [filter:drop-shadow(0_1px_3px_rgba(0,0,0,0.7))] max-md:w-8 max-md:h-8";
    } else if (isPlacement) {
      iconClass =
        "w-[30px] h-[30px] object-contain [filter:drop-shadow(0_1px_3px_rgba(0,0,0,0.6))] max-md:w-[26px] max-md:h-[26px]";
    } else if (isTR) {
      iconClass =
        "w-8 h-8 object-contain [filter:drop-shadow(0_1px_3px_rgba(0,0,0,0.6))] max-md:w-7 max-md:h-7";
    } else if (isCard) {
      iconClass =
        "w-[30px] h-[30px] object-contain [filter:drop-shadow(0_1px_3px_rgba(0,0,0,0.6))] max-md:w-[26px] max-md:h-[26px]";
    } else if (isTag) {
      iconClass =
        "w-[26px] h-[26px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] max-md:w-[22px] max-md:h-[22px]";
    }

    const finalIconClass = !isAffordable
      ? `${iconClass} opacity-40 [filter:grayscale(0.7)_drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]`
      : iconClass;

    return <img src={icon} alt={cleanType} className={finalIconClass} />;
  };

  interface LayoutRequirement {
    totalIcons: number;
    separatorCount: number;
    separatorPositions: number[];
    behaviorType: string;
    needsMultipleRows: boolean;
    maxHorizontalIcons: number;
  }

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

  const calculateIconRequirements = (
    behavior: any,
    behaviorType: string,
  ): LayoutRequirement => {
    const MAX_HORIZONTAL = 7;
    let totalIcons = 0;
    let separatorCount = 0;
    let separatorPositions: number[] = [];

    // Handle choice-based behaviors
    if (behavior.choices && behavior.choices.length > 0) {
      // For choices, we need to count the total icons across all choice options
      const choiceIcons = behavior.choices.reduce(
        (sum: number, choice: any) => {
          let choiceSum = 0;

          if (choice.inputs && choice.inputs.length > 0) {
            choiceSum += choice.inputs.reduce(
              (inputSum: number, input: any) => {
                const analysis = analyzeResourceDisplayWithConstraints(
                  input,
                  MAX_HORIZONTAL,
                  false,
                );
                return inputSum + analysis.iconCount;
              },
              0,
            );
          }

          if (choice.outputs && choice.outputs.length > 0) {
            choiceSum += choice.outputs.reduce(
              (outputSum: number, output: any) => {
                const analysis = analyzeResourceDisplayWithConstraints(
                  output,
                  MAX_HORIZONTAL,
                  false,
                );
                return outputSum + analysis.iconCount;
              },
              0,
            );
          }

          // Add separator for choice (if it has both inputs and outputs)
          if (choice.inputs?.length > 0 && choice.outputs?.length > 0) {
            choiceSum += 1;
          }

          return sum + choiceSum;
        },
        0,
      );

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
            MAX_HORIZONTAL,
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
              MAX_HORIZONTAL,
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
      needsMultipleRows: totalIcons > MAX_HORIZONTAL,
      maxHorizontalIcons: MAX_HORIZONTAL,
    };
  };

  const createLayoutPlan = (
    behavior: any,
    behaviorType: string,
  ): LayoutPlan => {
    const requirements = calculateIconRequirements(behavior, behaviorType);
    const MAX_HORIZONTAL = 7;

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
                  MAX_HORIZONTAL,
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
                  MAX_HORIZONTAL,
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
                MAX_HORIZONTAL,
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
                MAX_HORIZONTAL,
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

  // Create balanced multi-row layout
  const createMultiRowLayout = (
    behavior: any,
    behaviorType: string,
    _requirements: LayoutRequirement,
  ): LayoutPlan => {
    const rows: IconDisplayInfo[][] = [];
    const MAX_HORIZONTAL = 7;

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
        Math.floor((MAX_HORIZONTAL - 1) / 2),
      ); // Reserve space for separator
      const outputRows = distributeIconsAcrossRows(
        outputs,
        Math.floor((MAX_HORIZONTAL - 1) / 2),
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
              MAX_HORIZONTAL,
              false,
            ),
          );
        });
      }

      const distributedRows = distributeIconsAcrossRows(
        allItems,
        MAX_HORIZONTAL,
      );
      const separators =
        behaviorType !== "auto-no-background"
          ? [{ position: 0, type: ":" }]
          : [];

      return {
        rows: distributedRows,
        separators,
        totalRows: distributedRows.length,
      };
    }
  };

  // Distribute icons across rows with upper-row priority for odd splits
  const distributeIconsAcrossRows = (
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

  const renderResourceFromDisplayInfo = (
    displayInfo: IconDisplayInfo,
    isInput: boolean = false,
    resource?: any,
    isGroupedWithOtherNegatives: boolean = false,
    context: "standalone" | "action" | "production" | "default" = "default",
    isAffordable: boolean = true,
  ): React.ReactNode => {
    const { resourceType, amount, displayMode } = displayInfo;

    const isCredits =
      resourceType === "credits" || resourceType === "credits-production";
    const isDiscount = resourceType === "discount";
    const isProduction = resourceType?.includes("-production");
    const hasPer = resource?.per;
    const isAttack = resource?.target === "any-player";

    // Handle production with per condition (e.g., 1 plant production per plant tag)
    if (isProduction && hasPer) {
      const baseResourceType = resourceType.replace("-production", "");

      let perIcon = null;
      if (hasPer.tag) {
        perIcon = getIconPath(hasPer.tag);
      } else if (hasPer.type) {
        perIcon = getIconPath(hasPer.type);
      }

      if (perIcon) {
        // Special handling for credits-production - use GameIcon with amount inside
        if (baseResourceType === "credits") {
          const itemClasses = !isAffordable
            ? "flex items-center gap-px relative opacity-40 [filter:grayscale(0.7)_drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
            : "flex items-center gap-px relative";

          return (
            <div className="flex flex-wrap gap-[3px] items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-1.5 py-[3px] shadow-[0_1px_3px_rgba(0,0,0,0.2)]">
              <div className={itemClasses}>
                <GameIcon
                  iconType="credits"
                  amount={Math.abs(amount)}
                  size="small"
                />
              </div>
              <span className="text-base font-bold text-white mx-[3px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
                /
              </span>
              <img
                src={perIcon}
                alt={hasPer.tag || hasPer.type}
                className="w-[26px] h-[26px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] max-md:w-[22px] max-md:h-[22px]"
              />
            </div>
          );
        } else {
          // For other resources, use regular icon with amount overlay
          const productionIcon = renderIcon(
            baseResourceType,
            false,
            isAttack,
            "production",
            isAffordable,
          );
          if (productionIcon) {
            return (
              <div className="flex flex-wrap gap-[3px] items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-1.5 py-[3px] shadow-[0_1px_3px_rgba(0,0,0,0.2)]">
                <div className="flex items-center gap-px relative">
                  {amount > 1 && (
                    <span className="text-lg font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)] leading-none flex items-center ml-0.5 max-md:text-xs">
                      {amount}
                    </span>
                  )}
                  {productionIcon}
                </div>
                <span className="text-base font-bold text-white mx-[3px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
                  /
                </span>
                <img
                  src={perIcon}
                  alt={hasPer.tag || hasPer.type}
                  className="w-[26px] h-[26px] object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] max-md:w-[22px] max-md:h-[22px]"
                />
              </div>
            );
          }
        }
      }
    }

    if (isCredits) {
      const creditsClasses = `flex items-center gap-0.5 relative ${isAttack ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_1px_rgba(244,67,54,0.9))_drop-shadow(0_0_2px_rgba(244,67,54,0.7))] animate-[attackPulseCredits_2s_ease-in-out_infinite]" : ""}`;

      // Show minus inside icon if not grouped with other negative resources
      const showMinusInside = amount < 0 && !isGroupedWithOtherNegatives;
      // Show minus outside icon if grouped with other negative resources
      const showMinusOutside =
        (isInput || amount < 0) && isGroupedWithOtherNegatives;

      const finalCreditsClasses = !isAffordable
        ? `${creditsClasses} opacity-40 [filter:grayscale(0.7)_drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]`
        : creditsClasses;

      return (
        <div className={finalCreditsClasses}>
          {showMinusOutside && (
            <span className="text-lg font-bold text-[#ffcdd2] mr-0.5 [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)] flex items-center">
              -
            </span>
          )}
          <GameIcon
            iconType="credits"
            amount={showMinusInside ? amount : Math.abs(amount)}
            size="small"
          />
        </div>
      );
    }

    if (isDiscount) {
      const discountClasses = !isAffordable
        ? "flex items-center gap-0.5 relative opacity-40 [filter:grayscale(0.7)_drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
        : "flex items-center gap-0.5 relative";

      return (
        <div className={discountClasses}>
          <GameIcon iconType="credits" amount={-amount} size="small" />
        </div>
      );
    }

    // Use the passed context or determine based on production status
    let iconContext = context;
    if (iconContext === "default" && isProduction) {
      iconContext = "production";
    }

    const iconElement = renderIcon(
      resourceType,
      false,
      isAttack,
      iconContext,
      isAffordable,
    );
    if (!iconElement) {
      return (
        <span className="text-xs font-semibold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)] max-md:text-[11px]">
          {isInput ? "-" : "+"}
          {amount} {resourceType}
        </span>
      );
    }

    if (displayMode === "individual") {
      const absoluteAmount = Math.abs(amount);
      return (
        <div className="flex items-center gap-px relative">
          {(isInput || amount < 0) && (
            <span className="text-lg font-bold text-[#ffcdd2] mr-0.5 [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)] flex items-center">
              -
            </span>
          )}
          {Array.from({ length: absoluteAmount }, (_, i) => (
            <React.Fragment key={i}>{iconElement}</React.Fragment>
          ))}
        </div>
      );
    } else {
      return (
        <div className="flex items-center gap-0.5 relative">
          {isInput && (
            <span className="text-lg font-bold text-[#ffcdd2] mr-0.5 [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)] flex items-center">
              -
            </span>
          )}
          <span className="text-[11px] font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] mr-px">
            {amount}
          </span>
          {iconElement}
        </div>
      );
    }
  };

  const renderManualActionLayout = (
    behavior: any,
    _layoutPlan: LayoutPlan,
  ): React.ReactNode => {
    // Handle choice-based behaviors
    if (behavior.choices && behavior.choices.length > 0) {
      return (
        <div className="flex flex-col gap-1.5 items-center w-full">
          {behavior.choices.map((choice: any, choiceIndex: number) => (
            <div
              key={`choice-${choiceIndex}`}
              className="flex items-center gap-1 w-full justify-center"
            >
              {/* Input side for this choice */}
              <div className="flex flex-col gap-0.5 items-center min-w-0">
                {choice.inputs &&
                  choice.inputs.map((input: any, inputIndex: number) => {
                    const displayInfo = analyzeResourceDisplayWithConstraints(
                      input,
                      3,
                      false,
                    );
                    return (
                      <React.Fragment
                        key={`choice-${choiceIndex}-input-${inputIndex}`}
                      >
                        {renderResourceFromDisplayInfo(
                          displayInfo,
                          true,
                          input,
                          false,
                          "action",
                          isResourceAffordable(input, true),
                        )}
                      </React.Fragment>
                    );
                  })}
              </div>

              {/* Arrow separator for this choice */}
              {choice.inputs?.length > 0 && choice.outputs?.length > 0 && (
                <div className="flex items-center justify-center text-white text-base font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] min-w-[20px] z-[1]">
                  →
                </div>
              )}

              {/* Output side for this choice */}
              <div className="flex flex-col gap-0.5 items-center min-w-0">
                {choice.outputs &&
                  choice.outputs.map((output: any, outputIndex: number) => {
                    const displayInfo = analyzeResourceDisplayWithConstraints(
                      output,
                      3,
                      false,
                    );
                    return (
                      <React.Fragment
                        key={`choice-${choiceIndex}-output-${outputIndex}`}
                      >
                        {renderResourceFromDisplayInfo(
                          displayInfo,
                          false,
                          output,
                          false,
                          "action",
                          isResourceAffordable(output, false),
                        )}
                      </React.Fragment>
                    );
                  })}
              </div>

              {/* Add "OR" separator between choices (except for the last one) */}
              {choiceIndex < behavior.choices.length - 1 && (
                <div className="text-xs font-semibold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] my-0.5 mx-2 bg-white/10 py-0.5 px-1.5 rounded-[2px] backdrop-blur-[2px]">
                  OR
                </div>
              )}
            </div>
          ))}
        </div>
      );
    }

    // Regular behavior handling
    return (
      <div className="flex items-center justify-center gap-2 w-full">
        {/* Input side */}
        <div className="flex flex-col gap-0.5 items-center min-w-0">
          {behavior.inputs &&
            behavior.inputs.map((input: any, inputIndex: number) => {
              const displayInfo = analyzeResourceDisplayWithConstraints(
                input,
                3,
                false,
              );
              return (
                <React.Fragment key={`input-${inputIndex}`}>
                  {renderResourceFromDisplayInfo(
                    displayInfo,
                    true,
                    input,
                    false,
                    "action",
                    isResourceAffordable(input, true),
                  )}
                </React.Fragment>
              );
            })}
        </div>

        {/* Arrow separator */}
        <div className="flex items-center justify-center text-white text-base font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] min-w-[20px] z-[1]">
          →
        </div>

        {/* Output side */}
        <div className="flex flex-col gap-0.5 items-center min-w-0">
          {behavior.outputs &&
            behavior.outputs.map((output: any, outputIndex: number) => {
              const displayInfo = analyzeResourceDisplayWithConstraints(
                output,
                3,
                false,
              );
              return (
                <React.Fragment key={`output-${outputIndex}`}>
                  {renderResourceFromDisplayInfo(
                    displayInfo,
                    false,
                    output,
                    false,
                    "action",
                    isResourceAffordable(output, false),
                  )}
                </React.Fragment>
              );
            })}
        </div>
      </div>
    );
  };

  // Render triggered effect layout (condition : outputs)
  const renderTriggeredEffectLayout = (
    behavior: any,
    _layoutPlan: LayoutPlan,
  ): React.ReactNode => {
    return (
      <div className="flex flex-col gap-[3px] items-center justify-center">
        <div className="flex gap-[3px] items-center justify-center">
          {/* Trigger conditions */}
          {behavior.triggers && behavior.triggers.length > 0 && (
            <>
              <div className="flex gap-[3px] items-center">
                {behavior.triggers.map((trigger: any, triggerIndex: number) => (
                  <span
                    key={triggerIndex}
                    className="text-xs font-semibold text-[#e0e0e0] capitalize [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)] max-md:text-[11px]"
                  >
                    {trigger.description || trigger.type || "trigger"}
                  </span>
                ))}
              </div>
              <span className="flex items-center justify-center text-white text-base font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] min-w-[20px] z-[1]">
                :
              </span>
            </>
          )}

          {/* Outputs in same row if they fit */}
          {behavior.outputs &&
            behavior.outputs.map((output: any, index: number) => {
              const displayInfo = analyzeResourceDisplayWithConstraints(
                output,
                6,
                false,
              );
              return (
                <React.Fragment key={`triggered-output-${index}`}>
                  {renderResourceFromDisplayInfo(
                    displayInfo,
                    false,
                    output,
                    false,
                    "default",
                    isResourceAffordable(output, false),
                  )}
                </React.Fragment>
              );
            })}
        </div>
      </div>
    );
  };

  // Render immediate resource layout (flexible grid for outputs only)
  const renderImmediateResourceLayout = (
    behavior: any,
    _layoutPlan: LayoutPlan,
  ): React.ReactNode => {
    if (!behavior.outputs || behavior.outputs.length === 0) return null;

    // Separate production and non-production outputs
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

    // Separate per-condition production (which already has its own wrapper) from regular production
    const perConditionProduction = productionOutputs.filter(
      (output: any) => output.per,
    );
    const regularProduction = productionOutputs.filter(
      (output: any) => !output.per,
    );

    // Separate negative and positive production outputs (only for regular production, not per-condition)
    const negativeProduction = regularProduction.filter(
      (output: any) => (output.amount ?? 1) < 0,
    );
    const positiveProduction = regularProduction.filter(
      (output: any) => (output.amount ?? 1) >= 0,
    );

    // Separate negative and positive non-production outputs
    const negativeOutputs = nonProductionOutputs.filter(
      (output: any) => (output.amount ?? 1) < 0,
    );
    const positiveOutputs = nonProductionOutputs.filter(
      (output: any) => (output.amount ?? 1) >= 0,
    );

    return (
      <div className="flex flex-wrap gap-[3px] items-center justify-center max-w-full">
        {/* Regular production outputs in wrapper with row-based grouping */}
        {regularProduction.length > 0 && (
          <div className="flex flex-wrap gap-[3px] items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-1.5 py-[3px] shadow-[0_1px_3px_rgba(0,0,0,0.2)]">
            <div className="flex flex-col gap-[3px] items-center justify-center">
              {/* Negative production on first row */}
              {negativeProduction.length > 0 && (
                <div className="flex gap-[3px] items-center justify-center">
                  {negativeProduction.map((output: any, index: number) => {
                    const displayInfo = analyzeResourceDisplayWithConstraints(
                      output,
                      7,
                      false,
                    );
                    const isGrouped = negativeProduction.length > 1;
                    return (
                      <React.Fragment key={`neg-prod-${index}`}>
                        {renderResourceFromDisplayInfo(
                          displayInfo,
                          false,
                          output,
                          isGrouped,
                          "standalone",
                        )}
                      </React.Fragment>
                    );
                  })}
                </div>
              )}

              {/* Positive production on second row */}
              {positiveProduction.length > 0 && (
                <div className="flex gap-[3px] items-center justify-center">
                  {positiveProduction.map((output: any, index: number) => {
                    const displayInfo = analyzeResourceDisplayWithConstraints(
                      output,
                      7,
                      false,
                    );
                    return (
                      <React.Fragment key={`pos-prod-${index}`}>
                        {renderResourceFromDisplayInfo(
                          displayInfo,
                          false,
                          output,
                          false,
                          "standalone",
                        )}
                      </React.Fragment>
                    );
                  })}
                </div>
              )}
            </div>
          </div>
        )}

        {/* Per-condition production outputs (already have their own wrapper) */}
        {perConditionProduction.length > 0 && (
          <div className="flex flex-col gap-[3px] items-center justify-center">
            {perConditionProduction.map((output: any, index: number) => {
              const displayInfo = analyzeResourceDisplayWithConstraints(
                output,
                7,
                false,
              );
              return (
                <React.Fragment key={`per-prod-${index}`}>
                  {renderResourceFromDisplayInfo(
                    displayInfo,
                    false,
                    output,
                    false,
                    "standalone",
                  )}
                </React.Fragment>
              );
            })}
          </div>
        )}

        {/* Non-production outputs */}
        {nonProductionOutputs.length > 0 && (
          <div className="flex flex-col gap-[3px] items-center justify-center">
            {/* Negative resources on first row */}
            {negativeOutputs.length > 0 && (
              <div className="flex gap-[3px] items-center justify-center">
                {negativeOutputs.map((output: any, index: number) => {
                  const displayInfo = analyzeResourceDisplayWithConstraints(
                    output,
                    7,
                    false,
                  );
                  const isGrouped = negativeOutputs.length > 1;
                  return (
                    <React.Fragment key={`neg-${index}`}>
                      {renderResourceFromDisplayInfo(
                        displayInfo,
                        false,
                        output,
                        isGrouped,
                        "standalone",
                      )}
                    </React.Fragment>
                  );
                })}
              </div>
            )}

            {/* Positive resources on second row */}
            {positiveOutputs.length > 0 && (
              <div className="flex gap-[3px] items-center justify-center">
                {positiveOutputs.map((output: any, index: number) => {
                  const displayInfo = analyzeResourceDisplayWithConstraints(
                    output,
                    7,
                    false,
                  );
                  return (
                    <React.Fragment key={`pos-${index}`}>
                      {renderResourceFromDisplayInfo(
                        displayInfo,
                        false,
                        output,
                        false,
                        "standalone",
                      )}
                    </React.Fragment>
                  );
                })}
              </div>
            )}
          </div>
        )}
      </div>
    );
  };

  // Render discount layout (: discount amount)
  const renderDiscountLayout = (
    behavior: any,
    _layoutPlan: LayoutPlan,
  ): React.ReactNode => {
    if (!behavior.outputs || behavior.outputs.length === 0) return null;

    return (
      <div className="flex gap-[3px] items-center justify-center">
        <span className="flex items-center justify-center text-white text-base font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] min-w-[20px] z-[1]">
          :
        </span>
        {behavior.outputs.map((output: any, index: number) => {
          const displayInfo = analyzeResourceDisplayWithConstraints(
            output,
            6,
            false,
          );
          return (
            <React.Fragment key={`discount-${index}`}>
              {renderResourceFromDisplayInfo(
                displayInfo,
                false,
                output,
                false,
                "standalone",
                isResourceAffordable(output, false),
              )}
            </React.Fragment>
          );
        })}
      </div>
    );
  };

  // Main behavior renderer using new layout system
  const renderBehaviorWithNewLayout = (
    classifiedBehavior: ClassifiedBehavior,
    index: number,
  ): React.ReactNode => {
    const { behavior, type } = classifiedBehavior;
    const layoutPlan = createLayoutPlan(behavior, type);

    let behaviorContent: React.ReactNode;

    switch (type) {
      case "manual-action":
        behaviorContent = renderManualActionLayout(behavior, layoutPlan);
        break;
      case "triggered-effect":
        behaviorContent = renderTriggeredEffectLayout(behavior, layoutPlan);
        break;
      case "auto-no-background":
        behaviorContent = renderImmediateResourceLayout(behavior, layoutPlan);
        break;
      case "discount":
        behaviorContent = renderDiscountLayout(behavior, layoutPlan);
        break;
      default:
        behaviorContent = renderImmediateResourceLayout(behavior, layoutPlan);
    }

    // Wrap in appropriate container based on type
    if (type === "auto-no-background" || type === "discount") {
      return (
        <div
          key={index}
          className="flex items-center justify-center my-px p-[3px] min-h-8 max-md:p-px max-md:my-px"
        >
          {behaviorContent}
        </div>
      );
    } else {
      const typeStyles = {
        "manual-action":
          "bg-[linear-gradient(135deg,rgba(33,150,243,0.35)_0%,rgba(25,118,210,0.3)_100%)] border-[rgba(33,150,243,0.5)] shadow-[0_2px_4px_rgba(33,150,243,0.3)] w-auto min-w-[100px] max-w-[calc(100%-20px)]",
        "triggered-effect":
          "bg-white/[0.08] border-white/20 shadow-[0_1px_3px_rgba(0,0,0,0.15)]",
        "immediate-production":
          "bg-[linear-gradient(135deg,rgba(139,89,42,0.35)_0%,rgba(101,67,33,0.3)_100%)] border-[rgba(139,89,42,0.5)] shadow-[0_2px_4px_rgba(139,89,42,0.25)]",
        "immediate-effect":
          "bg-white/[0.08] border-white/20 shadow-[0_1px_3px_rgba(0,0,0,0.15)]",
      };

      return (
        <div
          key={index}
          className={`rounded-[3px] px-2 py-1 min-h-8 my-px border border-white/10 backdrop-blur-[2px] flex items-center w-[calc(100%-20px)] ${typeStyles[type] || ""} max-md:px-1.5 max-md:py-[3px] max-md:min-h-7 max-md:my-px`}
        >
          <div className="flex items-center gap-1.5 flex-nowrap w-full justify-center max-md:gap-1">
            {behaviorContent}
          </div>
        </div>
      );
    }
  };

  interface CardLayoutPlan {
    behaviors: Array<{
      behaviorIndex: number;
      layoutPlan: LayoutPlan;
      estimatedRows: number;
    }>;
    totalEstimatedRows: number;
    needsOverflowHandling: boolean;
    maxRows: number;
  }

  // Analyze all behaviors on a card and create coordinated layout plan
  const analyzeCardLayout = (
    classifiedBehaviors: ClassifiedBehavior[],
  ): CardLayoutPlan => {
    const MAX_CARD_ROWS = 4;
    const behaviorPlans = classifiedBehaviors.map(
      (classifiedBehavior, index) => {
        const { behavior, type } = classifiedBehavior;
        const layoutPlan = createLayoutPlan(behavior, type);

        // Estimate rows needed for this behavior
        let estimatedRows = layoutPlan.totalRows;

        // Add extra row for behavior container if it's not auto-no-background
        if (type !== "auto-no-background") {
          estimatedRows = Math.max(1, estimatedRows);
        }

        return {
          behaviorIndex: index,
          layoutPlan,
          estimatedRows,
        };
      },
    );

    const totalEstimatedRows = behaviorPlans.reduce(
      (sum, plan) => sum + plan.estimatedRows,
      0,
    );
    const needsOverflowHandling = totalEstimatedRows > MAX_CARD_ROWS;

    return {
      behaviors: behaviorPlans,
      totalEstimatedRows,
      needsOverflowHandling,
      maxRows: MAX_CARD_ROWS,
    };
  };

  // Optimize behaviors for space when multiple behaviors are present
  const optimizeBehaviorsForSpace = (
    classifiedBehaviors: ClassifiedBehavior[],
    cardLayoutPlan: CardLayoutPlan,
  ): ClassifiedBehavior[] => {
    if (!cardLayoutPlan.needsOverflowHandling) {
      return classifiedBehaviors; // No optimization needed
    }

    // Strategy: Convert more resources to NxIcon format to save space
    const optimizedBehaviors = classifiedBehaviors.map((classifiedBehavior) => {
      const { behavior, type } = classifiedBehavior;

      // Create optimized behavior with more aggressive number formatting
      const optimizedBehavior = {
        ...behavior,
        inputs: behavior.inputs?.map((input: any) => ({
          ...input,
          // Force number format for amounts > 2 instead of > 4
          forceNumberFormat: input.amount > 2,
        })),
        outputs: behavior.outputs?.map((output: any) => ({
          ...output,
          // Force number format for amounts > 2 instead of > 4
          forceNumberFormat: output.amount > 2,
        })),
      };

      return { behavior: optimizedBehavior, type };
    });

    return optimizedBehaviors;
  };

  // Enhanced resource display analysis that considers space constraints
  const analyzeResourceDisplayWithConstraints = (
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
    const maxIndividualIcons =
      forceCompact || resource.forceNumberFormat ? 2 : 3;
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

  // Merge multiple auto production behaviors into a single behavior
  const mergeAutoProductionBehaviors = (
    classifiedBehaviors: ClassifiedBehavior[],
  ): ClassifiedBehavior[] => {
    const autoProductionBehaviors: ClassifiedBehavior[] = [];
    const otherBehaviors: ClassifiedBehavior[] = [];

    // Separate auto production behaviors from others
    classifiedBehaviors.forEach((classifiedBehavior) => {
      const { behavior, type } = classifiedBehavior;
      const isAutoProduction =
        type === "auto-no-background" &&
        behavior.outputs &&
        behavior.outputs.every(
          (output: any) =>
            output.type?.includes("production") ||
            output.resourceType?.includes("production") ||
            output.isProduction,
        );

      if (isAutoProduction) {
        autoProductionBehaviors.push(classifiedBehavior);
      } else {
        otherBehaviors.push(classifiedBehavior);
      }
    });

    // If we have multiple auto production behaviors, merge them
    if (autoProductionBehaviors.length > 1) {
      const mergedOutputs: any[] = [];

      autoProductionBehaviors.forEach((classifiedBehavior) => {
        if (classifiedBehavior.behavior.outputs) {
          mergedOutputs.push(...classifiedBehavior.behavior.outputs);
        }
      });

      const mergedBehavior: ClassifiedBehavior = {
        behavior: {
          triggers: [{ type: "auto" }],
          outputs: mergedOutputs,
        },
        type: "auto-no-background",
      };

      return [mergedBehavior, ...otherBehaviors];
    }

    return classifiedBehaviors;
  };

  // Render behaviors with card-level coordination
  const renderBehaviorsWithCoordination = (
    classifiedBehaviors: ClassifiedBehavior[],
  ): React.ReactNode => {
    // First, merge auto production behaviors if needed
    const mergedBehaviors = mergeAutoProductionBehaviors(classifiedBehaviors);

    const cardLayoutPlan = analyzeCardLayout(mergedBehaviors);
    const optimizedBehaviors = optimizeBehaviorsForSpace(
      mergedBehaviors,
      cardLayoutPlan,
    );

    // If overflow is needed, prepare for future rolling effect
    const containerClass = cardLayoutPlan.needsOverflowHandling
      ? "absolute bottom-[45px] left-2 right-2 flex flex-col gap-[3px] items-center z-0 max-h-[120px] overflow-y-auto scroll-smooth [scrollbar-width:thin] [&::-webkit-scrollbar]:w-0.5 [&::-webkit-scrollbar-track]:bg-white/10 [&::-webkit-scrollbar-track]:rounded-px [&::-webkit-scrollbar-thumb]:bg-white/30 [&::-webkit-scrollbar-thumb]:rounded-px max-md:bottom-[35px] max-md:left-1.5 max-md:right-1.5 max-md:gap-px"
      : "absolute bottom-[45px] left-2 right-2 flex flex-col gap-[3px] items-center z-0 max-md:bottom-[35px] max-md:left-1.5 max-md:right-1.5 max-md:gap-px";

    return (
      <div className={containerClass}>
        {optimizedBehaviors.map((classifiedBehavior, index) =>
          renderBehaviorWithNewLayout(classifiedBehavior, index),
        )}

        {/* Future: Add rolling effect indicators here when needed */}
        {cardLayoutPlan.needsOverflowHandling && (
          <div className="flex items-center justify-center h-4 text-[10px] text-white/60 italic">
            {/* This could be a visual indicator that there are more behaviors */}
          </div>
        )}
      </div>
    );
  };

  // Use the new coordinated layout system
  return renderBehaviorsWithCoordination(classifiedBehaviors);
};

export default BehaviorSection;
