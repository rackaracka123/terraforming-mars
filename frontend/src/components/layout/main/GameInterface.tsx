import { useState, useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import GameLayout from "./GameLayout.tsx";
import CardsPlayedModal from "../../ui/modals/CardsPlayedModal.tsx";
import TagsModal from "../../ui/modals/TagsModal.tsx";
import VictoryPointsModal from "../../ui/modals/VictoryPointsModal.tsx";
import ActionsModal from "../../ui/modals/ActionsModal.tsx";
import CardEffectsModal from "../../ui/modals/CardEffectsModal.tsx";
import { webSocketService } from "../../../services/webSocketService.ts";
import { GameDto, PlayerDto } from "../../../types/generated/api-types.ts";

// Mock interface for GameLayout compatibility
interface MockGameState {
  id: string;
  players: MockPlayer[];
  currentPlayer: string;
  generation: number;
  phase: string;
  globalParameters: {
    temperature: number;
    oxygen: number;
    oceans: number;
  };
}

interface MockPlayer {
  id: string;
  name: string;
  resources: {
    credits: number;
    steel: number;
    titanium: number;
    plants: number;
    energy: number;
    heat: number;
  };
  production: {
    credits: number;
    steel: number;
    titanium: number;
    plants: number;
    energy: number;
    heat: number;
  };
  terraformRating: number;
  victoryPoints: number;
  corporation?: string;
  passed?: boolean;
  availableActions?: number;
}

export default function GameInterface() {
  const location = useLocation();
  const navigate = useNavigate();
  const [game, setGame] = useState<GameDto | null>(null);
  const [mockGameState, setMockGameState] = useState<MockGameState | null>(
    null,
  );
  const [isConnected, setIsConnected] = useState(false);
  const [currentPlayer, setCurrentPlayer] = useState<MockPlayer | null>(null);
  const [showCorporationModal, setShowCorporationModal] = useState(false);

  // New modal states
  const [showCardsPlayedModal, setShowCardsPlayedModal] = useState(false);
  const [showTagsModal, setShowTagsModal] = useState(false);
  const [showVictoryPointsModal, setShowVictoryPointsModal] = useState(false);
  const [showActionsModal, setShowActionsModal] = useState(false);
  const [showCardEffectsModal, setShowCardEffectsModal] = useState(false);

  // Helper function to convert Game to MockGameState for compatibility
  const convertGameToMockState = (
    gameData: GameDto,
    playerId: string,
  ): MockGameState => {
    return {
      id: gameData.id,
      players:
        gameData.players?.map((p: PlayerDto) => ({
          id: p.id,
          name: p.name,
          resources: p.resources || {
            credits: 0,
            steel: 0,
            titanium: 0,
            plants: 0,
            energy: 0,
            heat: 0,
          },
          production: p.production || {
            credits: 0,
            steel: 0,
            titanium: 0,
            plants: 0,
            energy: 0,
            heat: 0,
          },
          terraformRating: p.terraformRating || 20,
          victoryPoints: 0,
          corporation: p.corporation,
          passed: false,
          availableActions: 2,
        })) || [],
      currentPlayer: playerId,
      generation: gameData.generation || 1,
      phase: gameData.currentPhase || "setup",
      globalParameters: {
        temperature: gameData.globalParameters?.temperature || -30,
        oxygen: gameData.globalParameters?.oxygen || 0,
        oceans: gameData.globalParameters?.oceans || 0,
      },
    };
  };

  useEffect(() => {
    // Check if we have real game state from routing
    const routeState = location.state as {
      game?: GameDto;
      playerId?: string;
      playerName?: string;
    } | null;
    if (!routeState?.game || !routeState?.playerId) {
      // No game data, redirect back to landing page
      navigate("/");
      return;
    }

    setGame(routeState.game);
    setIsConnected(true);

    // Convert real game to mock format for GameLayout compatibility
    const mockState = convertGameToMockState(
      routeState.game,
      routeState.playerId,
    );
    setMockGameState(mockState);

    const player = mockState.players.find((p) => p.id === routeState.playerId);
    setCurrentPlayer(player || null);

    // Set up WebSocket listeners for real-time updates
    const handleGameUpdated = (updatedGame: GameDto) => {
      setGame(updatedGame);

      // Update mock state for compatibility
      const updatedMockState = convertGameToMockState(
        updatedGame,
        routeState.playerId,
      );
      setMockGameState(updatedMockState);

      const updatedPlayer = updatedMockState.players.find(
        (p) => p.id === routeState.playerId,
      );
      setCurrentPlayer(updatedPlayer || null);

      // Show corporation modal if player hasn't selected a corporation yet
      if (updatedPlayer && !updatedPlayer.corporation) {
        setShowCorporationModal(true);
      } else {
        setShowCorporationModal(false);
      }
    };

    const handleError = () => {
      // Could show error modal
    };

    const handleDisconnect = () => {
      setIsConnected(false);
      // Could redirect back to landing page
      navigate("/");
    };

    webSocketService.on("game-updated", handleGameUpdated);
    webSocketService.on("error", handleError);
    webSocketService.on("disconnect", handleDisconnect);

    return () => {
      webSocketService.off("game-updated", handleGameUpdated);
      webSocketService.off("error", handleError);
      webSocketService.off("disconnect", handleDisconnect);
    };
  }, [location.state, navigate]);

  // const handleCorporationSelection = (corporationId: string) => {
  //   webSocketService.playAction("select-corporation", { corporationId });
  //   setShowCorporationModal(false);
  // };

  // Demo data for the new modals (in a real app, this would come from game state)
  const demoCards = [
    {
      id: "card-1",
      name: "Mining Guild",
      type: "corporation" as const,
      cost: 0,
      description:
        "You start with 30 M€, 5 steel, and 1 steel production. Increase steel production 1 step for each steel and titanium resource on the board.",
      tags: ["building" as const, "space" as const],
      victoryPoints: 0,
      playOrder: 1,
    },
    {
      id: "card-2",
      name: "Power Plant",
      type: "automated" as const,
      cost: 4,
      description: "Increase your energy production 1 step.",
      tags: ["power" as const, "building" as const],
      victoryPoints: 0,
      playOrder: 2,
    },
    {
      id: "card-3",
      name: "Research",
      type: "active" as const,
      cost: 11,
      description: "Action: Spend 1 M€ to draw a card.",
      tags: ["science" as const],
      victoryPoints: 1,
      playOrder: 3,
    },
  ];

  const demoActions = [
    {
      id: "action-1",
      name: "Power Plant",
      type: "standard" as const,
      cost: 11,
      description: "Increase energy production 1 step.",
      available: true,
      immediate: true,
    },
    {
      id: "action-2",
      name: "Draw Cards",
      type: "card" as const,
      cost: 1,
      description: "Draw 1 card from the deck.",
      source: "Research",
      available: true,
      immediate: true,
    },
    {
      id: "action-3",
      name: "Diversifier",
      type: "milestone" as const,
      cost: 8,
      description: "Claim the Diversifier milestone.",
      requirement: "Have 8 different types of production",
      available: false,
    },
  ];

  const demoEffects = [
    {
      id: "effect-1",
      cardId: "card-1",
      cardName: "Mining Guild",
      cardType: "corporation" as const,
      effectType: "ongoing" as const,
      name: "Steel Production Bonus",
      description: "Get +1 steel production for each steel/titanium on board",
      isActive: true,
      category: "production" as const,
      resource: "steel",
      value: 1,
    },
    {
      id: "effect-2",
      cardId: "card-3",
      cardName: "Research",
      cardType: "active" as const,
      effectType: "triggered" as const,
      name: "Research Bonus",
      description: "Get bonus cards when researching",
      isActive: true,
      category: "bonus" as const,
    },
  ];

  const handleActionSelect = () => {
    // In a real app, emit to server
  };

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
        }
      }
    };

    window.addEventListener("keydown", handleKeyPress);
    return () => window.removeEventListener("keydown", handleKeyPress);
  }, []);

  if (!isConnected || !mockGameState || !game) {
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
        <p>Connecting to game...</p>
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
    showCardEffectsModal;

  return (
    <>
      <GameLayout
        gameState={mockGameState}
        currentPlayer={currentPlayer}
        socket={webSocketService}
        isAnyModalOpen={isAnyModalOpen}
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
        cards={demoCards}
        playerName={currentPlayer?.name}
      />

      <TagsModal
        isVisible={showTagsModal}
        onClose={() => setShowTagsModal(false)}
        cards={demoCards}
        playerName={currentPlayer?.name}
      />

      <VictoryPointsModal
        isVisible={showVictoryPointsModal}
        onClose={() => setShowVictoryPointsModal(false)}
        cards={demoCards}
        terraformRating={currentPlayer?.terraformRating}
        playerName={currentPlayer?.name}
      />

      <ActionsModal
        isVisible={showActionsModal}
        onClose={() => setShowActionsModal(false)}
        actions={demoActions}
        playerName={currentPlayer?.name}
        onActionSelect={handleActionSelect}
      />

      <CardEffectsModal
        isVisible={showCardEffectsModal}
        onClose={() => setShowCardEffectsModal(false)}
        effects={demoEffects}
        cards={demoCards}
        playerName={currentPlayer?.name}
      />
    </>
  );
}
