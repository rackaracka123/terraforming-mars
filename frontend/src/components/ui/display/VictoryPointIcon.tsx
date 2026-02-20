import React from "react";
import { getIconPath } from "@/utils/iconStore.ts";

interface VictoryPointIconProps {
  value?: number | string;
  vpConditions?: any[];
  onHoverDescription?: (description: string | null) => void;
}

const VP_CLIP_PATH = "polygon(0 0, 100% 0, 100% calc(100% - 8px), calc(100% - 8px) 100%, 0 100%)";

const VictoryPointIcon: React.FC<VictoryPointIconProps> = ({
  value,
  vpConditions,
  onHoverDescription,
}) => {
  const vpDescription = vpConditions?.find((c: any) => c.description)?.description ?? null;

  const handleMouseEnter = () => {
    if (onHoverDescription && vpDescription) {
      onHoverDescription(vpDescription);
    }
  };

  const handleMouseLeave = () => {
    if (onHoverDescription) {
      onHoverDescription(null);
    }
  };

  const renderBox = (content: React.ReactNode) => (
    <div className="relative -mt-[2px] w-fit">
      <div
        className="inline-flex items-center gap-1 px-1.5 py-px bg-[rgba(5,5,10,0.95)] border border-[rgba(60,60,70,0.7)] border-t-0 text-white font-orbitron"
        style={{ clipPath: VP_CLIP_PATH }}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
      >
        {content}
        <span className="text-[9px] text-white/50 font-semibold tracking-wider">VP</span>
      </div>
      <svg className="absolute bottom-0 right-0 w-2 h-2 pointer-events-none" viewBox="0 0 8 8">
        <line x1="8" y1="0" x2="0" y2="8" stroke="rgba(60,60,70,0.7)" strokeWidth="1.5" />
      </svg>
    </div>
  );

  if (vpConditions && Array.isArray(vpConditions) && vpConditions.length > 0) {
    const totalConditions = vpConditions.length;

    if (totalConditions === 1) {
      const condition = vpConditions[0];

      if (condition.condition === "fixed" || condition.condition === "once") {
        if (condition.amount === 0) return null;
        return renderBox(<span className="text-[13px] font-bold">{condition.amount}</span>);
      } else if (condition.condition === "per" && condition.per) {
        const perCondition = condition.per;
        const resourceType = perCondition.tag || perCondition.type;
        const resourceIcon = resourceType ? getIconPath(resourceType) : null;
        const perAmount = perCondition.amount || 1;

        return renderBox(
          <div className="flex items-center gap-0.5">
            <span className="text-[11px] font-bold">{condition.amount}</span>
            <span className="text-[9px] text-white/40">/</span>
            {perAmount > 1 && <span className="text-[11px] font-bold">{perAmount}</span>}
            {resourceIcon && (
              <img src={resourceIcon} alt={resourceType} className="w-3.5 h-3.5 object-contain" />
            )}
          </div>,
        );
      }
    } else {
      let totalFixed = 0;
      let firstPerCondition = null;

      for (const condition of vpConditions) {
        if (condition.condition === "fixed" || condition.condition === "once") {
          totalFixed += condition.amount;
        } else if (condition.condition === "per" && !firstPerCondition) {
          firstPerCondition = condition;
        }
      }

      if (firstPerCondition && firstPerCondition.per) {
        const perCondition = firstPerCondition.per;
        const resourceType = perCondition.tag || perCondition.type;
        const resourceIcon = resourceType ? getIconPath(resourceType) : null;
        const perAmount = perCondition.amount || 1;

        return renderBox(
          <div className="flex items-center gap-0.5">
            <span className="text-[11px] font-bold">{firstPerCondition.amount}</span>
            <span className="text-[9px] text-white/40">/</span>
            {perAmount > 1 && <span className="text-[11px] font-bold">{perAmount}</span>}
            {resourceIcon && (
              <img src={resourceIcon} alt={resourceType} className="w-3.5 h-3.5 object-contain" />
            )}
          </div>,
        );
      } else if (totalFixed > 0) {
        return renderBox(<span className="text-[13px] font-bold">{totalFixed}</span>);
      }
    }

    return null;
  }

  if (value === 0 || !value) {
    return null;
  }

  return renderBox(<span className="text-[13px] font-bold">{value}</span>);
};

export default VictoryPointIcon;
