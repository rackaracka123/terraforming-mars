import React, { useEffect, useRef, useState } from "react";
import { useMainContent } from "@/contexts/MainContentContext.tsx";
import { GameDto } from "@/types/generated/api-types.ts";
import SettingsDropdown from "../../ui/dropdown/SettingsDropdown.tsx";

interface TopMenuBarProps {
  gameState?: GameDto | null;
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
  const [showSettingsDropdown, setShowSettingsDropdown] = useState(false);
  const settingsButtonRef = useRef<HTMLButtonElement>(null);

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
    { id: "debug" as const, label: "ADMIN TOOLS", color: "#9b59b6" },
  ];

  // Mock data for different content types - normally this would come from game state
  const getMockData = (
    type: "milestones" | "projects" | "awards" | "debug",
  ) => {
    if (type === "milestones") {
      return {
        milestones: [
          {
            id: "terraformer",
            name: "Terraformer",
            description: "Have a terraform rating of at least 35",
            reward: "5 VP",
            cost: 8,
            claimed: false,
            available: true,
          },
          {
            id: "mayor",
            name: "Mayor",
            description: "Own at least 3 city tiles",
            reward: "5 VP",
            cost: 8,
            claimed: true,
            claimedBy: "Alice Chen",
            available: false,
          },
          {
            id: "gardener",
            name: "Gardener",
            description: "Own at least 3 greenery tiles",
            reward: "5 VP",
            cost: 8,
            claimed: false,
            available: true,
          },
        ],
      };
    }

    if (type === "awards") {
      return {
        awards: [
          {
            id: "landlord",
            name: "Landlord",
            description: "Most tiles on Mars",
            fundingCost: 8,
            funded: true,
            fundedBy: "Bob Martinez",
            winner: "Alice Chen",
            available: false,
          },
          {
            id: "banker",
            name: "Banker",
            description: "Highest M€ production",
            fundingCost: 8,
            funded: false,
            available: true,
          },
          {
            id: "scientist",
            name: "Scientist",
            description: "Most science tags",
            fundingCost: 8,
            funded: true,
            fundedBy: "Carol Kim",
            available: false,
          },
        ],
      };
    }

    return {};
  };

  const handleTabClick = (
    tabId: "milestones" | "projects" | "awards" | "debug",
  ) => {
    // For debug, we'll emit a custom event instead of using content context
    if (tabId === "debug") {
      window.dispatchEvent(new CustomEvent("toggle-debug-dropdown"));
      return;
    }
    // For standard projects, toggle the popover
    if (tabId === "projects") {
      onToggleStandardProjectsPopover?.();
      return;
    }
    const data = getMockData(tabId);
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
              onMouseEnter={(e) =>
                (e.currentTarget.style.borderColor = item.color)
              }
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
          <button
            ref={settingsButtonRef}
            onClick={() => setShowSettingsDropdown(!showSettingsDropdown)}
            className={`bg-white/10 border text-white py-2 px-3 rounded cursor-pointer text-xs hover:bg-white/20 max-lg:py-1.5 max-lg:px-2.5 max-lg:text-[11px] max-md:py-1.5 max-md:px-2.5 max-md:text-[11px] max-sm:py-1 max-sm:px-1.5 max-sm:text-[9px] ${
              showSettingsDropdown ? "border-space-blue-400" : "border-[#333]"
            }`}
          >
            ⚙️ Settings
          </button>
          <button className="bg-white/10 border border-[#333] text-white py-2 px-3 rounded cursor-pointer text-xs hover:bg-white/20 max-lg:py-1.5 max-lg:px-2.5 max-lg:text-[11px] max-md:py-1.5 max-md:px-2.5 max-md:text-[11px] max-sm:py-1 max-sm:px-1.5 max-sm:text-[9px]">
            📊 Stats
          </button>
        </div>
      </div>

      {/* Dev Mode Chip */}
      {gameState?.settings?.developmentMode && (
        <div className="absolute top-full left-1/2 -translate-x-1/2 bg-[#ff6b35] text-white text-[10px] font-bold py-1 px-3 rounded-b-lg border border-[#e55a2e] border-t-0 z-[99] whitespace-nowrap shadow-[0_2px_4px_rgba(0,0,0,0.3)]">
          DEV MODE
        </div>
      )}

      {/* Settings Dropdown */}
      <SettingsDropdown
        isVisible={showSettingsDropdown}
        onClose={() => setShowSettingsDropdown(false)}
        buttonRef={settingsButtonRef}
      />
    </div>
  );
};

export default TopMenuBar;
