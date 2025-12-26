import React from "react";
import GameIcon from "../../../display/GameIcon.tsx";

interface ValueModifierLayoutProps {
  behavior: any;
}

/**
 * ValueModifierLayout displays value modifiers like Phobolog's titanium bonus
 * or Advanced Alloys' steel/titanium bonus.
 *
 * Display format: [resource-icon] : + [credits-icon-with-amount]
 * Example: [titanium] : + [1 MC] means "Each titanium is worth 1 MC extra"
 */
const ValueModifierLayout: React.FC<ValueModifierLayoutProps> = ({ behavior }) => {
  if (!behavior.outputs || behavior.outputs.length === 0) return null;

  const valueModifierOutput = behavior.outputs.find(
    (output: any) => output.type === "value-modifier",
  );
  if (!valueModifierOutput) return null;

  const amount = valueModifierOutput.amount ?? 1;
  const affectedResources = valueModifierOutput.affectedResources || [];

  if (affectedResources.length === 0) return null;

  return (
    <div className="flex gap-[3px] items-center justify-center">
      {/* Left side: affected resources */}
      <div className="flex gap-[3px] items-center">
        {affectedResources.map((resourceType: string, resIndex: number) => (
          <React.Fragment key={`res-${resIndex}`}>
            {resIndex > 0 && (
              <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
                /
              </span>
            )}
            <GameIcon iconType={resourceType} size="small" />
          </React.Fragment>
        ))}
      </div>

      {/* Separator: colon */}
      <span className="text-base font-bold text-white mx-[3px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
        :
      </span>

      {/* Plus sign */}
      <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
        +
      </span>

      {/* Right side: Credits icon with amount inside */}
      <GameIcon iconType="credit" amount={amount} size="small" />
    </div>
  );
};

export default ValueModifierLayout;
