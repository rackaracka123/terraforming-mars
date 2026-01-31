import React from "react";
import LeftSidebar from "../panels/LeftSidebar.tsx";
import TopMenuBar from "../panels/TopMenuBar.tsx";
import RightSidebar from "../panels/RightSidebar.tsx";
import MainContentDisplay from "../../ui/display/MainContentDisplay.tsx";
import { TileHighlightMode } from "../../game/board/ProjectedHexTile.tsx";
import { TileVPIndicator } from "../../ui/overlay/EndGameOverlay.tsx";
import BottomResourceBar from "../../ui/overlay/BottomResourceBar.tsx";
import PlayerOverlay from "../../ui/overlay/PlayerOverlay.tsx";
import CorporationViewer from "../../ui/display/CorporationViewer.tsx";
import {
  GameDto,
  PlayerDto,
  OtherPlayerDto,
  PlayerActionDto,
  CardDto,
  TriggeredEffectDto,
} from "../../../types/generated/api-types.ts";

type TransitionPhase = "idle" | "lobby" | "fadeOutLobby" | "animateUI" | "complete";

interface GameLayoutProps {
  gameState: GameDto;
  currentPlayer: PlayerDto | null;
  playedCards?: CardDto[];
  corporationCard?: CardDto | null;
  isAnyModalOpen?: boolean;
  isLobbyPhase?: boolean;
  showCardSelection?: boolean;
  transitionPhase?: TransitionPhase;
  animateHexEntrance?: boolean;
  changedPaths?: Set<string>;
  tileHighlightMode?: TileHighlightMode;
  vpIndicators?: TileVPIndicator[];
  triggeredEffects?: TriggeredEffectDto[];
  onOpenCardEffectsModal?: () => void;
  onOpenCardsPlayedModal?: () => void;
  onOpenActionsModal?: () => void;
  onActionSelect?: (action: PlayerActionDto) => void;
  onConvertPlantsToGreenery?: () => void;
  onConvertHeatToTemperature?: () => void;
  showStandardProjectsPopover?: boolean;
  onToggleStandardProjectsPopover?: () => void;
  standardProjectsButtonRef?: React.RefObject<HTMLButtonElement | null>;
  showMilestonePopover?: boolean;
  onToggleMilestonePopover?: () => void;
  milestonesButtonRef?: React.RefObject<HTMLButtonElement | null>;
  showAwardPopover?: boolean;
  onToggleAwardPopover?: () => void;
  awardsButtonRef?: React.RefObject<HTMLButtonElement | null>;
  onLeaveGame?: () => void;
  onSkyboxReady?: () => void;
}

const GameLayout: React.FC<GameLayoutProps> = ({
  gameState,
  currentPlayer,
  playedCards = [],
  corporationCard = null,
  isAnyModalOpen: _isAnyModalOpen = false,
  isLobbyPhase: _isLobbyPhase = false,
  showCardSelection = false,
  transitionPhase = "idle",
  animateHexEntrance = false,
  changedPaths = new Set(),
  tileHighlightMode,
  vpIndicators = [],
  triggeredEffects = [],
  onOpenCardEffectsModal,
  onOpenCardsPlayedModal,
  onOpenActionsModal,
  onActionSelect,
  onConvertPlantsToGreenery,
  onConvertHeatToTemperature,
  showStandardProjectsPopover = false,
  onToggleStandardProjectsPopover,
  standardProjectsButtonRef,
  showMilestonePopover = false,
  onToggleMilestonePopover,
  milestonesButtonRef,
  showAwardPopover = false,
  onToggleAwardPopover,
  awardsButtonRef,
  onLeaveGame,
  onSkyboxReady,
}) => {
  // Create a map of all players (current + others) for easy lookup
  const playerMap = new Map<string, PlayerDto | OtherPlayerDto>();
  if (gameState?.currentPlayer) {
    playerMap.set(gameState.currentPlayer.id, gameState.currentPlayer);
  }
  gameState?.otherPlayers?.forEach((otherPlayer) => {
    playerMap.set(otherPlayer.id, otherPlayer);
  });

  // Construct allPlayers using the turn order from the backend
  const allPlayers: (PlayerDto | OtherPlayerDto)[] =
    (gameState?.turnOrder
      ?.map((playerId) => playerMap.get(playerId))
      .filter((player) => player !== undefined) as (PlayerDto | OtherPlayerDto)[]) || [];

  // Find the current turn player for the right sidebar
  const currentTurnPlayer =
    allPlayers.find((player) => player.id === gameState?.currentTurn) || null;

  const showUI =
    transitionPhase === "animateUI" || transitionPhase === "complete" || transitionPhase === "idle";
  const isAnimatingIn = transitionPhase === "animateUI";
  const uiAnimationClass = isAnimatingIn ? "animate-[uiFadeIn_1200ms_ease-out_both]" : "";
  const corpAnimationClass = isAnimatingIn ? "animate-[corpFadeIn_800ms_ease-out_800ms_both]" : "";

  return (
    <div className="relative w-screen h-screen bg-[#000000] text-white overflow-hidden">
      {/* CSS animations for transition */}
      <style>{`
        @keyframes uiFadeIn {
          from { opacity: 0; }
          to { opacity: 1; }
        }
        @keyframes corpFadeIn {
          from { opacity: 0; }
          to { opacity: 1; }
        }
      `}</style>

      {/* Game content takes full screen - hidden during lobby (SpaceBackground shown instead) */}
      {transitionPhase !== "lobby" && (
        <div className="absolute inset-0">
          <MainContentDisplay
            gameState={gameState}
            tileHighlightMode={tileHighlightMode}
            vpIndicators={vpIndicators}
            animateHexEntrance={animateHexEntrance}
            onSkyboxReady={onSkyboxReady}
          />
        </div>
      )}

      {/* TopMenuBar overlays on top */}
      {showUI && !showCardSelection && (
        <div className={uiAnimationClass}>
          <TopMenuBar
            gameState={gameState}
            showStandardProjectsPopover={showStandardProjectsPopover}
            onToggleStandardProjectsPopover={onToggleStandardProjectsPopover}
            standardProjectsButtonRef={standardProjectsButtonRef}
            showMilestonePopover={showMilestonePopover}
            onToggleMilestonePopover={onToggleMilestonePopover}
            milestonesButtonRef={milestonesButtonRef}
            showAwardPopover={showAwardPopover}
            onToggleAwardPopover={onToggleAwardPopover}
            awardsButtonRef={awardsButtonRef}
            onLeaveGame={onLeaveGame}
            gameId={gameState?.id}
          />
        </div>
      )}

      {/* Overlay Components */}
      {showUI && (
        <div className={uiAnimationClass}>
          <LeftSidebar
            players={allPlayers}
            currentPlayer={currentPlayer}
            turnPlayerId={gameState?.currentTurn || ""}
            currentPhase={gameState?.currentPhase}
            hasPendingTilePlacement={!!currentPlayer?.pendingTileSelection}
            triggeredEffects={triggeredEffects}
          />

          <RightSidebar
            globalParameters={gameState?.globalParameters}
            generation={gameState?.generation}
            currentPlayer={currentTurnPlayer}
          />

          <PlayerOverlay players={allPlayers} currentPlayer={currentPlayer} />
        </div>
      )}

      {showUI && !showCardSelection && (
        <>
          <div className={uiAnimationClass}>
            <BottomResourceBar
              currentPlayer={currentPlayer}
              gameState={gameState}
              playedCards={playedCards}
              changedPaths={changedPaths}
              onOpenCardEffectsModal={onOpenCardEffectsModal}
              onOpenCardsPlayedModal={onOpenCardsPlayedModal}
              onOpenActionsModal={onOpenActionsModal}
              onActionSelect={onActionSelect}
              onConvertPlantsToGreenery={onConvertPlantsToGreenery}
              onConvertHeatToTemperature={onConvertHeatToTemperature}
            />
          </div>

          {corporationCard && (
            <div className={corpAnimationClass}>
              <CorporationViewer corporation={corporationCard} />
            </div>
          )}
        </>
      )}
    </div>
  );
};

export default GameLayout;
