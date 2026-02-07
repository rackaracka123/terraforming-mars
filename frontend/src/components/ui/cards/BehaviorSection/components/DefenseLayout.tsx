import React from "react";
import GameIcon from "../../../display/GameIcon.tsx";

interface DefenseLayoutProps {
  behavior: any;
}

const DefenseLayout: React.FC<DefenseLayoutProps> = ({ behavior }) => {
  if (!behavior.outputs || behavior.outputs.length === 0) return null;

  const defenseOutput = behavior.outputs.find((output: any) => output.type === "defense");
  if (!defenseOutput) return null;

  const affectedResources: string[] = defenseOutput.affectedResources || [];

  return (
    <div className="flex gap-[6px] items-center">
      <span className="text-[10px] font-semibold text-white bg-[rgba(60,60,60,0.8)] px-1.5 py-0.5 rounded [text-shadow:0_0_2px_rgba(0,0,0,0.6)]">
        Protect
      </span>

      <div className="flex gap-[3px] items-center">
        {affectedResources.map((resourceType: string, index: number) => (
          <React.Fragment key={`res-${index}`}>
            {index > 0 && (
              <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
                /
              </span>
            )}
            <GameIcon iconType={resourceType} size="small" />
          </React.Fragment>
        ))}
      </div>
    </div>
  );
};

export default DefenseLayout;
