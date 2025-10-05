import React, { useEffect, useRef } from "react";
import {
  GameDto,
  GameStatusActive,
  GamePhaseAction,
  ResourceTypeCredits,
} from "@/types/generated/api-types.ts";
import {
  StandardProject,
  STANDARD_PROJECTS,
  StandardProjectCard,
} from "@/types/cards.tsx";
import GameIcon from "../display/GameIcon.tsx";
import { canPerformActions } from "@/utils/actionUtils.ts";

interface StandardProjectsPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  onProjectSelect: (project: StandardProject) => void;
  gameState?: GameDto;
  anchorRef: React.RefObject<HTMLButtonElement | null>;
}

// Check if player can afford a standard project
const canAffordProject = (
  project: StandardProjectCard,
  credits: number,
): boolean => {
  return credits >= project.cost;
};

// Check if a standard project is available (affordability + global parameter limits)
const isProjectAvailable = (
  project: StandardProjectCard,
  gameState?: GameDto,
): boolean => {
  if (!gameState?.currentPlayer) return false;

  // Check affordability
  const canAfford = canAffordProject(
    project,
    gameState.currentPlayer.resources.credits,
  );
  if (!canAfford) return false;

  // Check global parameter limits for projects that modify them
  const globalParams = gameState.globalParameters;
  if (!globalParams) return true;

  // Extract outputs from behaviors to check for global parameter limits
  const outputs =
    project.behaviors?.[0]?.outputs?.map((o) => o.type as string) || [];

  if (outputs.includes("temperature")) {
    return globalParams.temperature < 8;
  }
  if (outputs.includes("oceans") || outputs.includes("ocean-tile")) {
    return globalParams.oceans < 9;
  }
  if (outputs.includes("oxygen") || outputs.includes("greenery-tile")) {
    return globalParams.oxygen < 14;
  }

  return true;
};

// Get tooltip message for project based on state
const getProjectTooltip = (
  project: StandardProjectCard,
  canExecuteProjects: boolean,
  isCurrentPlayerTurn: boolean,
  isAvailable: boolean,
  gameState?: GameDto,
): string => {
  if (!canExecuteProjects) {
    if (!isCurrentPlayerTurn) {
      return "Wait for your turn";
    }
    return "Actions not available in this phase";
  }

  if (!isAvailable) {
    if (
      project.cost > 0 &&
      gameState?.currentPlayer &&
      gameState.currentPlayer.resources.credits < project.cost
    ) {
      return `Need ${project.cost - gameState.currentPlayer.resources.credits} more M€`;
    }
    return "Global parameter maxed out";
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

  // Get all standard projects as array
  const projects = Object.values(STANDARD_PROJECTS);

  // Calculate affordable projects count
  const affordableCount = projects.filter((p) =>
    isProjectAvailable(p, gameState),
  ).length;

  const handleProjectClick = (project: StandardProjectCard) => {
    if (!canExecuteProjects) return;
    if (!isProjectAvailable(project, gameState)) return;
    onProjectSelect(project.id as StandardProject);
  };

  // Render effect icons from behaviors
  const renderEffects = (project: StandardProjectCard) => {
    const effects: React.ReactElement[] = [];

    if (!project.behaviors || project.behaviors.length === 0) {
      return effects;
    }

    const outputs = project.behaviors[0].outputs || [];

    outputs.forEach((output, idx) => {
      const outputType = output.type as string;
      effects.push(
        <GameIcon
          key={`output-${idx}`}
          iconType={outputType}
          amount={output.amount}
          size="small"
        />,
      );
    });

    return effects;
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
          const isAvailable = isProjectAvailable(project, gameState);
          const isExecutable = canExecuteProjects && isAvailable;
          const effects = renderEffects(project);

          return (
            <div
              key={project.id}
              className={`mb-2 last:mb-0 border rounded-lg p-3 transition-all duration-200 ${
                isAvailable
                  ? "border-[#4a90e2] bg-[#4a90e2]/20 hover:bg-[#4a90e2]/30"
                  : "border-[#4a90e2]/30 bg-[#4a90e2]/10 opacity-60"
              } ${isExecutable ? "cursor-pointer" : "cursor-not-allowed"}`}
              onClick={() => isExecutable && handleProjectClick(project)}
              title={getProjectTooltip(
                project,
                canExecuteProjects,
                isCurrentPlayerTurn,
                isAvailable,
                gameState,
              )}
            >
              <div className="flex items-start justify-between gap-3 mb-2">
                {/* Left: Name, Cost, Effects */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-2">
                    {project.icon && (
                      <div className="opacity-70">{project.icon}</div>
                    )}
                    <h3 className="text-white text-sm font-bold font-orbitron m-0">
                      {project.name}
                    </h3>
                    {project.behaviors?.[0]?.outputs?.some((o) =>
                      (o.type as string).includes("-tile"),
                    ) && (
                      <span className="text-[10px] text-white/60 bg-[#4a90e2]/30 px-1.5 py-0.5 rounded">
                        Tile
                      </span>
                    )}
                  </div>

                  <div className="flex items-center gap-2">
                    <GameIcon
                      iconType={ResourceTypeCredits}
                      amount={project.cost}
                      size="small"
                    />
                    {effects.length > 0 && (
                      <>
                        <span className="text-white/60 text-xs">→</span>
                        {effects}
                      </>
                    )}
                  </div>
                </div>

                {/* Right: Execute Button */}
                {canExecuteProjects && (
                  <button
                    className={`flex-shrink-0 px-3 py-1.5 rounded text-xs font-semibold transition-all ${
                      isAvailable
                        ? "bg-green-600/80 hover:bg-green-600 text-white shadow-sm hover:shadow-md"
                        : "bg-gray-600/50 text-gray-400 cursor-not-allowed"
                    }`}
                    onClick={(e) => {
                      e.stopPropagation();
                      if (isExecutable) handleProjectClick(project);
                    }}
                    disabled={!isAvailable}
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
