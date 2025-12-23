import { FC, useEffect, useState, useCallback } from "react";
import type { GameDto, FinalScoreDto } from "../../../types/generated/api-types";
import { TileHighlightType } from "./TileSection";

export type VPSequencePhase =
  | "intro"
  | "tr"
  | "milestones"
  | "awards"
  | "greenery"
  | "cities"
  | "cards"
  | "summary"
  | "rankings"
  | "complete";

interface PhaseConfig {
  phase: VPSequencePhase;
  duration: number; // Base duration in ms
  autoAdvance: boolean; // Whether to auto-advance after duration
}

const PHASE_SEQUENCE: PhaseConfig[] = [
  { phase: "intro", duration: 2000, autoAdvance: true },
  { phase: "tr", duration: 1500, autoAdvance: false }, // Waits for animation callback
  { phase: "milestones", duration: 2000, autoAdvance: false },
  { phase: "awards", duration: 2500, autoAdvance: false },
  { phase: "greenery", duration: 3000, autoAdvance: false },
  { phase: "cities", duration: 4000, autoAdvance: false },
  { phase: "cards", duration: 5000, autoAdvance: false },
  { phase: "summary", duration: 2000, autoAdvance: false },
  { phase: "rankings", duration: 3000, autoAdvance: false },
  { phase: "complete", duration: 0, autoAdvance: false },
];

interface VPSequenceControllerProps {
  /** Current game state */
  game: GameDto;
  /** Current player's ID */
  playerId: string;
  /** Player's final score */
  playerScore: FinalScoreDto;
  /** Callback when phase changes */
  onPhaseChange?: (phase: VPSequencePhase) => void;
  /** Callback to control tile highlighting on 3D board */
  onTileHighlight?: (type: TileHighlightType) => void;
  /** Callback when entire sequence completes */
  onSequenceComplete?: () => void;
  /** Skip animations and show final state */
  skipAnimation?: boolean;
}

interface VPSequenceState {
  currentPhase: VPSequencePhase;
  currentPhaseIndex: number;
  isPhaseAnimating: boolean;
}

/**
 * VPSequenceController - State machine for end game VP counting animation
 *
 * Manages the progression through 9 phases:
 * 1. intro - "Game Complete" title
 * 2. tr - Terraform Rating count
 * 3. milestones - Milestone badges reveal
 * 4. awards - Award placements reveal
 * 5. greenery - Greenery tile VP with board highlighting
 * 6. cities - City adjacency VP with board highlighting
 * 7. cards - Card VP reveal
 * 8. summary - Bar chart animation
 * 9. rankings - Winner podium
 */
const VPSequenceController: FC<VPSequenceControllerProps> = ({
  game: _game,
  playerId: _playerId,
  playerScore: _playerScore,
  onPhaseChange,
  onTileHighlight,
  onSequenceComplete,
  skipAnimation = false,
}) => {
  // Props prefixed with _ are available for future use but not currently used
  void _game;
  void _playerId;
  void _playerScore;
  const [state, setState] = useState<VPSequenceState>({
    currentPhase: skipAnimation ? "complete" : "intro",
    currentPhaseIndex: skipAnimation ? PHASE_SEQUENCE.length - 1 : 0,
    isPhaseAnimating: !skipAnimation,
  });

  // Advance to next phase
  const advancePhase = useCallback(() => {
    setState((prev) => {
      const nextIndex = prev.currentPhaseIndex + 1;
      if (nextIndex >= PHASE_SEQUENCE.length) {
        onSequenceComplete?.();
        return { ...prev, currentPhase: "complete", isPhaseAnimating: false };
      }

      const nextPhase = PHASE_SEQUENCE[nextIndex];
      return {
        currentPhase: nextPhase.phase,
        currentPhaseIndex: nextIndex,
        isPhaseAnimating: true,
      };
    });
  }, [onSequenceComplete]);

  // Handle phase completion callback from child components
  const handlePhaseComplete = useCallback(() => {
    // Small delay before advancing for visual polish
    setTimeout(() => {
      advancePhase();
    }, 500);
  }, [advancePhase]);

  // Notify parent of phase changes
  useEffect(() => {
    onPhaseChange?.(state.currentPhase);

    // Update tile highlighting based on phase
    const phase = state.currentPhase as VPSequencePhase;
    if (phase === "greenery") {
      onTileHighlight?.("greenery");
    } else if (phase === "cities") {
      onTileHighlight?.("city");
    } else {
      onTileHighlight?.(null);
    }
  }, [state.currentPhase, onPhaseChange, onTileHighlight]);

  // Handle auto-advancing phases
  useEffect(() => {
    if (skipAnimation) return;

    const currentConfig = PHASE_SEQUENCE[state.currentPhaseIndex];
    if (!currentConfig || !currentConfig.autoAdvance) return;

    const timer = setTimeout(() => {
      advancePhase();
    }, currentConfig.duration);

    return () => clearTimeout(timer);
  }, [state.currentPhaseIndex, skipAnimation, advancePhase]);

  // Export state and handlers for parent component
  return (
    <VPSequenceContext.Provider
      value={{
        currentPhase: state.currentPhase,
        isPhaseAnimating: state.isPhaseAnimating,
        onPhaseComplete: handlePhaseComplete,
        skipToEnd: () => {
          setState({
            currentPhase: "complete",
            currentPhaseIndex: PHASE_SEQUENCE.length - 1,
            isPhaseAnimating: false,
          });
          onTileHighlight?.(null);
          onSequenceComplete?.();
        },
      }}
    >
      {null}
    </VPSequenceContext.Provider>
  );
};

// Context for sharing sequence state with child components
import { createContext, useContext } from "react";

interface VPSequenceContextValue {
  currentPhase: VPSequencePhase;
  isPhaseAnimating: boolean;
  onPhaseComplete: () => void;
  skipToEnd: () => void;
}

const VPSequenceContext = createContext<VPSequenceContextValue | null>(null);

export const useVPSequence = () => {
  const context = useContext(VPSequenceContext);
  if (!context) {
    throw new Error("useVPSequence must be used within VPSequenceController");
  }
  return context;
};

// Hook for components to use the sequence state
export const useVPSequenceState = () => {
  const context = useContext(VPSequenceContext);
  return context;
};

export default VPSequenceController;
