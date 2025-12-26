import React, { useEffect, useState, useRef } from "react";
import { useMainContent } from "@/contexts/MainContentContext.tsx";
import { GameDto } from "@/types/generated/api-types.ts";

interface TopMenuBarProps {
  gameState: GameDto;
  showStandardProjectsPopover?: boolean;
  onToggleStandardProjectsPopover?: () => void;
  standardProjectsButtonRef?: React.RefObject<HTMLButtonElement | null>;
  onLeaveGame?: () => void;
  gameId?: string;
}

const TopMenuBar: React.FC<TopMenuBarProps> = ({
  gameState,
  showStandardProjectsPopover,
  onToggleStandardProjectsPopover,
  standardProjectsButtonRef,
  onLeaveGame,
  gameId,
}) => {
  const { setContentType, setContentData } = useMainContent();
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  // Close menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setMenuOpen(false);
      }
    };
    if (menuOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [menuOpen]);

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

        <div className="relative" ref={menuRef}>
          <button
            onClick={() => setMenuOpen(!menuOpen)}
            className="bg-white/10 border border-[#333] text-white p-2 rounded cursor-pointer hover:bg-white/20 transition-colors"
            aria-label="Menu"
          >
            {/* Hamburger icon */}
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
          </button>

          {menuOpen && (
            <>
              <div className="absolute right-0 top-full mt-1 bg-black/95 border border-[#444] rounded-lg shadow-lg min-w-[180px] overflow-hidden z-50 animate-[menuSlideDown_0.2s_ease-out]">
                <button
                  onClick={() => void handleCopyGameLink()}
                  className="w-full flex items-center gap-3 px-4 py-3 text-white text-sm hover:bg-white/10 transition-colors text-left"
                >
                  {/* Copy icon */}
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
                <div className="border-t border-[#333]" />
                <button
                  onClick={handleLeaveGame}
                  className="w-full flex items-center gap-3 px-4 py-3 text-red-400 text-sm hover:bg-white/10 transition-colors text-left"
                >
                  {/* Leave/exit icon */}
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
              <style>{`
                @keyframes menuSlideDown {
                  from {
                    opacity: 0;
                    transform: translateY(-10px);
                  }
                  to {
                    opacity: 1;
                    transform: translateY(0);
                  }
                }
              `}</style>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default TopMenuBar;
