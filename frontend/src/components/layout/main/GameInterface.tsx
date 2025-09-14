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
import CardFanOverlay from "../../ui/overlay/CardFanOverlay.tsx";
import { globalWebSocketManager } from "../../../services/globalWebSocketManager.ts";
import { getTabManager } from "../../../utils/tabManager.ts";
import audioService from "../../../services/audioService.ts";
import {
  CardDto,
  FullStatePayload,
  GameDto,
  GamePhaseStartingCardSelection,
  GameStatusActive,
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
  const [playerId, setPlayerId] = useState<string | null>(null); // Track player ID separately
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
  const [cardDetails, setCardDetails] = useState<CardDto[]>([]);

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
  const handleGameUpdated = useCallback(
    (updatedGame: GameDto) => {
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
      setIsConnected(true);

      // If we were reconnecting, mark reconnection as successful
      if (isReconnecting) {
        console.log("✅ Reconnection successful");
        setIsReconnecting(false);
      }

      // Set current player from updated game data
      const updatedPlayer = updatedGame.currentPlayer;
      setCurrentPlayer(updatedPlayer || null);

      // Show corporation modal if player hasn't selected a corporation yet
      if (updatedPlayer && !updatedPlayer.corporation) {
        setShowCorporationModal(true);
      } else {
        setShowCorporationModal(false);
      }
    },
    [isReconnecting],
  );

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

    // Only start reconnection if we were actually connected to a game
    if (currentPlayerIdRef.current) {
      // Start in-place reconnection instead of redirecting
      setIsReconnecting(true);

      const savedGameData = localStorage.getItem("terraforming-mars-game");
      if (savedGameData) {
        console.log("🔄 Connection lost, attempting reconnection...");
        // Attempt to reconnect in place
        attemptReconnection();
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
      // Play production phase sound
      void audioService.playProductionSound();

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

  const handleCardSelection = useCallback(async (selectedCardIds: string[]) => {
    try {
      // Send card selection to server
      await globalWebSocketManager.selectStartingCard(selectedCardIds);
      // Modal will auto-hide when backend clears startingSelection field
    } catch (error) {
      console.error("Failed to select cards:", error);
    }
  }, []);

  const handlePlayCard = useCallback(async (cardId: string) => {
    try {
      console.log(`🎯 Playing card: ${cardId}`);
      await globalWebSocketManager.playCard(cardId);
      console.log(`✅ Card ${cardId} played successfully`);
    } catch (error) {
      console.error(`❌ Failed to play card ${cardId}:`, error);
      throw error; // Re-throw to allow CardFanOverlay to handle the error
    }
  }, []);

  // Attempt reconnection to the game
  const attemptReconnection = useCallback(async () => {
    try {
      const savedGameData = localStorage.getItem("terraforming-mars-game");
      if (!savedGameData) {
        console.error("No saved game data for reconnection");
        navigate("/", { replace: true });
        return;
      }

      const { gameId, playerId, playerName } = JSON.parse(savedGameData);
      console.log("🔄 Reconnecting to game:", { gameId, playerId, playerName });

      // Fetch current game state from server first
      const response = await fetch(
        `http://localhost:3001/api/v1/games/${gameId}`,
      );
      if (!response.ok) {
        throw new Error(`Game not found: ${response.status}`);
      }

      const gameData = await response.json();
      console.log("✅ Game state fetched successfully");

      // Update local state with fetched game data
      setGame(gameData.game);
      setPlayerId(playerId);

      // Set current player from fetched game data
      const player = gameData.game.currentPlayer;
      setCurrentPlayer(player || null);

      // Store player ID for WebSocket handlers
      currentPlayerIdRef.current = playerId;

      // Now establish WebSocket connection
      console.log("🔌 Establishing WebSocket connection...");
      globalWebSocketManager.connect(gameId, playerId, playerName);

      // Connection will be marked as successful when WebSocket connects
      // and we receive a game-updated or player-reconnected event
      console.log("⏳ Waiting for WebSocket connection confirmation...");
    } catch (error) {
      console.error("❌ Reconnection failed:", error);
      setIsReconnecting(false);
      // Don't navigate away - let user try manual reconnection
      // or they can manually navigate to home if needed
      console.error(
        "Failed to reconnect to game. Please check your connection and try again.",
      );
    }
  }, [navigate]);

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

        // Store player ID for WebSocket handlers and component state
        currentPlayerIdRef.current = routeState.playerId;
        setPlayerId(routeState.playerId);
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
        // No route state, check if we should attempt reconnection
        const savedGameData = localStorage.getItem("terraforming-mars-game");
        if (savedGameData) {
          console.log(
            "🔄 No route state, attempting reconnection from saved data...",
          );
          // Start in-place reconnection instead of redirecting
          setIsReconnecting(true);
          attemptReconnection();
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

      // Store player ID for WebSocket handlers and component state
      currentPlayerIdRef.current = routeState.playerId;
      setPlayerId(routeState.playerId);

      // CRITICAL FIX: Ensure globalWebSocketManager knows the current player ID
      // This is essential for reconnection scenarios where the player ID must be preserved
      globalWebSocketManager.setCurrentPlayerId(routeState.playerId);

      // CRITICAL FIX: Send the player-connect WebSocket message to complete reconnection
      // This is necessary after page refresh when we have the game state from route
      // but need to re-establish the WebSocket connection with the backend
      void globalWebSocketManager.playerConnect(
        routeState.playerName,
        routeState.game.id,
        routeState.playerId,
      );
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

  // Extract card details directly from game data (backend now sends full card objects)
  const extractCardDetails = useCallback((cards: CardDto[]) => {
    console.log("📥 Extracting card details from backend:", cards);
    setCardDetails(cards);
  }, []);

  // Show/hide starting card selection overlay based on backend state
  useEffect(() => {
    const cards = game?.currentPlayer?.startingSelection;
    const hasCardSelection = cards && cards.length > 0;

    console.log("🔍 Checking overlay conditions:", {
      gamePhase: game?.currentPhase,
      gameStatus: game?.status,
      hasCardSelection,
      currentlyShowing: showCardSelection,
      cardCount: cards?.length,
    });

    if (
      game?.currentPhase === GamePhaseStartingCardSelection &&
      game?.status === GameStatusActive &&
      hasCardSelection &&
      !showCardSelection
    ) {
      console.log("🎪 Showing overlay - backend has starting cards");
      extractCardDetails(cards);
      setShowCardSelection(true);
    } else if (showCardSelection && !hasCardSelection) {
      console.log(
        "🎪 Hiding overlay - backend cleared starting cards (selection complete)",
      );
      setShowCardSelection(false);
    }
  }, [
    game?.currentPhase,
    game?.status,
    game?.currentPlayer?.startingSelection,
    showCardSelection,
    extractCardDetails,
  ]);

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

      {isLobbyPhase && game && playerId && (
        <WaitingRoomOverlay game={game} playerId={playerId} />
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
        cards={cardDetails}
        playerCredits={currentPlayer?.resources?.credits || 40}
        onConfirmSelection={handleCardSelection}
      />

      {/* Card fan overlay for hand cards */}
      <CardFanOverlay
        cards={currentPlayer?.cards || []}
        hideWhenModalOpen={showCardSelection || isLobbyPhase}
        onCardSelect={(cardId) => {
          // TODO: Implement card selection logic (view details, etc.)
          console.log("Card selected:", cardId);
        }}
        onPlayCard={handlePlayCard}
      />

      {/* Reconnection overlay */}
      {isReconnecting && (
        <div
          style={{
            position: "fixed",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: "rgba(0, 0, 0, 0.7)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 9999,
            color: "white",
            fontSize: "18px",
            textAlign: "center",
            flexDirection: "column",
            gap: "20px",
          }}
        >
          <div style={{ fontSize: "24px", fontWeight: "bold" }}>
            🔄 Reconnecting to Game...
          </div>
          <div>Please wait while we restore your connection</div>
          <div
            style={{
              width: "40px",
              height: "40px",
              border: "4px solid rgba(255, 255, 255, 0.3)",
              borderTop: "4px solid white",
              borderRadius: "50%",
              animation: "spin 1s linear infinite",
            }}
          />
          <style>{`
            @keyframes spin {
              0% { transform: rotate(0deg); }
              100% { transform: rotate(360deg); }
            }
          `}</style>
        </div>
      )}
    </>
  );
}
