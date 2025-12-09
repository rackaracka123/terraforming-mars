import React, { useEffect, useRef } from "react";
import {
  GameDto,
  GameStatusActive,
  GamePhaseAction,
  ResourceTypeCredits,
  StandardProjectDto,
} from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import { canPerformActions } from "@/utils/actionUtils.ts";

interface StandardProjectsPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  onProjectSelect: (projectType: string) => void;
  gameState?: GameDto;
  anchorRef: React.RefObject<HTMLButtonElement | null>;
}

// Get tooltip message for project based on state
const getProjectTooltip = (
  project: StandardProjectDto,
  canExecuteProjects: boolean,
  isCurrentPlayerTurn: boolean,
): string => {
  if (!canExecuteProjects) {
    if (!isCurrentPlayerTurn) {
      return "Wait for your turn";
    }
    return "Actions not available in this phase";
  }

  if (!project.isAvailable && project.unavailableReasons.length > 0) {
    // Show first unavailable reason
    return project.unavailableReasons[0].message;
  }

  if (!project.isAvailable) {
    return "Not available";
  }

  return "Click to execute";
};

const StandardProjectPopover: React.FC<StandardProjectsPopoverProps> = ({
  isVisible,
  onClose,
  onProjectSelect,
  gameState,
  anchorRef,
}) => {
  const popoverRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };

    const handleClickOutside = (event: MouseEvent) => {
      if (
        popoverRef.current &&
        !popoverRef.current.contains(event.target as Node) &&
        anchorRef.current &&
        !anchorRef.current.contains(event.target as Node)
      ) {
        onClose();
      }
    };

    if (isVisible) {
      document.addEventListener("keydown", handleEscape);
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("keydown", handleEscape);
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isVisible, onClose, anchorRef]);

  if (!isVisible) return null;

  // Determine if projects can be executed
  const isGameActive = gameState?.status === GameStatusActive;
  const isActionPhase = gameState?.currentPhase === GamePhaseAction;
  const isCurrentPlayerTurn =
    gameState?.currentTurn === gameState?.viewingPlayerId;

  // Projects should be clickable only if all conditions are met
  const canExecuteProjects =
    isGameActive &&
    isActionPhase &&
    isCurrentPlayerTurn &&
    canPerformActions(gameState);

  // Get standard projects from backend
  const projects = gameState?.standardProjects || [];

  // Calculate affordable projects count
  const affordableCount = projects.filter((p) => p.isAvailable).length;

  const handleProjectClick = (project: StandardProjectDto) => {
    if (!canExecuteProjects) return;
    if (!project.isAvailable) return;
    onProjectSelect(project.type);
  };

  return (
    <div
      ref={popoverRef}
      className="fixed top-[60px] left-[20px] w-[500px] max-h-[calc(100vh-80px)] bg-space-black-darker/98 border-2 border-[#4a90e2] rounded-xl overflow-hidden shadow-[0_10px_40px_rgba(0,0,0,0.8),0_0_20px_rgba(74,144,226,0.5)] backdrop-blur-space z-[3000] animate-[popoverSlideDown_0.3s_ease-out]"
    >
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-black/40 border-b border-[#4a90e2]">
        <div className="flex items-center gap-3">
          <h2 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
            Standard Projects
          </h2>
          <div className="flex gap-2 text-xs">
            <span className="bg-[#4a90e2]/20 border border-[#4a90e2]/30 rounded px-2 py-0.5 text-white/80">
              {affordableCount}/{projects.length} Affordable
            </span>
          </div>
        </div>
        <button
          className="text-white/70 hover:text-white text-xl leading-none transition-colors"
          onClick={onClose}
        >
          ×
        </button>
      </div>

      {/* Projects List */}
      <div className="max-h-[calc(100vh-140px)] overflow-y-auto [scrollbar-width:thin] [scrollbar-color:rgba(74,144,226,0.5)_rgba(10,10,15,0.3)] p-2">
        {projects.map((project) => {
          const isExecutable = canExecuteProjects && project.isAvailable;

          return (
            <div
              key={project.id}
              className={`mb-2 last:mb-0 border rounded-lg p-3 transition-all duration-200 ${
                project.isAvailable
                  ? "border-[#4a90e2] bg-[#4a90e2]/20 hover:bg-[#4a90e2]/30"
                  : "border-[#4a90e2]/30 bg-[#4a90e2]/10 opacity-60"
              } ${isExecutable ? "cursor-pointer" : "cursor-not-allowed"}`}
              onClick={() => isExecutable && handleProjectClick(project)}
              title={getProjectTooltip(
                project,
                canExecuteProjects,
                isCurrentPlayerTurn,
              )}
            >
              <div className="flex items-start justify-between gap-3 mb-2">
                {/* Left: Name and Cost */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-2">
                    <h3 className="text-white text-sm font-bold font-orbitron m-0">
                      {project.name}
                    </h3>
                  </div>

                  <div className="flex items-center gap-2">
                    <GameIcon
                      iconType={ResourceTypeCredits}
                      amount={project.cost}
                      size="small"
                    />
                  </div>
                </div>

                {/* Right: Execute Button */}
                {canExecuteProjects && (
                  <button
                    className={`flex-shrink-0 px-3 py-1.5 rounded text-xs font-semibold transition-all ${
                      project.isAvailable
                        ? "bg-green-600/80 hover:bg-green-600 text-white shadow-sm hover:shadow-md"
                        : "bg-gray-600/50 text-gray-400 cursor-not-allowed"
                    }`}
                    onClick={(e) => {
                      e.stopPropagation();
                      if (isExecutable) handleProjectClick(project);
                    }}
                    disabled={!project.isAvailable}
                  >
                    Execute
                  </button>
                )}
              </div>

              <p className="text-white/70 text-xs leading-relaxed m-0 text-left">
                {project.description}
              </p>
            </div>
          );
        })}
      </div>

      <style>{`
        @keyframes popoverSlideDown {
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
    </div>
  );
};

export default StandardProjectPopover;
