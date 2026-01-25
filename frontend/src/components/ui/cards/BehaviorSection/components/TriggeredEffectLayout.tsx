import React from "react";
import ResourceDisplay from "./ResourceDisplay.tsx";
import GameIcon from "../../../display/GameIcon.tsx";
import { CardBehaviorDto } from "@/types/generated/api-types.ts";

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

// Render a single trigger icon based on its condition type
const renderTriggerIcon = (trigger: any, triggerIndex: number): React.ReactNode => {
  // Check if trigger has condition with affectedResources (e.g., placement-bonus-gained)
  const hasAffectedResources =
    trigger.condition?.affectedResources && trigger.condition.affectedResources.length > 0;

  // Check if trigger has condition with affectedCardTypes (e.g., card-played with event filter)
  const hasAffectedCardTypes =
    trigger.condition?.affectedCardTypes && trigger.condition.affectedCardTypes.length > 0;

  // Check if trigger has condition with affectedTags (e.g., card-played with tag filter)
  const hasAffectedTags =
    trigger.condition?.affectedTags && trigger.condition.affectedTags.length > 0;

  // Check if trigger is city-placed condition (e.g., Tharsis Republic)
  const isCityPlaced = trigger.condition?.type === "city-placed";

  if (isCityPlaced) {
    // Render city tile icon for city-placed trigger
    // Red glow for self-player (like attack indicators), asterisk (*) for any-player
    const target = trigger.condition?.target || "self-player";
    const isSelfPlayer = target === "self-player";

    // Red glow filter matching the attack indicator style (without pulse animation)
    const redGlowClass = isSelfPlayer
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

  if (hasAffectedTags) {
    // Render icons for affected tags (e.g., plant / animal)
    // Use -tag suffix to force tag icon lookup (otherwise resource icons are shown)
    return (
      <div key={triggerIndex} className="flex gap-[2px] items-center">
        {trigger.condition.affectedTags.map((tag: string, tagIndex: number) => (
          <React.Fragment key={`${triggerIndex}-${tagIndex}`}>
            {tagIndex > 0 && <span className="text-[#e0e0e0] text-xs font-bold mx-[2px]">/</span>}
            <GameIcon iconType={`${tag}-tag`} size="small" />
          </React.Fragment>
        ))}
      </div>
    );
  }

  if (hasAffectedResources) {
    // Render icons for affected resources (e.g., steel / titanium)
    return (
      <div key={triggerIndex} className="flex gap-[2px] items-center">
        {trigger.condition.affectedResources.map((resource: string, resIndex: number) => (
          <React.Fragment key={`${triggerIndex}-${resIndex}`}>
            {resIndex > 0 && <span className="text-[#e0e0e0] text-xs font-bold mx-[2px]">/</span>}
            <GameIcon iconType={resource} size="small" />
          </React.Fragment>
        ))}
      </div>
    );
  }

  if (hasAffectedCardTypes) {
    // Render icons/text for affected card types (e.g., event card icon)
    return (
      <div key={triggerIndex} className="flex gap-[2px] items-center">
        {trigger.condition.affectedCardTypes.map((cardType: string, typeIndex: number) => (
          <React.Fragment key={`${triggerIndex}-${typeIndex}`}>
            {typeIndex > 0 && <span className="text-[#e0e0e0] text-xs font-bold mx-[2px]">/</span>}
            {cardType === "event" ? (
              <GameIcon iconType="event" size="small" />
            ) : (
              <span className="text-xs font-semibold text-[#e0e0e0] capitalize [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)] max-md:text-[11px]">
                {cardType}
              </span>
            )}
          </React.Fragment>
        ))}
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
    <div key={`behavior-row-${rowIndex}`} className="flex gap-[3px] items-center justify-center">
      {/* Trigger conditions - hide for global-parameter-lenience */}
      {!isGlobalParameterLenience && behavior.triggers && behavior.triggers.length > 0 && (
        <>
          <div className="flex gap-[3px] items-center">
            {(() => {
              // Check if any trigger has requiredOriginalCost
              const triggersWithCost = behavior.triggers.filter(
                (trigger: any) => trigger.condition?.requiredOriginalCost !== undefined,
              );

              // If we have cost-based triggers, deduplicate and show once
              if (triggersWithCost.length > 0) {
                // Get unique cost requirements
                const uniqueCosts: string[] = Array.from(
                  new Set(
                    triggersWithCost.map((trigger: any) => {
                      const costReq = trigger.condition.requiredOriginalCost;
                      const hasMin = costReq.min !== undefined;
                      const hasMax = costReq.max !== undefined;
                      const value = hasMin ? costReq.min : costReq.max;
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
