import React from "react";
import styles from "./BehaviorSection.module.css";
import { CardBehaviorDto } from "@/types/generated/api-types.ts";
import MegaCreditIcon from "../display/MegaCreditIcon.tsx";

interface BehaviorSectionProps {
  behaviors?: CardBehaviorDto[];
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

const BehaviorSection: React.FC<BehaviorSectionProps> = ({ behaviors }) => {
  if (!behaviors || behaviors.length === 0) {
    return null;
  }

  const getResourceIcon = (resourceType: string): string | null => {
    const iconMap: { [key: string]: string } = {
      credits: "/assets/resources/megacredit.png",
      steel: "/assets/resources/steel.png",
      titanium: "/assets/resources/titanium.png",
      plants: "/assets/resources/plant.png",
      energy: "/assets/resources/power.png",
      heat: "/assets/resources/heat.png",
      microbes: "/assets/resources/microbe.png",
      animals: "/assets/resources/animal.png",
      floater: "/assets/resources/floater.png",
      floaters: "/assets/resources/floater.png",
      science: "/assets/resources/science.png",
      asteroid: "/assets/resources/asteroid.png",
      "card-draw": "/assets/resources/card.png",
      "card-take": "/assets/resources/card.png",
      "card-peek": "/assets/resources/card.png",
      "city-placement": "/assets/tiles/city.png",
      "ocean-placement": "/assets/tiles/ocean.png",
      "greenery-placement": "/assets/tiles/greenery.png",
      "city-tile": "/assets/tiles/city.png",
      "ocean-tile": "/assets/tiles/ocean.png",
      "greenery-tile": "/assets/tiles/greenery.png",
      temperature: "/assets/global-parameters/temperature.png",
      oxygen: "/assets/global-parameters/oxygen.png",
      tr: "/assets/resources/tr.png",
      "credits-production": "/assets/resources/megacredit.png",
      "steel-production": "/assets/resources/steel.png",
      "titanium-production": "/assets/resources/titanium.png",
      "plants-production": "/assets/resources/plant.png",
      "energy-production": "/assets/resources/power.png",
      "heat-production": "/assets/resources/heat.png",
      tag: "/assets/tags/wild.png",
      discount: "/assets/resources/megacredit.png",
    };

    const cleanType = resourceType?.toLowerCase().replace(/[_\s]/g, "-");
    return iconMap[cleanType] || iconMap[resourceType?.toLowerCase()] || null;
  };

  const getTagIcon = (tagName: string): string | null => {
    const tagMap: { [key: string]: string } = {
      earth: "/assets/tags/earth.png",
      science: "/assets/tags/science.png",
      plant: "/assets/tags/plant.png",
      microbe: "/assets/tags/microbe.png",
      animal: "/assets/tags/animal.png",
      power: "/assets/tags/power.png",
      space: "/assets/tags/space.png",
      building: "/assets/tags/building.png",
      city: "/assets/tags/city.png",
      jovian: "/assets/tags/jovian.png",
      venus: "/assets/tags/venus.png",
      event: "/assets/tags/event.png",
      mars: "/assets/tags/mars.png",
      moon: "/assets/tags/moon.png",
      wild: "/assets/tags/wild.png",
      clone: "/assets/tags/clone.png",
      crime: "/assets/tags/crime.png",
    };

    const cleanTag = tagName?.toLowerCase().replace(/[_\s]/g, "-");
    return tagMap[cleanTag] || tagMap[tagName?.toLowerCase()] || null;
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

  // Helper function to get just the icon URL (string) without React element
  const getIconUrl = (resourceType: string): string | null => {
    const cleanType = resourceType?.toLowerCase().replace(/[_\s]/g, "-");
    const tagIcon = getTagIcon(cleanType);

    if (tagIcon) {
      return tagIcon;
    } else {
      let icon = getResourceIcon(cleanType);

      if (!icon && cleanType.includes("production")) {
        const baseResourceType = cleanType.replace("-production", "");
        icon = getResourceIcon(baseResourceType);
      }

      return icon;
    }
  };

  const renderIcon = (
    resourceType: string,
    _isProduction: boolean = false,
    isAttack: boolean = false,
    context: "standalone" | "action" | "production" | "default" = "default",
  ): React.ReactNode => {
    const cleanType = resourceType?.toLowerCase().replace(/[_\s]/g, "-");
    const icon = getIconUrl(resourceType);

    if (!icon) return null;

    let iconClass = styles.resourceIcon;
    const isTag = getTagIcon(cleanType);
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
      iconClass = styles.attackResourceIcon;
    } else if (shouldUseStandaloneSize) {
      iconClass = styles.standaloneIcon;
    } else if (isPlacement) {
      iconClass = styles.placementIcon;
    } else if (isTR) {
      iconClass = styles.trIcon;
    } else if (isCard) {
      iconClass = styles.cardIcon;
    } else if (isTag) {
      iconClass = styles.tagIcon;
    }

    return <img src={icon} alt={cleanType} className={iconClass} />;
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
        perIcon = getTagIcon(hasPer.tag);
      } else if (hasPer.type) {
        perIcon = getIconUrl(hasPer.type);
      }

      if (perIcon) {
        // Special handling for credits-production - use MegaCreditIcon with value inside
        if (baseResourceType === "credits") {
          return (
            <div className={styles.gridProductionWrapper}>
              <div className={styles.flexibleResourceItem}>
                <MegaCreditIcon value={Math.abs(amount)} size="small" />
              </div>
              <span className={styles.perSeparator}>/</span>
              <img
                src={perIcon}
                alt={hasPer.tag || hasPer.type}
                className={styles.tagIcon}
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
          );
          if (productionIcon) {
            return (
              <div className={styles.gridProductionWrapper}>
                <div className={styles.flexibleResourceItem}>
                  {amount > 1 && (
                    <span className={styles.resourceAmount}>{amount}</span>
                  )}
                  {productionIcon}
                </div>
                <span className={styles.perSeparator}>/</span>
                <img
                  src={perIcon}
                  alt={hasPer.tag || hasPer.type}
                  className={styles.tagIcon}
                />
              </div>
            );
          }
        }
      }
    }

    if (isCredits) {
      const creditsClasses = `${styles.flexibleResourceItem} ${styles.numbered} ${isAttack ? styles.attackCreditsWrapper : ""}`;

      // Show minus inside icon if not grouped with other negative resources
      const showMinusInside = amount < 0 && !isGroupedWithOtherNegatives;
      // Show minus outside icon if grouped with other negative resources
      const showMinusOutside =
        (isInput || amount < 0) && isGroupedWithOtherNegatives;

      return (
        <div className={creditsClasses}>
          {showMinusOutside && <span className={styles.minusSign}>-</span>}
          <MegaCreditIcon
            value={showMinusInside ? amount : Math.abs(amount)}
            size="small"
          />
        </div>
      );
    }

    if (isDiscount) {
      return (
        <div className={`${styles.flexibleResourceItem} ${styles.numbered}`}>
          <MegaCreditIcon value={-amount} size="small" />
        </div>
      );
    }

    // Use the passed context or determine based on production status
    let iconContext = context;
    if (iconContext === "default" && isProduction) {
      iconContext = "production";
    }

    const iconElement = renderIcon(resourceType, false, isAttack, iconContext);
    if (!iconElement) {
      return (
        <span className={styles.resourceText}>
          {isInput ? "-" : "+"}
          {amount} {resourceType}
        </span>
      );
    }

    if (displayMode === "individual") {
      const absoluteAmount = Math.abs(amount);
      return (
        <div className={`${styles.flexibleResourceItem} ${styles.individual}`}>
          {(isInput || amount < 0) && (
            <span className={styles.minusSign}>-</span>
          )}
          {Array.from({ length: absoluteAmount }, (_, i) => (
            <React.Fragment key={i}>{iconElement}</React.Fragment>
          ))}
        </div>
      );
    } else {
      return (
        <div
          className={`${styles.flexibleResourceItem} ${styles.numbered} ${styles.numberIconDisplay}`}
        >
          {isInput && <span className={styles.minusSign}>-</span>}
          <span className={styles.resourceAmount}>{amount}</span>
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
        <div className={styles.choicesContainer}>
          {behavior.choices.map((choice: any, choiceIndex: number) => (
            <div key={`choice-${choiceIndex}`} className={styles.choiceOption}>
              {/* Input side for this choice */}
              <div className={styles.actionInputs}>
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
                        )}
                      </React.Fragment>
                    );
                  })}
              </div>

              {/* Arrow separator for this choice */}
              {choice.inputs?.length > 0 && choice.outputs?.length > 0 && (
                <div className={styles.gridSeparator}>→</div>
              )}

              {/* Output side for this choice */}
              <div className={styles.actionOutputs}>
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
                        )}
                      </React.Fragment>
                    );
                  })}
              </div>

              {/* Add "OR" separator between choices (except for the last one) */}
              {choiceIndex < behavior.choices.length - 1 && (
                <div className={styles.choiceSeparator}>OR</div>
              )}
            </div>
          ))}
        </div>
      );
    }

    // Regular behavior handling
    return (
      <div className={styles.balancedActionLayout}>
        {/* Input side */}
        <div className={styles.actionInputs}>
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
                  )}
                </React.Fragment>
              );
            })}
        </div>

        {/* Arrow separator */}
        <div className={styles.gridSeparator}>→</div>

        {/* Output side */}
        <div className={styles.actionOutputs}>
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
      <div className={styles.iconGrid}>
        <div className={styles.iconGridRow}>
          {/* Trigger conditions */}
          {behavior.triggers && behavior.triggers.length > 0 && (
            <>
              <div className={styles.triggerConditions}>
                {behavior.triggers.map((trigger: any, triggerIndex: number) => (
                  <span key={triggerIndex} className={styles.triggerText}>
                    {trigger.description || trigger.type || "trigger"}
                  </span>
                ))}
              </div>
              <span className={styles.gridSeparator}>:</span>
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
      <div className={styles.immediateResourceGrid}>
        {/* Regular production outputs in wrapper with row-based grouping */}
        {regularProduction.length > 0 && (
          <div className={styles.gridProductionWrapper}>
            <div className={styles.iconGrid}>
              {/* Negative production on first row */}
              {negativeProduction.length > 0 && (
                <div className={styles.iconGridRow}>
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
                <div className={styles.iconGridRow}>
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
          <div className={styles.iconGrid}>
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
          <div className={styles.iconGrid}>
            {/* Negative resources on first row */}
            {negativeOutputs.length > 0 && (
              <div className={styles.iconGridRow}>
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
              <div className={styles.iconGridRow}>
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
      <div className={styles.iconGridRow}>
        <span className={styles.gridSeparator}>:</span>
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
        <div key={index} className={styles.autoEffectContainer}>
          {behaviorContent}
        </div>
      );
    } else {
      return (
        <div
          key={index}
          className={`${styles.behaviorItem} ${styles[type]} ${type === "manual-action" ? styles.fitContent : ""}`}
        >
          <div className={styles.behaviorContent}>{behaviorContent}</div>
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
      ? `${styles.behaviorSection} ${styles.behaviorOverflow}`
      : styles.behaviorSection;

    return (
      <div className={containerClass}>
        {optimizedBehaviors.map((classifiedBehavior, index) =>
          renderBehaviorWithNewLayout(classifiedBehavior, index),
        )}

        {/* Future: Add rolling effect indicators here when needed */}
        {cardLayoutPlan.needsOverflowHandling && (
          <div className={styles.overflowIndicator}>
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
