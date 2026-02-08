import React from "react";
import ResourceDisplay from "./ResourceDisplay.tsx";
import GameIcon from "../../../display/GameIcon.tsx";
import { CardBehaviorDto, SelectorDto, MinMaxValueDto } from "@/types/generated/api-types.ts";

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

interface TriggeredEffectLayoutProps {
  behavior: any;
  mergedBehaviors?: CardBehaviorDto[];
  layoutPlan: LayoutPlan;
  isResourceAffordable: (resource: any, isInput: boolean) => boolean;
  analyzeResourceDisplayWithConstraints: (
    resource: any,
    availableSpace: number,
    forceCompact: boolean,
  ) => IconDisplayInfo;
  tileScaleInfo: TileScaleInfo;
}

// Extract requiredOriginalCost from selectors (new location) or condition level (legacy)
const getRequiredOriginalCost = (
  selectors: SelectorDto[] | undefined,
  conditionLevelCost: MinMaxValueDto | undefined,
): MinMaxValueDto | undefined => {
  if (selectors) {
    for (const sel of selectors) {
      if (sel.requiredOriginalCost) {
        return sel.requiredOriginalCost;
      }
    }
  }
  return conditionLevelCost;
};

// Render a single selector (AND logic: tags together, then card type)
const renderSelector = (
  selector: any,
  selectorIndex: number,
  triggerIndex: number,
  redGlowClass: string,
): React.ReactNode => {
  const elements: React.ReactNode[] = [];

  if (selector.tags && selector.tags.length > 0) {
    selector.tags.forEach((tag: string, tagIndex: number) => {
      elements.push(
        <GameIcon
          key={`tag-${triggerIndex}-${selectorIndex}-${tagIndex}`}
          iconType={`${tag}-tag`}
          size="small"
        />,
      );
    });
  }

  if (selector.cardTypes && selector.cardTypes.length > 0) {
    selector.cardTypes.forEach((cardType: string, typeIndex: number) => {
      if (cardType === "event") {
        elements.push(
          <GameIcon
            key={`type-${triggerIndex}-${selectorIndex}-${typeIndex}`}
            iconType="event"
            size="small"
          />,
        );
      } else {
        elements.push(
          <span
            key={`type-${triggerIndex}-${selectorIndex}-${typeIndex}`}
            className="text-xs font-semibold text-[#e0e0e0] capitalize [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)] max-md:text-[11px]"
          >
            {cardType}
          </span>,
        );
      }
    });
  }

  return (
    <div
      key={`selector-${triggerIndex}-${selectorIndex}`}
      className={`flex gap-[2px] items-center ${redGlowClass}`}
    >
      {elements}
    </div>
  );
};

// Render a single trigger icon based on its condition type
const renderTriggerIcon = (trigger: any, triggerIndex: number): React.ReactNode => {
  // Check if trigger has selectors (new system)
  const hasSelectors = trigger.condition?.selectors && trigger.condition.selectors.length > 0;

  // Check if trigger is city-placed condition (e.g., Tharsis Republic)
  const isCityPlaced = trigger.condition?.type === "city-placed";

  // Handle selectors first (new system with AND within selector, OR between selectors)
  if (hasSelectors) {
    const target = trigger.condition?.target || "self-player";
    const isAnyPlayer = target === "any-player";

    const redGlowClass = isAnyPlayer
      ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_2px_rgba(244,67,54,0.9))_drop-shadow(0_0_4px_rgba(244,67,54,0.7))]"
      : "";

    return (
      <div key={triggerIndex} className="flex gap-[2px] items-center">
        {trigger.condition.selectors.map((selector: any, selectorIndex: number) => (
          <React.Fragment key={`${triggerIndex}-${selectorIndex}`}>
            {selectorIndex > 0 && (
              <span className="text-[#e0e0e0] text-xs font-bold mx-[2px]">/</span>
            )}
            {renderSelector(selector, selectorIndex, triggerIndex, redGlowClass)}
          </React.Fragment>
        ))}
      </div>
    );
  }

  if (isCityPlaced) {
    const target = trigger.condition?.target || "self-player";
    const isAnyPlayer = target === "any-player";

    const redGlowClass = isAnyPlayer
      ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_2px_rgba(244,67,54,0.9))_drop-shadow(0_0_4px_rgba(244,67,54,0.7))]"
      : "";

    return (
      <div key={triggerIndex} className="flex gap-[2px] items-center justify-center">
        <div className={`flex items-center justify-center ${redGlowClass}`}>
          <GameIcon iconType="city-tile" size="small" />
        </div>
      </div>
    );
  }

  // Check if trigger is ocean-placed condition (e.g., Arctic Algae)
  const isOceanPlaced = trigger.condition?.type === "ocean-placed";

  if (isOceanPlaced) {
    const target = trigger.condition?.target || "self-player";
    const isAnyPlayer = target === "any-player";

    const redGlowClass = isAnyPlayer
      ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_2px_rgba(244,67,54,0.9))_drop-shadow(0_0_4px_rgba(244,67,54,0.7))]"
      : "";

    return (
      <div key={triggerIndex} className="flex gap-[2px] items-center justify-center">
        <div className={`flex items-center justify-center ${redGlowClass}`}>
          <GameIcon iconType="ocean-tile" size="small" />
        </div>
      </div>
    );
  }

  // Check if trigger has requiredOriginalCost (from selectors or legacy condition level)
  const requiredOriginalCost = getRequiredOriginalCost(
    trigger.condition?.selectors,
    trigger.condition?.requiredOriginalCost,
  );
  const hasRequiredOriginalCost = requiredOriginalCost !== undefined;

  if (hasRequiredOriginalCost) {
    const costReq = requiredOriginalCost;
    const hasMin = costReq.min !== undefined;
    const hasMax = costReq.max !== undefined;
    const value = (hasMin ? costReq.min : hasMax ? costReq.max : 0) ?? 0;
    const isMax = hasMax && !hasMin;

    return (
      <div key={triggerIndex} className="flex gap-[3px] items-center">
        {isMax && (
          <span className="text-xs font-semibold text-[#e0e0e0] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)] max-md:text-[11px]">
            Max
          </span>
        )}
        <GameIcon iconType="credit" amount={-value} size="small" />
      </div>
    );
  }

  // Fallback to text display for other trigger types
  return (
    <span
      key={triggerIndex}
      className="text-xs font-semibold text-[#e0e0e0] capitalize [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)] max-md:text-[11px]"
    >
      {trigger.description || trigger.type || "trigger"}
    </span>
  );
};

// Render a single behavior row (trigger : outputs)
const renderBehaviorRow = (
  behavior: any,
  rowIndex: number,
  isResourceAffordable: (resource: any, isInput: boolean) => boolean,
  analyzeResourceDisplayWithConstraints: (
    resource: any,
    availableSpace: number,
    forceCompact: boolean,
  ) => IconDisplayInfo,
  tileScaleInfo: TileScaleInfo,
): React.ReactNode => {
  // Check if this is a global-parameter-lenience effect (special case)
  const isGlobalParameterLenience =
    behavior.outputs?.some((output: any) => output.type === "global-parameter-lenience") ?? false;

  return (
    <React.Fragment key={`behavior-${rowIndex}`}>
      <div className="flex gap-[3px] items-center justify-center">
        {/* Trigger conditions - hide for global-parameter-lenience */}
        {!isGlobalParameterLenience && behavior.triggers && behavior.triggers.length > 0 && (
          <>
            <div className="flex gap-[3px] items-center">
              {(() => {
                // Check if any trigger has requiredOriginalCost (from selectors or condition level)
                const triggersWithCost = behavior.triggers.filter(
                  (trigger: any) =>
                    getRequiredOriginalCost(
                      trigger.condition?.selectors,
                      trigger.condition?.requiredOriginalCost,
                    ) !== undefined,
                );

                // If we have cost-based triggers, deduplicate and show once
                if (triggersWithCost.length > 0) {
                  // Get unique cost requirements
                  const uniqueCosts: string[] = Array.from(
                    new Set(
                      triggersWithCost.map((trigger: any) => {
                        const costReq = getRequiredOriginalCost(
                          trigger.condition?.selectors,
                          trigger.condition?.requiredOriginalCost,
                        );
                        const hasMin = costReq?.min !== undefined;
                        const hasMax = costReq?.max !== undefined;
                        const value = hasMin ? costReq?.min : costReq?.max;
                        const prefix = hasMax && !hasMin ? "Max-" : "";
                        return `${prefix}${value}`;
                      }),
                    ),
                  );

                  // Render unique cost requirements
                  return uniqueCosts.map((costKey: string, idx: number) => {
                    const isMax = costKey.startsWith("Max-");
                    const value = parseInt(costKey.replace("Max-", ""), 10);

                    return (
                      <div key={`cost-${idx}`} className="flex gap-[3px] items-center">
                        {isMax && (
                          <span className="text-xs font-semibold text-[#e0e0e0] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)] max-md:text-[11px]">
                            Max
                          </span>
                        )}
                        <GameIcon iconType="credit" amount={-value} size="small" />
                      </div>
                    );
                  });
                }

                // Otherwise, render other trigger types normally
                return behavior.triggers.map((trigger: any, triggerIndex: number) =>
                  renderTriggerIcon(trigger, triggerIndex),
                );
              })()}
            </div>
            <span className="flex items-center justify-center text-white text-base font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] min-w-[20px] z-[1]">
              :
            </span>
          </>
        )}

        {/* Outputs in same row if they fit */}
        {behavior.outputs &&
          behavior.outputs.map((output: any, index: number) => {
            const displayInfo = analyzeResourceDisplayWithConstraints(output, 6, false);
            return (
              <React.Fragment key={`triggered-output-${rowIndex}-${index}`}>
                <ResourceDisplay
                  displayInfo={displayInfo}
                  isInput={false}
                  resource={output}
                  isGroupedWithOtherNegatives={false}
                  context="default"
                  isAffordable={isResourceAffordable(output, false)}
                  tileScaleInfo={tileScaleInfo}
                />
              </React.Fragment>
            );
          })}

        {behavior.generationalEventRequirements?.length > 0 && (
          <span className="text-white font-bold text-sm ml-1">*</span>
        )}
      </div>

      {/* All choices on Row 2 - self-player plain, any-card with card background */}
      {behavior.choices && behavior.choices.length > 0 && (
        <div className="flex gap-[6px] items-center justify-center">
          {behavior.choices.map((choice: any, idx: number) => (
            <React.Fragment key={`choice-${rowIndex}-${idx}`}>
              {idx > 0 && <span className="text-[#e0e0e0] text-xs font-bold mx-[2px]">/</span>}
              {choice.outputs?.map((output: any, outputIndex: number) => {
                const displayInfo = analyzeResourceDisplayWithConstraints(output, 6, false);
                return (
                  <ResourceDisplay
                    key={`choice-${rowIndex}-${idx}-output-${outputIndex}`}
                    displayInfo={displayInfo}
                    isInput={false}
                    resource={output}
                    isGroupedWithOtherNegatives={false}
                    context="default"
                    isAffordable={isResourceAffordable(output, false)}
                    tileScaleInfo={tileScaleInfo}
                  />
                );
              })}
            </React.Fragment>
          ))}
        </div>
      )}
    </React.Fragment>
  );
};

const TriggeredEffectLayout: React.FC<TriggeredEffectLayoutProps> = ({
  behavior,
  mergedBehaviors,
  layoutPlan: _layoutPlan,
  isResourceAffordable,
  analyzeResourceDisplayWithConstraints,
  tileScaleInfo,
}) => {
  // Collect all behaviors to render (primary + merged)
  const allBehaviors = [behavior, ...(mergedBehaviors || [])];

  return (
    <div className="flex flex-col gap-2 items-center justify-center">
      {allBehaviors.map((b, index) =>
        renderBehaviorRow(
          b,
          index,
          isResourceAffordable,
          analyzeResourceDisplayWithConstraints,
          tileScaleInfo,
        ),
      )}
    </div>
  );
};

export default TriggeredEffectLayout;
