import React, { useEffect, useRef } from "react";
import {
  GameDto,
  GameStatusActive,
  GamePhaseAction,
  ResourceTypeCredit,
  PlayerStandardProjectDto,
} from "@/types/generated/api-types.ts";
import { StandardProject, STANDARD_PROJECTS } from "@/types/cards.tsx";
import GameIcon from "../display/GameIcon.tsx";
import { canPerformActions } from "@/utils/actionUtils.ts";

interface StandardProjectsPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  onProjectSelect: (project: StandardProject) => void;
  gameState?: GameDto;
  anchorRef: React.RefObject<HTMLButtonElement | null>;
}

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

  // Get standard projects from backend player state
  const playerProjects = gameState?.currentPlayer?.standardProjects ?? [];

  // Calculate available projects count (using backend state)
  const availableCount = playerProjects.filter((p) => p.available).length;

  const handleProjectClick = (project: PlayerStandardProjectDto) => {
    if (!canExecuteProjects) return;
    if (!project.available) return;
    onProjectSelect(project.projectType as StandardProject);
  };

  // Render effect icons from behaviors (use static definitions for display)
  const renderEffects = (projectType: string) => {
    const effects: React.ReactElement[] = [];
    const staticProject = STANDARD_PROJECTS[projectType as StandardProject];

    if (!staticProject?.behaviors || staticProject.behaviors.length === 0) {
      return effects;
    }

    const outputs = staticProject.behaviors[0].outputs || [];

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

  // Get static project info for display (name, description, icon)
  const getStaticProjectInfo = (projectType: string) => {
    return STANDARD_PROJECTS[projectType as StandardProject];
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
              {availableCount}/{playerProjects.length} Available
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
        {playerProjects.map((project) => {
          const staticInfo = getStaticProjectInfo(project.projectType);
          const isExecutable = canExecuteProjects && project.available;
          const effects = renderEffects(project.projectType);

          // Get effective cost (backend calculates discounts)
          const effectiveCreditCost = project.effectiveCost["credit"] ?? 0;
          const baseCreditCost = project.baseCost["credit"] ?? 0;
          const hasDiscount = effectiveCreditCost < baseCreditCost;

          return (
            <div
              key={project.projectType}
              className={`relative mb-2 last:mb-0 border rounded-lg p-3 transition-all duration-200 ${
                project.available
                  ? "border-[#4a90e2] bg-[#4a90e2]/20 hover:bg-[#4a90e2]/30"
                  : "border-[#4a90e2]/30 bg-[#4a90e2]/10 opacity-60"
              }`}
              onClick={() => isExecutable && handleProjectClick(project)}
            >
              {/* Unavailable indicator (shows when project cannot be executed) */}
              {!project.available &&
                project.errors &&
                project.errors.length > 0 && (
                  <div className="absolute top-2 right-2 z-[4] bg-[linear-gradient(135deg,#e74c3c,#c0392b)] text-white text-[9px] font-bold px-2 py-1 rounded border border-[rgba(231,76,60,0.8)] shadow-[0_2px_8px_rgba(231,76,60,0.4)] flex items-center gap-1">
                    <span>⚠</span>
                    <span
                      className="max-w-[140px] truncate"
                      title={project.errors.map((e) => e.message).join(", ")}
                    >
                      {project.errors[0].message}
                      {project.errors.length > 1 &&
                        ` (+${project.errors.length - 1})`}
                    </span>
                  </div>
                )}

              <div className="flex items-start justify-between gap-3 mb-2">
                {/* Left: Name, Cost, Effects */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-2">
                    {staticInfo?.icon && (
                      <div className="opacity-70">{staticInfo.icon}</div>
                    )}
                    <h3 className="text-white text-sm font-bold font-orbitron m-0">
                      {staticInfo?.name ?? project.projectType}
                    </h3>
                    {staticInfo?.behaviors?.[0]?.outputs?.some((o) =>
                      (o.type as string).includes("-tile"),
                    ) && (
                      <span className="text-[10px] text-white/60 bg-[#4a90e2]/30 px-1.5 py-0.5 rounded">
                        Tile
                      </span>
                    )}
                  </div>

                  <div className="flex items-center gap-2">
                    {/* Show cost with discount in card style: [original greyed] -> [discounted] */}
                    {hasDiscount ? (
                      <div className="flex items-center gap-1">
                        {/* Original cost (greyed out) */}
                        <div className="grayscale-[0.7] flex items-center">
                          <GameIcon
                            iconType={ResourceTypeCredit}
                            amount={baseCreditCost}
                            size="small"
                          />
                        </div>
                        {/* Right arrow */}
                        <svg
                          width="10"
                          height="8"
                          viewBox="0 0 10 8"
                          className="opacity-70 mx-0.5 flex-shrink-0"
                        >
                          <path
                            d="M10 4 L4 0 L4 2 L0 2 L0 6 L4 6 L4 8 Z"
                            fill="rgba(76, 175, 80, 0.9)"
                          />
                        </svg>
                        {/* Discounted cost (clear) */}
                        <div className="flex items-center">
                          <GameIcon
                            iconType={ResourceTypeCredit}
                            amount={effectiveCreditCost}
                            size="small"
                          />
                        </div>
                      </div>
                    ) : (
                      <GameIcon
                        iconType={ResourceTypeCredit}
                        amount={effectiveCreditCost}
                        size="small"
                      />
                    )}
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
                      project.available
                        ? "bg-green-600/80 hover:bg-green-600 text-white shadow-sm hover:shadow-md cursor-pointer"
                        : "bg-gray-600/50 text-gray-400"
                    }`}
                    onClick={(e) => {
                      e.stopPropagation();
                      if (isExecutable) handleProjectClick(project);
                    }}
                    disabled={!project.available}
                  >
                    Execute
                  </button>
                )}
              </div>

              <p className="text-white/70 text-xs leading-relaxed m-0 text-left">
                {staticInfo?.description ?? ""}
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
