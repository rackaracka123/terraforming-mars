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
} from "../../../types/generated/api-types.ts";

interface GameLayoutProps {
  gameState: GameDto;
  currentPlayer: PlayerDto | null;
  playedCards?: CardDto[];
  corporationCard?: CardDto | null;
  isAnyModalOpen?: boolean;
  isLobbyPhase?: boolean;
  showCardSelection?: boolean;
  changedPaths?: Set<string>;
  tileHighlightMode?: TileHighlightMode;
  vpIndicators?: TileVPIndicator[];
  onOpenCardEffectsModal?: () => void;
  onOpenCardsPlayedModal?: () => void;
  onOpenVictoryPointsModal?: () => void;
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
}

const GameLayout: React.FC<GameLayoutProps> = ({
  gameState,
  currentPlayer,
  playedCards = [],
  corporationCard = null,
  isAnyModalOpen: _isAnyModalOpen = false,
  isLobbyPhase = false,
  showCardSelection = false,
  changedPaths = new Set(),
  tileHighlightMode,
  vpIndicators = [],
  onOpenCardEffectsModal,
  onOpenCardsPlayedModal,
  onOpenVictoryPointsModal,
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

  return (
    <div className="relative w-screen h-screen bg-[#000011] bg-[url('/assets/background-noise.png')] [background-attachment:fixed] bg-repeat text-white overflow-hidden">
      {/* Game content takes full screen */}
      <div className="absolute inset-0">
        <MainContentDisplay
          gameState={gameState}
          tileHighlightMode={tileHighlightMode}
          vpIndicators={vpIndicators}
        />
      </div>

      {/* TopMenuBar overlays on top */}
      {!isLobbyPhase && !showCardSelection && (
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
      )}

      {/* Overlay Components */}
      <LeftSidebar
        players={allPlayers}
        currentPlayer={currentPlayer}
        turnPlayerId={gameState?.currentTurn || ""}
        currentPhase={gameState?.currentPhase}
      />

      <RightSidebar
        globalParameters={gameState?.globalParameters}
        generation={gameState?.generation}
        currentPlayer={currentTurnPlayer}
      />

      <PlayerOverlay players={allPlayers} currentPlayer={currentPlayer} />

      {!isLobbyPhase && !showCardSelection && (
        <>
          <BottomResourceBar
            currentPlayer={currentPlayer}
            gameState={gameState}
            playedCards={playedCards}
            changedPaths={changedPaths}
            onOpenCardEffectsModal={onOpenCardEffectsModal}
            onOpenCardsPlayedModal={onOpenCardsPlayedModal}
            onOpenVictoryPointsModal={onOpenVictoryPointsModal}
            onOpenActionsModal={onOpenActionsModal}
            onActionSelect={onActionSelect}
            onConvertPlantsToGreenery={onConvertPlantsToGreenery}
            onConvertHeatToTemperature={onConvertHeatToTemperature}
          />

          {corporationCard && <CorporationViewer corporation={corporationCard} />}
        </>
      )}
    </div>
  );
};

export default GameLayout;
