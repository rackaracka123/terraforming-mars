import React, { useState, useEffect } from "react";
import GameLayout from "./GameLayout.tsx";
import CorporationSelectionModal from "../../ui/modals/CorporationSelectionModal.tsx";
import CardsPlayedModal from "../../ui/modals/CardsPlayedModal.tsx";
import TagsModal from "../../ui/modals/TagsModal.tsx";
import VictoryPointsModal from "../../ui/modals/VictoryPointsModal.tsx";
import ActionsModal from "../../ui/modals/ActionsModal.tsx";
import CardEffectsModal from "../../ui/modals/CardEffectsModal.tsx";
import {
  mockSocketService,
  GameState,
  Player,
} from "../../../services/mockGameService.ts";

// Types imported from mockGameService

export default function GameInterface() {
  const [gameState, setGameState] = useState<GameState | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [currentPlayer, setCurrentPlayer] = useState<Player | null>(null);
  const [showCorporationModal, setShowCorporationModal] = useState(false);
  const [availableCorporations, setAvailableCorporations] = useState<any[]>([]);

  // New modal states
  const [showCardsPlayedModal, setShowCardsPlayedModal] = useState(false);
  const [showTagsModal, setShowTagsModal] = useState(false);
  const [showVictoryPointsModal, setShowVictoryPointsModal] = useState(false);
  const [showActionsModal, setShowActionsModal] = useState(false);
  const [showCardEffectsModal, setShowCardEffectsModal] = useState(false);

  useEffect(() => {
    // Set up mock service listeners
    mockSocketService.on("connect", () => {
      setIsConnected(true);
      console.log("Connected to mock service");
      // Auto-join with a default name
      mockSocketService.emit("join-game", {
        gameId: "demo",
        playerName: "Player",
      });
    });

    mockSocketService.on("game-updated", (updatedGameState: GameState) => {
      setGameState(updatedGameState);
      const player = updatedGameState.players.find(
        (p) => p.id === mockSocketService.id,
      );
      setCurrentPlayer(player || null);

      // Show corporation modal if player hasn't selected a corporation yet
      if (player && !player.corporation) {
        setShowCorporationModal(true);
      } else {
        setShowCorporationModal(false);
      }
    });

    mockSocketService.on("corporations-available", (corporations: any[]) => {
      setAvailableCorporations(corporations);
    });

    mockSocketService.on("disconnect", () => {
      setIsConnected(false);
      console.log("Disconnected from mock service");
    });

    return () => {
      mockSocketService.disconnect();
    };
  }, []);

  const handleCorporationSelection = (corporationId: string) => {
    mockSocketService.emit("select-corporation", { corporationId });
    setShowCorporationModal(false);
  };

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

  const handleActionSelect = (action: any) => {
    console.log("Action selected:", action);
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

  if (!isConnected || !gameState) {
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
        <p>Setting up mock game data...</p>
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
        gameState={gameState}
        currentPlayer={currentPlayer}
        socket={mockSocketService}
        isAnyModalOpen={isAnyModalOpen}
        onOpenCardEffectsModal={() => setShowCardEffectsModal(true)}
        onOpenActionsModal={() => setShowActionsModal(true)}
        onOpenCardsPlayedModal={() => setShowCardsPlayedModal(true)}
        onOpenTagsModal={() => setShowTagsModal(true)}
        onOpenVictoryPointsModal={() => setShowVictoryPointsModal(true)}
      />

      {/* Original Corporation Selection Modal */}
      <CorporationSelectionModal
        corporations={availableCorporations}
        onSelectCorporation={handleCorporationSelection}
        isVisible={showCorporationModal}
      />

      {/* New Enhanced Modals */}
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
