import React from "react";
import { ResourceType } from "@/types/generated/api-types.ts";
import GameIcon from "./GameIcon.tsx";

interface GameGroupIconResource {
  resourceType: ResourceType;
  amount: number;
  isAttack?: boolean;
}

interface GameGroupIconProps {
  resources: GameGroupIconResource[];
  size?: "small" | "medium" | "large";
  className?: string;
}

const GameGroupIcon: React.FC<GameGroupIconProps> = ({
  resources,
  size = "medium",
  className = "",
}) => {
  if (!resources || resources.length === 0) {
    return null;
  }

  const productionResources = resources.filter((r) =>
    r.resourceType.endsWith("-production"),
  );
  const nonProductionResources = resources.filter(
    (r) => !r.resourceType.endsWith("-production"),
  );

  const separateBySign = (
    items: GameGroupIconResource[],
  ): {
    negatives: GameGroupIconResource[];
    positives: GameGroupIconResource[];
  } => {
    return {
      negatives: items.filter((item) => item.amount < 0),
      positives: items.filter((item) => item.amount >= 0),
    };
  };

  const renderResourceIcon = (
    resource: GameGroupIconResource,
    index: number,
  ) => {
    const showAsNumber = Math.abs(resource.amount) > 2;

    if (showAsNumber) {
      return (
        <div key={index} className="flex items-center gap-0.5">
          <span className="text-[11px] font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
            {Math.abs(resource.amount)}
          </span>
          <GameIcon
            resourceType={resource.resourceType}
            isAttack={resource.isAttack}
            size={size}
          />
        </div>
      );
    } else {
      const absoluteAmount = Math.abs(resource.amount);
      return (
        <div key={index} className="flex items-center gap-px">
          {Array.from({ length: absoluteAmount }, (_, i) => (
            <GameIcon
              key={i}
              resourceType={resource.resourceType}
              isAttack={resource.isAttack}
              size={size}
            />
          ))}
        </div>
      );
    }
  };

  const renderResourceRow = (
    items: GameGroupIconResource[],
    showMinus: boolean = false,
  ) => {
    if (items.length === 0) return null;

    return (
      <div className="flex gap-[3px] items-center justify-center">
        {showMinus && (
          <span className="text-lg font-bold text-[#ffcdd2] mr-0.5 [text-shadow:1px_1px_2px_rgba(0,0,0,0.7)]">
            -
          </span>
        )}
        {items.map((resource, index) => renderResourceIcon(resource, index))}
      </div>
    );
  };

  const renderProductionGroup = () => {
    if (productionResources.length === 0) return null;

    const { negatives, positives } = separateBySign(productionResources);

    return (
      <div className="inline-flex items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-1.5 py-[3px] shadow-[0_1px_3px_rgba(0,0,0,0.2)]">
        <div className="flex flex-col gap-[3px] items-center justify-center">
          {renderResourceRow(negatives, true)}
          {renderResourceRow(positives, false)}
        </div>
      </div>
    );
  };

  const renderNonProductionGroup = () => {
    if (nonProductionResources.length === 0) return null;

    const { negatives, positives } = separateBySign(nonProductionResources);

    return (
      <div className="flex flex-col gap-[3px] items-center justify-center">
        {renderResourceRow(negatives, true)}
        {renderResourceRow(positives, false)}
      </div>
    );
  };

  return (
    <div
      className={`flex flex-wrap gap-[3px] items-center justify-center ${className}`}
    >
      {renderProductionGroup()}
      {renderNonProductionGroup()}
    </div>
  );
};

export default GameGroupIcon;
