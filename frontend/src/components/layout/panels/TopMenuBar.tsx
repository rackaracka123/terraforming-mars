import React, { useEffect, useState, useRef } from "react";
import { GameDto, PlayerDto } from "@/types/generated/api-types.ts";
import { StandardProject } from "@/types/cards.tsx";
import SoundToggleButton from "../../ui/buttons/SoundToggleButton.tsx";
import StandardProjectPopover from "../../ui/popover/StandardProjectPopover.tsx";
import MilestonePopover from "../../ui/popover/MilestonePopover.tsx";
import AwardPopover from "../../ui/popover/AwardPopover.tsx";
import { GamePopover } from "../../ui/GamePopover";
import GameMenuButton from "../../ui/buttons/GameMenuButton.tsx";

interface TopMenuBarProps {
  gameState: GameDto;
  currentPlayer?: PlayerDto | null;
  onStandardProjectSelect?: (project: StandardProject) => void;
  onLeaveGame?: () => void;
  gameId?: string;
}

const TopMenuBar: React.FC<TopMenuBarProps> = ({
  gameState,
  currentPlayer,
  onStandardProjectSelect,
  onLeaveGame,
  gameId,
}) => {
  const [menuOpen, setMenuOpen] = useState(false);
  const hamburgerButtonRef = useRef<HTMLButtonElement>(null);

  const [showStandardProjectsPopover, setShowStandardProjectsPopover] = useState(false);
  const [showMilestonePopover, setShowMilestonePopover] = useState(false);
  const [showAwardPopover, setShowAwardPopover] = useState(false);
  const standardProjectsButtonRef = useRef<HTMLButtonElement>(null);
  const milestonesButtonRef = useRef<HTMLButtonElement>(null);
  const awardsButtonRef = useRef<HTMLButtonElement>(null);

  const handleCopyGameLink = async () => {
    if (gameId) {
      const url = `${window.location.origin}/game/${gameId}`;
      await navigator.clipboard.writeText(url);
      setMenuOpen(false);
    }
  };

  const handleLeaveGame = () => {
    setMenuOpen(false);
    onLeaveGame?.();
  };

  const handleStandardProjectSelect = (project: StandardProject) => {
    setShowStandardProjectsPopover(false);
    onStandardProjectSelect?.(project);
  };

  // Reset button inline border style when popovers close
  useEffect(() => {
    if (!showStandardProjectsPopover && standardProjectsButtonRef.current) {
      standardProjectsButtonRef.current.style.borderColor = "rgba(255,255,255,0.2)";
    }
  }, [showStandardProjectsPopover]);

  useEffect(() => {
    if (!showMilestonePopover && milestonesButtonRef.current) {
      milestonesButtonRef.current.style.borderColor = "rgba(255,255,255,0.2)";
    }
  }, [showMilestonePopover]);

  useEffect(() => {
    if (!showAwardPopover && awardsButtonRef.current) {
      awardsButtonRef.current.style.borderColor = "rgba(255,255,255,0.2)";
    }
  }, [showAwardPopover]);

  const menuItems = [
    { id: "projects" as const, label: "STANDARD PROJECTS", color: "#4a90e2" },
    { id: "milestones" as const, label: "MILESTONES", color: "#ff6b35" },
    { id: "awards" as const, label: "AWARDS", color: "#f39c12" },
  ];

  const handleTabClick = (tabId: "milestones" | "projects" | "awards") => {
    if (currentPlayer?.pendingTileSelection) return;

    if (tabId === "projects") {
      setShowStandardProjectsPopover((prev) => !prev);
    } else if (tabId === "milestones") {
      setShowMilestonePopover((prev) => !prev);
    } else if (tabId === "awards") {
      setShowAwardPopover((prev) => !prev);
    }
  };

  // Get the appropriate ref for each button
  const getButtonRef = (itemId: "projects" | "milestones" | "awards") => {
    if (itemId === "projects") return standardProjectsButtonRef;
    if (itemId === "milestones") return milestonesButtonRef;
    if (itemId === "awards") return awardsButtonRef;
    return null;
  };

  // Check if a popover is currently open for a given item
  const isPopoverOpen = (itemId: "projects" | "milestones" | "awards") => {
    if (itemId === "projects") return showStandardProjectsPopover;
    if (itemId === "milestones") return showMilestonePopover;
    if (itemId === "awards") return showAwardPopover;
    return false;
  };

  return (
    <div className="bg-transparent relative z-[100] pointer-events-none">
      <div className="flex justify-between items-center px-5 h-[60px] max-lg:px-[15px] max-lg:h-[50px] max-md:px-2.5 max-md:flex-wrap max-sm:px-2.5 max-sm:flex-wrap">
        <div className="flex gap-3 max-md:order-2 max-md:flex-[0_0_100%] max-md:mt-2.5">
          {menuItems.map((item) => (
            <GameMenuButton
              key={item.id}
              ref={getButtonRef(item.id)}
              variant="toolbar"
              size="sm"
              onClick={() => handleTabClick(item.id)}
              style={{ "--item-color": item.color } as React.CSSProperties}
              onMouseEnter={(e) => (e.currentTarget.style.borderColor = item.color)}
              onMouseLeave={(e) => {
                if (!isPopoverOpen(item.id)) {
                  e.currentTarget.style.borderColor = "rgba(255,255,255,0.2)";
                }
              }}
              className="pointer-events-auto max-lg:text-xs max-lg:py-2 max-lg:px-[15px] max-md:py-2 max-md:px-[15px] max-md:text-xs max-sm:py-1.5 max-sm:px-3 max-sm:text-[11px]"
            >
              {item.label}
            </GameMenuButton>
          ))}
        </div>

        <GameMenuButton
          ref={hamburgerButtonRef}
          variant="toolbar"
          onClick={() => setMenuOpen(!menuOpen)}
          aria-label="Menu"
          className="pointer-events-auto !p-2"
        >
          <svg
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
          >
            <line x1="3" y1="6" x2="21" y2="6" />
            <line x1="3" y1="12" x2="21" y2="12" />
            <line x1="3" y1="18" x2="21" y2="18" />
          </svg>
        </GameMenuButton>

        <GamePopover
          isVisible={menuOpen}
          onClose={() => setMenuOpen(false)}
          position={{ type: "anchor", anchorRef: hamburgerButtonRef, placement: "below" }}
          theme="menu"
          width={200}
          maxHeight="auto"
          animation="slideDown"
          excludeRef={hamburgerButtonRef}
        >
          <div className="py-1">
            <button
              onClick={() => void handleCopyGameLink()}
              className="w-full flex items-center gap-3 px-4 py-3 text-white text-sm hover:bg-white/10 transition-colors text-left"
            >
              <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
              </svg>
              Copy game link
            </button>
            <div className="border-t border-[#333] mx-2" />
            <SoundToggleButton />
            <div className="border-t border-[#333] mx-2" />
            <button
              onClick={handleLeaveGame}
              className="w-full flex items-center gap-3 px-4 py-3 text-red-400 text-sm hover:bg-white/10 transition-colors text-left"
            >
              <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                <polyline points="16 17 21 12 16 7" />
                <line x1="21" y1="12" x2="9" y2="12" />
              </svg>
              Leave game
            </button>
          </div>
        </GamePopover>
      </div>

      <StandardProjectPopover
        isVisible={showStandardProjectsPopover}
        onClose={() => setShowStandardProjectsPopover(false)}
        onProjectSelect={handleStandardProjectSelect}
        gameState={gameState}
        anchorRef={standardProjectsButtonRef}
      />

      <MilestonePopover
        isVisible={showMilestonePopover}
        onClose={() => setShowMilestonePopover(false)}
        gameState={gameState}
        anchorRef={milestonesButtonRef}
      />

      <AwardPopover
        isVisible={showAwardPopover}
        onClose={() => setShowAwardPopover(false)}
        gameState={gameState}
        anchorRef={awardsButtonRef}
      />
    </div>
  );
};

export default TopMenuBar;
