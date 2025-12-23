import React, { useEffect } from "react";
import { useMainContent } from "@/contexts/MainContentContext.tsx";
import { GameDto } from "@/types/generated/api-types.ts";

interface TopMenuBarProps {
  gameState: GameDto;
  showStandardProjectsPopover?: boolean;
  onToggleStandardProjectsPopover?: () => void;
  standardProjectsButtonRef?: React.RefObject<HTMLButtonElement | null>;
}

const TopMenuBar: React.FC<TopMenuBarProps> = ({
  gameState,
  showStandardProjectsPopover,
  onToggleStandardProjectsPopover,
  standardProjectsButtonRef,
}) => {
  const { setContentType, setContentData } = useMainContent();

  // Reset button inline border style when popover closes
  useEffect(() => {
    if (!showStandardProjectsPopover && standardProjectsButtonRef?.current) {
      standardProjectsButtonRef.current.style.borderColor = "transparent";
    }
  }, [showStandardProjectsPopover, standardProjectsButtonRef]);

  const menuItems = [
    { id: "milestones" as const, label: "MILESTONES", color: "#ff6b35" },
    { id: "projects" as const, label: "STANDARD PROJECTS", color: "#4a90e2" },
    { id: "awards" as const, label: "AWARDS", color: "#f39c12" },
  ];

  // Transform player-specific milestones to display format (includes eligibility from backend)
  const getMilestoneData = () => ({
    milestones: gameState.currentPlayer.milestones.map((m) => ({
      id: m.type,
      name: m.name,
      description: m.description,
      reward: "5 VP",
      cost: m.claimCost,
      claimed: m.isClaimed,
      claimedBy: m.claimedBy,
      available: m.available, // Backend-calculated eligibility
      progress: m.progress,
      required: m.required,
    })),
  });

  // Transform player-specific awards to display format (includes eligibility from backend)
  const getAwardData = () => ({
    awards: gameState.currentPlayer.awards.map((a) => ({
      id: a.type,
      name: a.name,
      description: a.description,
      fundingCost: a.fundingCost,
      funded: a.isFunded,
      fundedBy: a.fundedBy,
      available: a.available, // Backend-calculated eligibility
    })),
  });

  const handleTabClick = (tabId: "milestones" | "projects" | "awards") => {
    // For standard projects, toggle the popover
    if (tabId === "projects") {
      onToggleStandardProjectsPopover?.();
      return;
    }

    const data = tabId === "milestones" ? getMilestoneData() : getAwardData();
    setContentData(data);
    setContentType(tabId);
  };

  return (
    <div className="bg-black/95 border-b border-[#333] relative z-[100]">
      <div className="flex justify-between items-center px-5 h-[60px] max-lg:px-[15px] max-lg:h-[50px] max-md:px-2.5 max-md:flex-wrap max-sm:px-2.5 max-sm:flex-wrap">
        <div className="flex gap-5 max-md:order-2 max-md:flex-[0_0_100%] max-md:mt-2.5">
          {menuItems.map((item) => (
            <button
              key={item.id}
              ref={item.id === "projects" ? standardProjectsButtonRef : null}
              className={`bg-none border-2 text-white text-sm font-bold py-2.5 px-5 cursor-pointer rounded transition-all duration-200 hover:bg-white/10 max-lg:text-xs max-lg:py-2 max-lg:px-[15px] max-md:py-2 max-md:px-[15px] max-md:text-xs max-sm:py-1.5 max-sm:px-3 max-sm:text-[11px] ${item.id === "projects" && showStandardProjectsPopover ? `border-[${item.color}]` : "border-transparent"}`}
              onClick={() => handleTabClick(item.id)}
              style={{ "--item-color": item.color } as React.CSSProperties}
              onMouseEnter={(e) => (e.currentTarget.style.borderColor = item.color)}
              onMouseLeave={(e) => {
                if (item.id !== "projects" || !showStandardProjectsPopover) {
                  e.currentTarget.style.borderColor = "transparent";
                }
              }}
            >
              {item.label}
            </button>
          ))}
        </div>

        <div className="flex gap-2.5 max-md:gap-2 max-sm:flex-col max-sm:gap-1">
          <button className="bg-white/10 border border-[#333] text-white py-2 px-3 rounded cursor-pointer text-xs hover:bg-white/20 max-lg:py-1.5 max-lg:px-2.5 max-lg:text-[11px] max-md:py-1.5 max-md:px-2.5 max-md:text-[11px] max-sm:py-1 max-sm:px-1.5 max-sm:text-[9px]">
            ⚙️ Settings
          </button>
          <button className="bg-white/10 border border-[#333] text-white py-2 px-3 rounded cursor-pointer text-xs hover:bg-white/20 max-lg:py-1.5 max-lg:px-2.5 max-lg:text-[11px] max-md:py-1.5 max-md:px-2.5 max-md:text-[11px] max-sm:py-1 max-sm:px-1.5 max-sm:text-[9px]">
            📊 Stats
          </button>
        </div>
      </div>
    </div>
  );
};

export default TopMenuBar;
