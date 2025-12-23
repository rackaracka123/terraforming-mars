import React from "react";
import GameIcon from "../../../display/GameIcon.tsx";

interface DiscountLayoutProps {
  behavior: any;
}

const DiscountLayout: React.FC<DiscountLayoutProps> = ({ behavior }) => {
  if (!behavior.outputs || behavior.outputs.length === 0) return null;

  const discountOutput = behavior.outputs.find((output: any) => output.type === "discount");
  if (!discountOutput) return null;

  const amount = Math.abs(discountOutput.amount ?? 0);
  const affectedTags = discountOutput.affectedTags || [];

  // Note: affectedStandardProjects is not rendered in BehaviorSection for now
  // Standard project discounts are shown in the standard project UI instead

  return (
    <div className="flex gap-[3px] items-center justify-center">
      {/* Left side: affected tags only */}
      <div className="flex gap-[3px] items-center">
        {/* Render affected tags - use -tag suffix to force tag icon lookup */}
        {affectedTags.map((tag: string, tagIndex: number) => (
          <React.Fragment key={`tag-${tagIndex}`}>
            {tagIndex > 0 && (
              <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
                /
              </span>
            )}
            <GameIcon iconType={`${tag.toLowerCase()}-tag`} size="small" />
          </React.Fragment>
        ))}
      </div>

      {/* Separator */}
      <span className="text-base font-bold text-white mx-[3px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
        :
      </span>

      {/* Right side: Credits icon with -amount inside */}
      <div className="relative flex items-center justify-center">
        <GameIcon iconType="credit" size="small" />
        <span className="absolute inset-0 flex items-center justify-center text-[13px] font-black font-[Prototype,Arial_Black,Arial,sans-serif] text-black [text-shadow:0_0_2px_rgba(255,255,255,0.3)] tracking-[0.5px] [-webkit-font-smoothing:antialiased] [-moz-osx-font-smoothing:grayscale] [text-rendering:optimizeLegibility] pointer-events-none max-md:text-[11px]">
          -{amount}
        </span>
      </div>
    </div>
  );
};

export default DiscountLayout;
