import React, { useEffect } from "react";
import {
  GameDto,
  GameStatusActive,
  GamePhaseAction,
} from "@/types/generated/api-types.ts";
import {
  StandardProject,
  STANDARD_PROJECTS,
  StandardProjectMetadata,
} from "@/types/cards.ts";
import CostDisplay from "../display/CostDisplay.tsx";
import ProductionDisplay from "../display/ProductionDisplay.tsx";
import { canPerformActions } from "@/utils/actionUtils.ts";

interface StandardProjectsModalProps {
  isVisible: boolean;
  onClose: () => void;
  onProjectSelect: (project: StandardProject) => void;
  gameState?: GameDto;
}

// Check if player can afford a standard project
const canAffordProject = (
  project: StandardProjectMetadata,
  credits: number,
): boolean => {
  return credits >= project.cost;
};

// Check if a standard project is available (affordability + global parameter limits)
const isProjectAvailable = (
  project: StandardProjectMetadata,
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

  switch (project.id) {
    case StandardProject.ASTEROID:
      // Temperature maxed out
      return globalParams.temperature < 8;
    case StandardProject.AQUIFER:
      // Oceans maxed out
      return globalParams.oceans < 9;
    case StandardProject.GREENERY:
      // Oxygen maxed out
      return globalParams.oxygen < 14;
    default:
      return true;
  }
};

const StandardProjectsModal: React.FC<StandardProjectsModalProps> = ({
  isVisible,
  onClose,
  onProjectSelect,
  gameState,
}) => {
  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };

    if (isVisible) {
      document.addEventListener("keydown", handleEscape);
      document.body.style.overflow = "hidden";
    }

    return () => {
      document.removeEventListener("keydown", handleEscape);
      document.body.style.overflow = "unset";
    };
  }, [isVisible, onClose]);

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

  const handleProjectClick = (project: StandardProjectMetadata) => {
    if (!canExecuteProjects) return;
    if (!isProjectAvailable(project, gameState)) return;
    onProjectSelect(project.id);
  };

  // Render effect icons with amounts
  const renderEffects = (project: StandardProjectMetadata) => {
    const effects: React.ReactElement[] = [];

    // Production effects
    if (project.effects.production) {
      project.effects.production.forEach((prod, idx) => {
        effects.push(
          <div key={`prod-${idx}`} className="flex items-center gap-1">
            <ProductionDisplay
              amount={prod.amount}
              resourceType={prod.type}
              size="small"
            />
          </div>,
        );
      });
    }

    // Global parameter effects
    if (project.effects.globalParameters) {
      project.effects.globalParameters.forEach((param, idx) => {
        const paramIcons: { [key: string]: string } = {
          temperature: "/assets/resources/heat.png",
          oxygen: "/assets/resources/oxygen.png",
          oceans: "/assets/resources/ocean.png",
        };
        effects.push(
          <div key={`param-${idx}`} className="flex items-center gap-1">
            <div className="relative w-6 h-6">
              <img
                src={paramIcons[param.type]}
                alt={param.type}
                className="w-full h-full object-contain"
              />
              <span className="absolute -bottom-1 -right-1 bg-space-black-darker border border-space-blue-500 rounded-full text-white text-[10px] font-bold px-1 min-w-[16px] text-center">
                +{param.amount === 2 ? "1" : param.amount}
              </span>
            </div>
          </div>,
        );
      });
    }

    // TR bonus
    if (project.grantsTR) {
      effects.push(
        <div key="tr" className="flex items-center gap-1">
          <div className="relative w-6 h-6">
            <img
              src="/assets/resources/tr.png"
              alt="TR"
              className="w-full h-full object-contain"
            />
            <span className="absolute -bottom-1 -right-1 bg-space-black-darker border border-space-blue-500 rounded-full text-white text-[10px] font-bold px-1 min-w-[16px] text-center">
              +1
            </span>
          </div>
        </div>,
      );
    }

    return effects;
  };

  return (
    <>
      {/* Backdrop overlay */}
      <div
        className="fixed top-0 left-0 right-0 bottom-0 z-[2999]"
        onClick={onClose}
      />

      {/* Compact dropdown positioned at bottom center */}
      <div className="fixed bottom-24 left-1/2 -translate-x-1/2 z-[3000] w-[600px] max-w-[90vw] animate-[modalSlideIn_0.2s_ease-out]">
        <div className="bg-space-black-darker/98 border-2 border-space-blue-600 rounded-xl overflow-hidden shadow-[0_10px_40px_rgba(0,0,0,0.8),0_0_20px_rgba(30,60,150,0.5)] backdrop-blur-space">
          {/* Header */}
          <div className="flex items-center justify-between px-4 py-3 bg-black/40 border-b border-space-blue-600">
            <div className="flex items-center gap-3">
              <h2 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
                Standard Projects
              </h2>
              <div className="flex gap-2 text-xs">
                <span className="bg-space-blue-600/30 border border-space-blue-600/50 rounded px-2 py-0.5 text-white/80">
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
          <div className="max-h-[60vh] overflow-y-auto [scrollbar-width:thin] [scrollbar-color:rgba(30,60,150,0.5)_rgba(10,10,15,0.3)] p-2">
            {projects.map((project) => {
              const isAvailable = isProjectAvailable(project, gameState);
              const isExecutable = canExecuteProjects && isAvailable;
              const effects = renderEffects(project);

              return (
                <div
                  key={project.id}
                  className={`mb-2 last:mb-0 border rounded-lg p-3 transition-all duration-200 ${
                    isAvailable
                      ? "border-space-blue-600 bg-space-blue-900/30 hover:bg-space-blue-900/40"
                      : "border-space-blue-600/30 bg-space-blue-900/10 opacity-60"
                  } ${isExecutable ? "cursor-pointer" : "cursor-not-allowed"}`}
                  onClick={() => isExecutable && handleProjectClick(project)}
                  title={
                    !canExecuteProjects
                      ? !isCurrentPlayerTurn
                        ? "Wait for your turn"
                        : "Actions not available in this phase"
                      : !isAvailable
                        ? project.cost > 0 &&
                          gameState?.currentPlayer &&
                          gameState.currentPlayer.resources.credits <
                            project.cost
                          ? `Need ${project.cost - gameState.currentPlayer.resources.credits} more M€`
                          : "Global parameter maxed out"
                        : "Click to execute"
                  }
                >
                  <div className="flex items-start justify-between gap-3">
                    {/* Left: Name, Cost, Effects */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-2">
                        {project.icon && (
                          <img
                            src={project.icon}
                            alt={project.name}
                            className="w-5 h-5 object-contain opacity-70"
                          />
                        )}
                        <h3 className="text-white text-sm font-bold font-orbitron m-0">
                          {project.name}
                        </h3>
                        {project.requiresTilePlacement && (
                          <span className="text-[10px] text-white/60 bg-space-blue-600/30 px-1.5 py-0.5 rounded">
                            Tile
                          </span>
                        )}
                      </div>

                      <div className="flex items-center gap-3 mb-2">
                        <CostDisplay cost={project.cost} size="small" />
                        {effects.length > 0 && (
                          <div className="flex items-center gap-2">
                            <span className="text-white/60 text-xs">→</span>
                            <div className="flex gap-2">{effects}</div>
                          </div>
                        )}
                      </div>

                      <p className="text-white/70 text-xs leading-relaxed m-0">
                        {project.description}
                      </p>
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
                </div>
              );
            })}
          </div>

          {/* Footer */}
          <div className="px-4 py-2 bg-black/30 border-t border-space-blue-600/30 text-center text-white/50 text-[10px]">
            Press ESC or click outside to close
          </div>
        </div>
      </div>
    </>
  );
};

export default StandardProjectsModal;
