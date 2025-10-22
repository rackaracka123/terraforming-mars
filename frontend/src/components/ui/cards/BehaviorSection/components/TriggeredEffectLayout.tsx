import React from "react";
import ResourceDisplay from "./ResourceDisplay.tsx";
import GameIcon from "../../../display/GameIcon.tsx";

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
  layoutPlan: LayoutPlan;
  isResourceAffordable: (resource: any, isInput: boolean) => boolean;
  analyzeResourceDisplayWithConstraints: (
    resource: any,
    availableSpace: number,
    forceCompact: boolean,
  ) => IconDisplayInfo;
  tileScaleInfo: TileScaleInfo;
}

const TriggeredEffectLayout: React.FC<TriggeredEffectLayoutProps> = ({
  behavior,
  layoutPlan: _layoutPlan,
  isResourceAffordable,
  analyzeResourceDisplayWithConstraints,
  tileScaleInfo,
}) => {
  // Check if this is a global-parameter-lenience effect (special case)
  const isGlobalParameterLenience =
    behavior.outputs?.some(
      (output: any) => output.type === "global-parameter-lenience",
    ) ?? false;

  return (
    <div className="flex flex-col gap-[3px] items-center justify-center">
      <div className="flex gap-[3px] items-center justify-center">
        {/* Trigger conditions - hide for global-parameter-lenience */}
        {!isGlobalParameterLenience &&
          behavior.triggers &&
          behavior.triggers.length > 0 && (
            <>
              <div className="flex gap-[3px] items-center">
                {behavior.triggers.map((trigger: any, triggerIndex: number) => {
                  // Check if trigger has condition with affectedResources (e.g., placement-bonus-gained)
                  const hasAffectedResources =
                    trigger.condition?.affectedResources &&
                    trigger.condition.affectedResources.length > 0;

                  if (hasAffectedResources) {
                    // Render icons for affected resources (e.g., steel / titanium)
                    return (
                      <div
                        key={triggerIndex}
                        className="flex gap-[2px] items-center"
                      >
                        {trigger.condition.affectedResources.map(
                          (resource: string, resIndex: number) => (
                            <React.Fragment key={`${triggerIndex}-${resIndex}`}>
                              {resIndex > 0 && (
                                <span className="text-[#e0e0e0] text-xs font-bold mx-[2px]">
                                  /
                                </span>
                              )}
                              <GameIcon iconType={resource} size="small" />
                            </React.Fragment>
                          ),
                        )}
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
                })}
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
      </div>
    </div>
  );
};

export default TriggeredEffectLayout;
