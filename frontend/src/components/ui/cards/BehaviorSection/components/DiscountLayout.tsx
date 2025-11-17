import React from "react";
import GameIcon from "../../../display/GameIcon.tsx";

interface DiscountLayoutProps {
  behavior: any;
}

const DiscountLayout: React.FC<DiscountLayoutProps> = ({ behavior }) => {
  if (!behavior.outputs || behavior.outputs.length === 0) return null;

  const discountOutput = behavior.outputs.find(
    (output: any) => output.type === "discount",
  );
  if (!discountOutput) return null;

  const amount = Math.abs(discountOutput.amount ?? 0);
  const affectedTags = discountOutput.affectedTags || [];
  const affectedStandardProjects =
    discountOutput.affectedStandardProjects || [];

  // Map standard project string to icon type
  const standardProjectIconMap: Record<string, string> = {
    "sell-patents": "card-draw",
    "power-plant": "energy",
    asteroid: "asteroid",
    aquifer: "ocean-tile",
    greenery: "greenery-tile",
    city: "city-tile",
    "convert-plants-to-greenery": "greenery-tile",
    "convert-heat-to-temperature": "temperature",
  };

  return (
    <div className="flex gap-[3px] items-center justify-center">
      {/* Left side: affected tags and standard projects */}
      <div className="flex gap-[3px] items-center">
        {/* Render affected tags */}
        {affectedTags.map((tag: string, tagIndex: number) => (
          <React.Fragment key={`tag-${tagIndex}`}>
            {tagIndex > 0 && (
              <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
                /
              </span>
            )}
            <GameIcon iconType={tag.toLowerCase()} size="small" />
          </React.Fragment>
        ))}

        {/* Separator between tags and standard projects */}
        {affectedTags.length > 0 && affectedStandardProjects.length > 0 && (
          <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
            /
          </span>
        )}

        {/* Render affected standard projects */}
        {affectedStandardProjects.map((project: string, projIndex: number) => {
          const iconType = standardProjectIconMap[project];
          return (
            <React.Fragment key={`proj-${projIndex}`}>
              {projIndex > 0 && (
                <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
                  /
                </span>
              )}
              {iconType ? (
                <GameIcon iconType={iconType} size="small" />
              ) : (
                <span className="text-xs font-semibold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
                  {project}
                </span>
              )}
            </React.Fragment>
          );
        })}
      </div>

      {/* Separator */}
      <span className="text-base font-bold text-white mx-[3px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
        :
      </span>

      {/* Right side: Credits icon with -amount inside */}
      <div className="relative flex items-center justify-center">
        <GameIcon iconType="credits" size="small" />
        <span className="absolute inset-0 flex items-center justify-center text-[13px] font-black font-[Prototype,Arial_Black,Arial,sans-serif] text-black [text-shadow:0_0_2px_rgba(255,255,255,0.3)] tracking-[0.5px] [-webkit-font-smoothing:antialiased] [-moz-osx-font-smoothing:grayscale] [text-rendering:optimizeLegibility] pointer-events-none max-md:text-[11px]">
          -{amount}
        </span>
      </div>
    </div>
  );
};

export default DiscountLayout;
