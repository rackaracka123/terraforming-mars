import { useCallback, useEffect, useRef, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import GameLayout from "./GameLayout.tsx";
import CardsPlayedModal from "../../ui/modals/CardsPlayedModal.tsx";
import TagsModal from "../../ui/modals/TagsModal.tsx";
import VictoryPointsModal from "../../ui/modals/VictoryPointsModal.tsx";
import ActionsModal from "../../ui/modals/ActionsModal.tsx";
import CardEffectsModal from "../../ui/modals/CardEffectsModal.tsx";
import ProductionPhaseModal from "../../ui/modals/ProductionPhaseModal.tsx";
import DebugDropdown from "../../ui/debug/DebugDropdown.tsx";
import WaitingRoomOverlay from "../../ui/overlay/WaitingRoomOverlay.tsx";
import TabConflictOverlay from "../../ui/overlay/TabConflictOverlay.tsx";
import StartingCardSelectionOverlay from "../../ui/overlay/StartingCardSelectionOverlay.tsx";
import { globalWebSocketManager } from "../../../services/globalWebSocketManager.ts";
import { getTabManager } from "../../../utils/tabManager.ts";
import {
  FullStatePayload,
  GameDto,
  GameStatusLobby,
  PlayerDisconnectedPayload,
  PlayerDto,
  PlayerReconnectedPayload,
  ProductionPhaseStartedPayload,
} from "../../../types/generated/api-types.ts";
import { deepClone, findChangedPaths } from "../../../utils/deepCompare.ts";

export default function GameInterface() {
  const location = useLocation();
  const navigate = useNavigate();
  const [game, setGame] = useState<GameDto | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [isReconnecting, setIsReconnecting] = useState(false);
  const [currentPlayer, setCurrentPlayer] = useState<PlayerDto | null>(null);
  const [showCorporationModal, setShowCorporationModal] = useState(false);

  // New modal states
  const [showCardsPlayedModal, setShowCardsPlayedModal] = useState(false);
  const [showTagsModal, setShowTagsModal] = useState(false);
  const [showVictoryPointsModal, setShowVictoryPointsModal] = useState(false);
  const [showActionsModal, setShowActionsModal] = useState(false);
  const [showCardEffectsModal, setShowCardEffectsModal] = useState(false);
  const [showDebugDropdown, setShowDebugDropdown] = useState(false);

  // Production phase modal state
  const [showProductionPhaseModal, setShowProductionPhaseModal] =
    useState(false);
  const [productionPhaseData, setProductionPhaseData] =
    useState<ProductionPhaseStartedPayload | null>(null);

  // Card selection state
  const [showCardSelection, setShowCardSelection] = useState(false);
  const [availableCards, setAvailableCards] = useState<any[]>([]);

  // Tab management
  const [showTabConflict, setShowTabConflict] = useState(false);
  const [conflictingTabInfo, setConflictingTabInfo] = useState<{
    gameId: string;
    playerName: string;
  } | null>(null);

  // Change detection
  const previousGameRef = useRef<GameDto | null>(null);
  const [changedPaths, setChangedPaths] = useState<Set<string>>(new Set());

  // WebSocket stability
  const isWebSocketInitialized = useRef(false);
  const currentPlayerIdRef = useRef<string | null>(null);

  // Stable WebSocket event handlers using useCallback
  const handleGameUpdated = useCallback((updatedGame: GameDto) => {
    const playerId = currentPlayerIdRef.current;
    if (!playerId) return;

    // Detect changes before updating
    if (previousGameRef.current) {
      const changes = findChangedPaths(previousGameRef.current, updatedGame);
      setChangedPaths(changes);

      // Clear changed paths after animation completes
      if (changes.size > 0) {
        setTimeout(() => {
          setChangedPaths(new Set());
        }, 1500);
      }
    }

    // Store the previous state for next comparison
    previousGameRef.current = deepClone(updatedGame);

    setGame(updatedGame);

    // Set current player from updated game data
    const updatedPlayer = updatedGame.currentPlayer;
    setCurrentPlayer(updatedPlayer || null);

    // Show corporation modal if player hasn't selected a corporation yet
    if (updatedPlayer && !updatedPlayer.corporation) {
      setShowCorporationModal(true);
    } else {
      setShowCorporationModal(false);
    }
  }, []);

  const handleFullState = useCallback(
    (statePayload: FullStatePayload) => {
      // Handle full-state message (e.g., on reconnection)
      if (statePayload.game) {
        handleGameUpdated(statePayload.game);
      }
    },
    [handleGameUpdated],
  );

  const handlePlayerConnected = useCallback(
    (payload: any) => {
      // Handle player-connected message that now includes game state
      if (payload.game) {
        handleGameUpdated(payload.game);
      }
    },
    [handleGameUpdated],
  );

  const handleError = useCallback(() => {
    // Could show error modal
  }, []);

  const handleDisconnect = useCallback(() => {
    // WebSocket connection closed - this client lost connection
    setIsConnected(false);

    // Only redirect if we were actually connected to a game
    if (currentPlayerIdRef.current) {
      // Attempting reconnection instead of returning to main menu
      setIsReconnecting(true);

      // Try to navigate to reconnecting page with game data
      const savedGameData = localStorage.getItem("terraforming-mars-game");
      if (savedGameData) {
        navigate("/reconnecting", { replace: true });
      } else {
        // No saved game data, go to main menu
        navigate("/", { replace: true });
      }
    }
  }, [navigate]);

  const handlePlayerReconnected = useCallback(
    (payload: PlayerReconnectedPayload) => {
      // Handle when any player (including this player) reconnects
      if (payload.game) {
        handleGameUpdated(payload.game);
      }
      // You could show a notification here about the reconnected player
      // Player reconnected: payload.playerName
    },
    [handleGameUpdated],
  );

  const handlePlayerDisconnected = useCallback(
    (payload: PlayerDisconnectedPayload) => {
      // Handle when any player disconnects (NOT this client)
      // Player disconnected from the game

      if (payload.game) {
        handleGameUpdated(payload.game);
      }
    },
    [handleGameUpdated],
  );

  const handleProductionPhaseStarted = useCallback(
    (payload: ProductionPhaseStartedPayload) => {
      // Show production phase modal with animation data
      setProductionPhaseData(payload);
      setShowProductionPhaseModal(true);

      // Update game state if provided
      if (payload.game) {
        handleGameUpdated(payload.game);
      }
    },
    [handleGameUpdated],
  );

  const handleAvailableCards = useCallback(
    (payload: any) => {
      // Available cards received - show selection overlay if in correct phase
      if (payload?.cards && Array.isArray(payload.cards)) {
        setAvailableCards(payload.cards);
        // Only show overlay if we're in the starting card selection phase
        const currentGamePhase = game?.currentPhase;
        if (currentGamePhase === "starting_card_selection") {
          setShowCardSelection(true);
        }
      }
    },
    [game],
  );

  const handleCardSelection = useCallback(
    async (selectedCardIds: string[]) => {
      try {
        // Send card selection to server
        await globalWebSocketManager.playAction({
          type: "select-starting-cards",
          cardIds: selectedCardIds,
        });
        // Close the overlay
        setShowCardSelection(false);
        setAvailableCards([]);
      } catch (error) {
        console.error("Failed to select cards:", error);
      }
    },
    [],
  );

  // Setup WebSocket listeners using global manager - only initialize once
  const setupWebSocketListeners = useCallback(() => {
    if (isWebSocketInitialized.current) {
      return () => {}; // Already initialized, return empty cleanup
    }

    globalWebSocketManager.on("game-updated", handleGameUpdated);
    globalWebSocketManager.on("full-state", handleFullState);
    globalWebSocketManager.on("player-connected", handlePlayerConnected);
    globalWebSocketManager.on("player-reconnected", handlePlayerReconnected);
    globalWebSocketManager.on("player-disconnected", handlePlayerDisconnected);
    globalWebSocketManager.on(
      "production-phase-started",
      handleProductionPhaseStarted,
    );
    globalWebSocketManager.on("available-cards", handleAvailableCards);
    globalWebSocketManager.on("error", handleError);
    globalWebSocketManager.on("disconnect", handleDisconnect);

    isWebSocketInitialized.current = true;

    return () => {
      globalWebSocketManager.off("game-updated", handleGameUpdated);
      globalWebSocketManager.off("full-state", handleFullState);
      globalWebSocketManager.off("player-connected", handlePlayerConnected);
      globalWebSocketManager.off("player-reconnected", handlePlayerReconnected);
      globalWebSocketManager.off(
        "player-disconnected",
        handlePlayerDisconnected,
      );
      globalWebSocketManager.off(
        "production-phase-started",
        handleProductionPhaseStarted,
      );
      globalWebSocketManager.off("available-cards", handleAvailableCards);
      globalWebSocketManager.off("error", handleError);
      globalWebSocketManager.off("disconnect", handleDisconnect);
      isWebSocketInitialized.current = false;
    };
  }, [
    handleGameUpdated,
    handleFullState,
    handlePlayerConnected,
    handlePlayerReconnected,
    handlePlayerDisconnected,
    handleProductionPhaseStarted,
    handleAvailableCards,
    handleError,
    handleDisconnect,
  ]);

  // Tab conflict handlers
  const handleTabTakeOver = () => {
    if (conflictingTabInfo) {
      const tabManager = getTabManager();
      tabManager.forceTakeOver(
        conflictingTabInfo.gameId,
        conflictingTabInfo.playerName,
      );
      setShowTabConflict(false);
      setConflictingTabInfo(null);

      // Now initialize the game with the route state
      const routeState = location.state as {
        game?: GameDto;
        playerId?: string;
        playerName?: string;
      } | null;

      if (routeState?.game && routeState?.playerId && routeState?.playerName) {
        setGame(routeState.game);
        setIsConnected(true);

        // Store game data for reconnection
        localStorage.setItem(
          "terraforming-mars-game",
          JSON.stringify({
            gameId: routeState.game.id,
            playerId: routeState.playerId,
            playerName: routeState.playerName,
            timestamp: Date.now(),
          }),
        );

        // Set current player from game data
        const player = routeState.game.currentPlayer;
        setCurrentPlayer(player || null);

        // Store player ID for WebSocket handlers
        currentPlayerIdRef.current = routeState.playerId;
      }
    }
  };

  const handleTabCancel = () => {
    setShowTabConflict(false);
    setConflictingTabInfo(null);
    // Return to main menu
    navigate("/", { replace: true });
  };

  useEffect(() => {
    const initializeGame = async () => {
      // Check if we have real game state from routing
      const routeState = location.state as {
        game?: GameDto;
        playerId?: string;
        playerName?: string;
        isReconnection?: boolean;
      } | null;

      if (
        !routeState?.game ||
        !routeState?.playerId ||
        !routeState?.playerName
      ) {
        // No route state, check if we should route to reconnection page
        const savedGameData = localStorage.getItem("terraforming-mars-game");
        if (savedGameData) {
          // Route to reconnecting page instead of attempting reconnection here
          navigate("/reconnecting", { replace: true });
          return;
        }

        // No saved data, return to main menu
        navigate("/", { replace: true });
        return;
      }

      // We have route state, try to claim the tab for this game session
      const tabManager = getTabManager();
      const canClaim = await tabManager.claimTab(
        routeState.game.id,
        routeState.playerName,
      );

      if (!canClaim) {
        // Another tab has this game open, show conflict overlay
        const activeTabInfo = tabManager.getActiveTabInfo();
        if (activeTabInfo) {
          setConflictingTabInfo(activeTabInfo);
          setShowTabConflict(true);
          return;
        }
      }

      // Successfully claimed tab or no conflict, initialize game
      setGame(routeState.game);
      setIsConnected(true);

      // Store game data for reconnection
      localStorage.setItem(
        "terraforming-mars-game",
        JSON.stringify({
          gameId: routeState.game.id,
          playerId: routeState.playerId,
          playerName: routeState.playerName,
          timestamp: Date.now(),
        }),
      );

      // Set current player from game data
      const player = routeState.game.currentPlayer;
      setCurrentPlayer(player || null);

      // Store player ID for WebSocket handlers
      currentPlayerIdRef.current = routeState.playerId;

      // CRITICAL FIX: Ensure globalWebSocketManager knows the current player ID
      // This is essential for reconnection scenarios where the player ID must be preserved
      globalWebSocketManager.setCurrentPlayerId(routeState.playerId);
    };

    void initializeGame();
  }, [location.state, navigate]);

  // Register event listeners when component mounts, unregister on unmount
  useEffect(() => {
    // Store player ID in global manager for event handling
    if (currentPlayerIdRef.current) {
      globalWebSocketManager.setCurrentPlayerId(currentPlayerIdRef.current);
    }

    return setupWebSocketListeners();
  }, [setupWebSocketListeners]);

  const handleActionSelect = () => {
    // In a real app, emit to server
  };

  // Listen for debug dropdown toggle from TopMenuBar
  useEffect(() => {
    const handleToggleDebug = () => {
      setShowDebugDropdown((prev) => !prev);
    };

    window.addEventListener("toggle-debug-dropdown", handleToggleDebug);
    return () => {
      window.removeEventListener("toggle-debug-dropdown", handleToggleDebug);
    };
  }, []);

  // Demo keyboard shortcuts
  useEffect(() => {
    const handleKeyPress = (event: KeyboardEvent) => {
      if (event.ctrlKey || event.metaKey) {
        switch (event.key) {
          case "1":
            event.preventDefault();
            setShowCardsPlayedModal(true);
            break;
          case "2":
            event.preventDefault();
            setShowTagsModal(true);
            break;
          case "3":
            event.preventDefault();
            setShowVictoryPointsModal(true);
            break;
          case "4":
            event.preventDefault();
            setShowActionsModal(true);
            break;
          case "5":
            event.preventDefault();
            setShowCardEffectsModal(true);
            break;
          case "d":
          case "D":
            event.preventDefault();
            setShowDebugDropdown(!showDebugDropdown);
            break;
        }
      }
    };

    window.addEventListener("keydown", handleKeyPress);
    return () => window.removeEventListener("keydown", handleKeyPress);
  }, [showDebugDropdown]);

  if (!isConnected || !game || isReconnecting) {
    return (
      <div
        style={{
          padding: "20px",
          color: "white",
          background: "#000011",
          minHeight: "100vh",
        }}
      >
        <h2>Loading Terraforming Mars...</h2>
        <p>
          {isReconnecting ? "Reconnecting to game..." : "Connecting to game..."}
        </p>
      </div>
    );
  }

  // Check if any modal is currently open
  const isAnyModalOpen =
    showCorporationModal ||
    showCardsPlayedModal ||
    showTagsModal ||
    showVictoryPointsModal ||
    showActionsModal ||
    showCardEffectsModal ||
    showProductionPhaseModal;

  // Check if game is in lobby phase
  const isLobbyPhase = game?.status === GameStatusLobby;

  return (
    <>
      <GameLayout
        gameState={game}
        currentPlayer={currentPlayer}
        isAnyModalOpen={isAnyModalOpen}
        isLobbyPhase={isLobbyPhase}
        onOpenCardEffectsModal={() => setShowCardEffectsModal(true)}
        onOpenActionsModal={() => setShowActionsModal(true)}
        onOpenCardsPlayedModal={() => setShowCardsPlayedModal(true)}
        onOpenTagsModal={() => setShowTagsModal(true)}
        onOpenVictoryPointsModal={() => setShowVictoryPointsModal(true)}
      />

      {/*<CorporationSelectionModal*/}
      {/*  corporations={availableCorporations}*/}
      {/*  onSelectCorporation={handleCorporationSelection}*/}
      {/*  isVisible={showCorporationModal}*/}
      {/*/>*/}

      <CardsPlayedModal
        isVisible={showCardsPlayedModal}
        onClose={() => setShowCardsPlayedModal(false)}
        cards={[]}
        playerName={currentPlayer?.name}
      />

      <TagsModal
        isVisible={showTagsModal}
        onClose={() => setShowTagsModal(false)}
        cards={[]}
        playerName={currentPlayer?.name}
      />

      <VictoryPointsModal
        isVisible={showVictoryPointsModal}
        onClose={() => setShowVictoryPointsModal(false)}
        cards={[]}
        terraformRating={currentPlayer?.terraformRating}
        playerName={currentPlayer?.name}
      />

      <ActionsModal
        isVisible={showActionsModal}
        onClose={() => setShowActionsModal(false)}
        actions={[]}
        playerName={currentPlayer?.name}
        onActionSelect={handleActionSelect}
      />

      <CardEffectsModal
        isVisible={showCardEffectsModal}
        onClose={() => setShowCardEffectsModal(false)}
        effects={[]}
        cards={[]}
        playerName={currentPlayer?.name}
      />

      <ProductionPhaseModal
        isOpen={showProductionPhaseModal}
        productionData={productionPhaseData}
        onClose={() => {
          setShowProductionPhaseModal(false);
          setProductionPhaseData(null);
        }}
      />

      <DebugDropdown
        isVisible={showDebugDropdown}
        onClose={() => setShowDebugDropdown(false)}
        gameState={game}
        changedPaths={changedPaths}
      />

      {isLobbyPhase && game && (
        <WaitingRoomOverlay game={game} playerId={currentPlayer?.id || ""} />
      )}

      {showTabConflict && conflictingTabInfo && (
        <TabConflictOverlay
          activeGameInfo={conflictingTabInfo}
          onTakeOver={handleTabTakeOver}
          onCancel={handleTabCancel}
        />
      )}

      {/* Starting card selection overlay */}
      <StartingCardSelectionOverlay
        isOpen={showCardSelection}
        cards={availableCards}
        playerCredits={currentPlayer?.resources?.credits || 40}
        onConfirmSelection={handleCardSelection}
      />
    </>
  );
}
