import { useCallback, useEffect, useRef, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import GameLayout from "./GameLayout.tsx";
import CardsPlayedModal from "../../ui/modals/CardsPlayedModal.tsx";
import VictoryPointsModal from "../../ui/modals/VictoryPointsModal.tsx";
import EffectsModal from "../../ui/modals/EffectsModal.tsx";
import ActionsModal from "../../ui/modals/ActionsModal.tsx";
import StandardProjectPopover from "../../ui/popover/StandardProjectPopover.tsx";
import ProductionPhaseModal from "../../ui/modals/ProductionPhaseModal.tsx";
import PaymentSelectionPopover from "../../ui/popover/PaymentSelectionPopover.tsx";
import DebugDropdown from "../../ui/debug/DebugDropdown.tsx";
import DevModeChip from "../../ui/debug/DevModeChip.tsx";
import WaitingRoomOverlay from "../../ui/overlay/WaitingRoomOverlay.tsx";
import TabConflictOverlay from "../../ui/overlay/TabConflictOverlay.tsx";
import StartingCardSelectionOverlay from "../../ui/overlay/StartingCardSelectionOverlay.tsx";
import PendingCardSelectionOverlay from "../../ui/overlay/PendingCardSelectionOverlay.tsx";
import CardDrawSelectionOverlay from "../../ui/overlay/CardDrawSelectionOverlay.tsx";
import CardFanOverlay from "../../ui/overlay/CardFanOverlay.tsx";
import LoadingSpinner from "../../game/view/LoadingSpinner.tsx";
import HexagonalShieldOverlay from "../../ui/overlay/HexagonalShieldOverlay.tsx";
import ChoiceSelectionPopover from "../../ui/popover/ChoiceSelectionPopover.tsx";
import CardStorageSelectionPopover from "../../ui/popover/CardStorageSelectionPopover.tsx";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import { getTabManager } from "@/utils/tabManager.ts";
import audioService from "../../../services/audioService.ts";
import { skyboxCache } from "@/services/SkyboxCache.ts";
import {
  CardDto,
  CardPaymentDto,
  FullStatePayload,
  GameDto,
  GamePhaseStartingCardSelection,
  GameStatusActive,
  GameStatusLobby,
  PlayerDisconnectedPayload,
  PlayerDto,
  PlayerActionDto,
  ResourceType,
} from "@/types/generated/api-types.ts";
import {
  UnplayableReason,
  fetchAllCards,
} from "@/utils/cardPlayabilityUtils.ts";
import {
  shouldShowPaymentModal,
  createDefaultPayment,
} from "@/utils/paymentUtils.ts";
import { deepClone, findChangedPaths } from "@/utils/deepCompare.ts";
import { StandardProject } from "@/types/cards.tsx";

export default function GameInterface() {
  const location = useLocation();
  const navigate = useNavigate();
  const [game, setGame] = useState<GameDto | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [isReconnecting, setIsReconnecting] = useState(false);
  const [reconnectionStep, setReconnectionStep] = useState<
    "game" | "environment" | null
  >(null);
  const [currentPlayer, setCurrentPlayer] = useState<PlayerDto | null>(null);
  const [playerId, setPlayerId] = useState<string | null>(null); // Track player ID separately
  const [showCorporationModal, setShowCorporationModal] = useState(false);
  const [corporationData, setCorporationData] = useState<CardDto | null>(null);

  // New modal states
  const [showCardsPlayedModal, setShowCardsPlayedModal] = useState(false);
  const [showVictoryPointsModal, setShowVictoryPointsModal] = useState(false);
  const [showCardEffectsModal, setShowCardEffectsModal] = useState(false);
  const [showActionsModal, setShowActionsModal] = useState(false);
  const [showStandardProjectsPopover, setShowStandardProjectsPopover] =
    useState(false);
  const [showDebugDropdown, setShowDebugDropdown] = useState(false);
  const standardProjectsButtonRef = useRef<HTMLButtonElement>(null);

  // Played cards state
  const [playedCards, setPlayedCards] = useState<CardDto[]>([]);

  // Fetch and resolve played cards when currentPlayer changes
  useEffect(() => {
    const loadPlayedCards = async () => {
      if (
        !currentPlayer?.playedCards ||
        currentPlayer.playedCards.length === 0
      ) {
        setPlayedCards([]);
        return;
      }

      try {
        const allCards = await fetchAllCards();
        const resolvedCards: CardDto[] = [];

        for (const cardId of currentPlayer.playedCards) {
          const card = allCards.get(cardId);
          if (card) {
            resolvedCards.push(card);
          }
        }

        setPlayedCards(resolvedCards);
      } catch (error) {
        console.error("Failed to load played cards:", error);
        setPlayedCards([]);
      }
    };

    void loadPlayedCards();
  }, [currentPlayer?.playedCards]);

  // Set corporation data directly from player (backend now sends full CardDto)
  useEffect(() => {
    if (currentPlayer?.corporation) {
      setCorporationData(currentPlayer.corporation);
    } else {
      setCorporationData(null);
    }
  }, [currentPlayer?.corporation]);

  // Production phase modal state
  const [showProductionPhaseModal, setShowProductionPhaseModal] =
    useState(false);
  const [isProductionModalHidden, setIsProductionModalHidden] = useState(false);
  const [openProductionToCardSelection, setOpenProductionToCardSelection] =
    useState(false);
  const isInitialMount = useRef(true);

  // Card selection state
  const [showCardSelection, setShowCardSelection] = useState(false);
  const [cardDetails, setCardDetails] = useState<CardDto[]>([]);

  // Pending card selection state (for sell patents, etc.)
  const [showPendingCardSelection, setShowPendingCardSelection] =
    useState(false);

  // Card draw selection state (for card-draw/peek/take/buy effects)
  const [showCardDrawSelection, setShowCardDrawSelection] = useState(false);

  // Unplayable card feedback state
  const [unplayableCard, setUnplayableCard] = useState<CardDto | null>(null);
  const [unplayableReason, setUnplayableReason] =
    useState<UnplayableReason | null>(null);

  // Choice selection state (for card play)
  const [showChoiceSelection, setShowChoiceSelection] = useState(false);
  const [cardPendingChoice, setCardPendingChoice] = useState<CardDto | null>(
    null,
  );
  const [pendingCardBehaviorIndex, setPendingCardBehaviorIndex] = useState(0);

  // Action choice selection state (for playing actions with choices)
  const [showActionChoiceSelection, setShowActionChoiceSelection] =
    useState(false);
  const [actionPendingChoice, setActionPendingChoice] =
    useState<PlayerActionDto | null>(null);

  // Payment selection state
  const [showPaymentSelection, setShowPaymentSelection] = useState(false);
  const [pendingCardPayment, setPendingCardPayment] = useState<{
    card: CardDto;
    choiceIndex?: number;
  } | null>(null);

  // Card storage selection state
  const [showCardStorageSelection, setShowCardStorageSelection] =
    useState(false);
  const [pendingCardStorage, setPendingCardStorage] = useState<{
    cardId: string;
    payment: CardPaymentDto;
    choiceIndex?: number;
    resourceType: ResourceType;
    amount: number;
  } | null>(null);

  // Action storage selection state
  const [showActionStorageSelection, setShowActionStorageSelection] =
    useState(false);
  const [pendingActionStorage, setPendingActionStorage] = useState<{
    cardId: string;
    behaviorIndex: number;
    choiceIndex?: number;
    resourceType: ResourceType;
    amount: number;
  } | null>(null);

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
        setIsReconnecting(false);
        setReconnectionStep(null);
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
        // Attempt to reconnect in place
        void attemptReconnection();
      } else {
        // No saved game data, go to main menu
        navigate("/", { replace: true });
      }
    }
  }, [navigate]);

  const handlePlayerDisconnected = useCallback(
    (_payload: PlayerDisconnectedPayload) => {
      // Handle when any player disconnects (NOT this client)
      // Player disconnected from the game
      // Note: PlayerDisconnectedPayload no longer contains game data
      // Game state updates will come through separate game-updated events
    },
    [handleGameUpdated],
  );

  // Check if we should show production phase modal based on game state
  useEffect(() => {
    if (!game || !currentPlayer) return;

    // Check if current player has production phase data
    const hasProductionData =
      currentPlayer.productionPhase &&
      currentPlayer.productionPhase.availableCards &&
      currentPlayer.productionPhase.availableCards.length >= 0;

    if (hasProductionData && !showProductionPhaseModal) {
      // Only play sound if this is not the initial mount (skip on page reload)
      if (!isInitialMount.current) {
        void audioService.playProductionSound();
      }
      setShowProductionPhaseModal(true);
      // Reset the flag for opening directly to card selection on new production phase
      setOpenProductionToCardSelection(false);
    } else if (!hasProductionData && showProductionPhaseModal) {
      // Production phase is over, hide the modal
      setShowProductionPhaseModal(false);
    }

    // Mark that initial mount is complete
    if (isInitialMount.current) {
      isInitialMount.current = false;
    }
  }, [currentPlayer?.productionPhase, game, showProductionPhaseModal]);

  const handleCardSelection = useCallback(
    async (selectedCardIds: string[], corporationId: string) => {
      try {
        // Send card and corporation selection to server - commits immediately
        await globalWebSocketManager.selectStartingCard(
          selectedCardIds,
          corporationId,
        );
        // Modal will close automatically when backend clears startingSelection
      } catch (error) {
        console.error("Failed to select cards:", error);
      }
    },
    [],
  );

  // Handle pending card selection (sell patents, etc.)
  const handlePendingCardSelection = useCallback(
    async (selectedCardIds: string[]) => {
      try {
        await globalWebSocketManager.selectCards(selectedCardIds);
        // Overlay closes automatically when backend clears pendingCardSelection
      } catch (error) {
        console.error("Failed to select cards:", error);
      }
    },
    [],
  );

  // Handle card draw selection confirmation
  const handleCardDrawConfirm = useCallback(
    async (cardsToTake: string[], cardsToBuy: string[]) => {
      try {
        await globalWebSocketManager.confirmCardDraw(cardsToTake, cardsToBuy);
        // Overlay closes automatically when backend clears pendingCardDrawSelection
      } catch (error) {
        console.error("Failed to confirm card draw:", error);
      }
    },
    [],
  );

  // Helper function to check if outputs need card storage selection
  const needsCardStorageSelection = useCallback(
    (
      outputs: any[] | undefined,
    ): { resourceType: ResourceType; amount: number } | null => {
      if (!outputs) return null;

      // Check for any-card targets with storage resource types
      const storageResources = [
        "animals",
        "microbes",
        "floaters",
        "science",
        "asteroid",
      ] as ResourceType[];

      for (const output of outputs) {
        if (
          output.target === "any-card" &&
          storageResources.includes(output.type as ResourceType)
        ) {
          return {
            resourceType: output.type as ResourceType,
            amount: output.amount || 1,
          };
        }
      }

      return null;
    },
    [],
  );

  const handlePlayCard = useCallback(
    async (cardId: string) => {
      try {
        // Block card plays when tile selection is pending
        if (currentPlayer?.pendingTileSelection) {
          const card = currentPlayer?.cards.find((c) => c.id === cardId);
          if (card) {
            setUnplayableCard(card);
            setUnplayableReason({
              type: "phase",
              requirement: null,
              message: "Complete tile placement first",
            });
          }
          return;
        }

        // Find the card to check if it has choices
        const card = currentPlayer?.cards.find((c) => c.id === cardId);
        if (!card) {
          console.error(`Card ${cardId} not found in player's hand`);
          return;
        }

        // Check if any AUTO-triggered behavior has choices
        // Manual-triggered behaviors (actions) will show choices when the action is played
        const behaviorWithChoices = card.behaviors?.findIndex(
          (b) =>
            b.choices &&
            b.choices.length > 0 &&
            b.triggers?.some((t) => t.type === "auto"),
        );

        if (
          behaviorWithChoices !== undefined &&
          behaviorWithChoices >= 0 &&
          card.behaviors?.[behaviorWithChoices]?.choices
        ) {
          // Card has auto-triggered choices, show the choice selection popover
          setCardPendingChoice(card);
          setPendingCardBehaviorIndex(behaviorWithChoices);
          setShowChoiceSelection(true);
        } else {
          // No auto-triggered choices, check if we need payment modal
          if (
            currentPlayer &&
            shouldShowPaymentModal(
              card,
              currentPlayer.resources,
              currentPlayer.paymentSubstitutes,
            )
          ) {
            // Show payment selection modal
            setPendingCardPayment({
              card: card,
              choiceIndex: undefined,
            });
            setShowPaymentSelection(true);
          } else {
            // No payment modal needed, use default all-credits payment
            const payment = createDefaultPayment(card.cost);

            // Check if card needs storage selection
            const autoTriggerBehaviors = card.behaviors?.filter((b) =>
              b.triggers?.some((t) => t.type === "auto"),
            );

            let storageNeeded: {
              resourceType: ResourceType;
              amount: number;
            } | null = null;
            for (const behavior of autoTriggerBehaviors || []) {
              storageNeeded = needsCardStorageSelection(behavior.outputs);
              if (storageNeeded) break;
            }

            if (storageNeeded) {
              // Show storage selection popover
              setPendingCardStorage({
                cardId: card.id,
                payment: payment,
                resourceType: storageNeeded.resourceType,
                amount: storageNeeded.amount,
              });
              setShowCardStorageSelection(true);
            } else {
              // No storage needed, play the card directly
              await globalWebSocketManager.playCard(cardId, payment);
            }
          }
        }
      } catch (error) {
        console.error(`❌ Failed to play card ${cardId}:`, error);
        throw error; // Re-throw to allow CardFanOverlay to handle the error
      }
    },
    [currentPlayer?.cards, needsCardStorageSelection],
  );

  const handleChoiceSelect = useCallback(
    async (choiceIndex: number) => {
      if (!cardPendingChoice || !currentPlayer) return;

      try {
        setShowChoiceSelection(false);

        // Check if we need payment modal
        if (
          shouldShowPaymentModal(
            cardPendingChoice,
            currentPlayer.resources,
            currentPlayer.paymentSubstitutes,
          )
        ) {
          // Show payment selection modal
          setPendingCardPayment({
            card: cardPendingChoice,
            choiceIndex: choiceIndex,
          });
          setShowPaymentSelection(true);
          setCardPendingChoice(null);
          setPendingCardBehaviorIndex(0);
        } else {
          // No payment modal needed, use default all-credits payment
          const payment = createDefaultPayment(cardPendingChoice.cost);

          // Get the selected choice
          const behavior =
            cardPendingChoice.behaviors?.[pendingCardBehaviorIndex];
          const selectedChoice = behavior?.choices?.[choiceIndex];

          // Check if the selected choice outputs need card storage selection
          const storageInfo = needsCardStorageSelection(
            selectedChoice?.outputs,
          );

          if (storageInfo) {
            // Show card storage selection popover
            setPendingCardStorage({
              cardId: cardPendingChoice.id,
              payment: payment,
              choiceIndex: choiceIndex,
              resourceType: storageInfo.resourceType,
              amount: storageInfo.amount,
            });
            setShowCardStorageSelection(true);
            setCardPendingChoice(null);
            setPendingCardBehaviorIndex(0);
          } else {
            // No card storage needed, play the card directly
            await globalWebSocketManager.playCard(
              cardPendingChoice.id,
              payment,
              choiceIndex,
            );
            setCardPendingChoice(null);
            setPendingCardBehaviorIndex(0);
          }
        }
      } catch (error) {
        console.error(
          `❌ Failed to play card ${cardPendingChoice.id} with choice ${choiceIndex}:`,
          error,
        );
        setCardPendingChoice(null);
        setPendingCardBehaviorIndex(0);
      }
    },
    [
      cardPendingChoice,
      currentPlayer,
      pendingCardBehaviorIndex,
      needsCardStorageSelection,
    ],
  );

  const handleChoiceCancel = useCallback(() => {
    setShowChoiceSelection(false);
    setCardPendingChoice(null);
    setPendingCardBehaviorIndex(0);
  }, []);

  const handleActionChoiceSelect = useCallback(
    async (choiceIndex: number) => {
      if (!actionPendingChoice) return;

      try {
        setShowActionChoiceSelection(false);

        // Get the selected choice
        const selectedChoice =
          actionPendingChoice.behavior.choices?.[choiceIndex];

        // Check if the selected choice outputs need card storage selection
        const storageInfo = needsCardStorageSelection(selectedChoice?.outputs);

        if (storageInfo) {
          // Show action storage selection popover
          setPendingActionStorage({
            cardId: actionPendingChoice.cardId,
            behaviorIndex: actionPendingChoice.behaviorIndex,
            choiceIndex: choiceIndex,
            resourceType: storageInfo.resourceType,
            amount: storageInfo.amount,
          });
          setShowActionStorageSelection(true);
          setActionPendingChoice(null);
        } else {
          // No card storage needed, execute action directly
          await globalWebSocketManager.playCardAction(
            actionPendingChoice.cardId,
            actionPendingChoice.behaviorIndex,
            choiceIndex,
          );
          setActionPendingChoice(null);
        }
      } catch (error) {
        console.error(
          `❌ Failed to play action ${actionPendingChoice.cardId} with choice ${choiceIndex}:`,
          error,
        );
        setActionPendingChoice(null);
      }
    },
    [actionPendingChoice, needsCardStorageSelection],
  );

  const handleActionChoiceCancel = useCallback(() => {
    setShowActionChoiceSelection(false);
    setActionPendingChoice(null);
  }, []);

  // Payment selection callbacks
  const handlePaymentConfirm = useCallback(
    async (payment: CardPaymentDto) => {
      if (!pendingCardPayment || !currentPlayer) return;

      try {
        setShowPaymentSelection(false);

        // Check if card storage selection is still needed
        const autoTriggerBehaviors = pendingCardPayment.card.behaviors?.filter(
          (b) =>
            b.triggers &&
            b.triggers.length > 0 &&
            b.triggers.some((t) => t.type === "auto"),
        );

        let storageNeeded: {
          resourceType: ResourceType;
          amount: number;
        } | null = null;
        for (const behavior of autoTriggerBehaviors || []) {
          storageNeeded = needsCardStorageSelection(behavior.outputs);
          if (storageNeeded) break;
        }

        if (storageNeeded) {
          // Show storage selection popover
          setPendingCardStorage({
            cardId: pendingCardPayment.card.id,
            payment: payment,
            choiceIndex: pendingCardPayment.choiceIndex,
            resourceType: storageNeeded.resourceType,
            amount: storageNeeded.amount,
          });
          setShowCardStorageSelection(true);
        } else {
          // No storage needed, play the card directly
          await globalWebSocketManager.playCard(
            pendingCardPayment.card.id,
            payment,
            pendingCardPayment.choiceIndex,
          );
        }

        setPendingCardPayment(null);
      } catch (error) {
        console.error(`❌ Failed to play card with payment:`, error);
        setPendingCardPayment(null);
      }
    },
    [pendingCardPayment, currentPlayer, needsCardStorageSelection],
  );

  const handlePaymentCancel = useCallback(() => {
    setShowPaymentSelection(false);
    setPendingCardPayment(null);
  }, []);

  const handleCardStorageSelect = useCallback(
    async (targetCardId: string) => {
      if (!pendingCardStorage) return;

      try {
        setShowCardStorageSelection(false);
        await globalWebSocketManager.playCard(
          pendingCardStorage.cardId,
          pendingCardStorage.payment,
          pendingCardStorage.choiceIndex,
          targetCardId,
        );
        setPendingCardStorage(null);
      } catch (error) {
        console.error(
          `❌ Failed to play card ${pendingCardStorage.cardId} with card storage target ${targetCardId}:`,
          error,
        );
        setPendingCardStorage(null);
      }
    },
    [pendingCardStorage],
  );

  const handleCardStorageCancel = useCallback(() => {
    setShowCardStorageSelection(false);
    setPendingCardStorage(null);
  }, []);

  const handleActionStorageSelect = useCallback(
    async (targetCardId: string) => {
      if (!pendingActionStorage) return;

      try {
        setShowActionStorageSelection(false);
        await globalWebSocketManager.playCardAction(
          pendingActionStorage.cardId,
          pendingActionStorage.behaviorIndex,
          pendingActionStorage.choiceIndex,
          targetCardId,
        );
        setPendingActionStorage(null);
      } catch (error) {
        console.error(
          `❌ Failed to play action ${pendingActionStorage.cardId} with card storage target ${targetCardId}:`,
          error,
        );
        setPendingActionStorage(null);
      }
    },
    [pendingActionStorage],
  );

  const handleActionStorageCancel = useCallback(() => {
    setShowActionStorageSelection(false);
    setPendingActionStorage(null);
  }, []);

  const handleUnplayableCard = useCallback(
    (card: CardDto | null, reason: UnplayableReason | null) => {
      setUnplayableCard(card);
      setUnplayableReason(reason);
    },
    [],
  );

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

      // Step 1: Reconnect to game
      setReconnectionStep("game");

      // Fetch current game state from server first
      const response = await fetch(
        `http://localhost:3001/api/v1/games/${gameId}`,
      );
      if (!response.ok) {
        throw new Error(`Game not found: ${response.status}`);
      }

      const gameData = await response.json();

      // Update local state with fetched game data
      setGame(gameData.game);
      setPlayerId(playerId);

      // Set current player from fetched game data
      const player = gameData.game.currentPlayer;
      setCurrentPlayer(player || null);

      // Store player ID for WebSocket handlers
      currentPlayerIdRef.current = playerId;

      // Step 2: Ensure 3D environment is loaded
      if (!skyboxCache.isReady()) {
        setReconnectionStep("environment");
        await skyboxCache.preload();
      }

      // Now establish WebSocket connection
      await globalWebSocketManager.playerConnect(playerName, gameId, playerId);
    } catch (error) {
      console.error("❌ Reconnection failed:", error);
      setIsReconnecting(false);
      setReconnectionStep(null);
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
    globalWebSocketManager.on("player-disconnected", handlePlayerDisconnected);
    globalWebSocketManager.on("error", handleError);
    globalWebSocketManager.on("disconnect", handleDisconnect);

    isWebSocketInitialized.current = true;

    return () => {
      globalWebSocketManager.off("game-updated", handleGameUpdated);
      globalWebSocketManager.off("full-state", handleFullState);
      globalWebSocketManager.off(
        "player-disconnected",
        handlePlayerDisconnected,
      );
      globalWebSocketManager.off("error", handleError);
      globalWebSocketManager.off("disconnect", handleDisconnect);
      isWebSocketInitialized.current = false;
    };
  }, [
    handleGameUpdated,
    handleFullState,
    handlePlayerDisconnected,
    handleError,
    handleDisconnect,
  ]);

  // Handle action selection from card actions
  const handleActionSelect = useCallback(
    (action: PlayerActionDto) => {
      // Block actions when tile selection is pending
      if (currentPlayer?.pendingTileSelection) {
        return;
      }

      // Check if this action has choices
      if (action.behavior.choices && action.behavior.choices.length > 0) {
        // Action has choices, show the choice selection popover
        setActionPendingChoice(action);
        setShowActionChoiceSelection(true);
      } else {
        // No choices, check if action outputs need card storage selection
        const storageInfo = needsCardStorageSelection(action.behavior.outputs);

        if (storageInfo) {
          // Show action storage selection popover
          setPendingActionStorage({
            cardId: action.cardId,
            behaviorIndex: action.behaviorIndex,
            resourceType: storageInfo.resourceType,
            amount: storageInfo.amount,
          });
          setShowActionStorageSelection(true);
        } else {
          // No card storage needed, execute action directly
          void globalWebSocketManager.playCardAction(
            action.cardId,
            action.behaviorIndex,
          );
        }
      }
    },
    [currentPlayer?.pendingTileSelection, needsCardStorageSelection],
  );

  // Standard project selection handler
  const handleStandardProjectSelect = useCallback(
    (project: StandardProject) => {
      // Block standard projects when tile selection is pending
      if (currentPlayer?.pendingTileSelection) {
        return;
      }

      // Close dropdown first
      setShowStandardProjectsPopover(false);

      // All standard projects execute immediately
      // Backend will create tile queue for projects requiring placement
      switch (project) {
        case StandardProject.SELL_PATENTS:
          // Initiate sell patents - backend will create pendingCardSelection
          void globalWebSocketManager.sellPatents();
          break;
        case StandardProject.POWER_PLANT:
          void globalWebSocketManager.buildPowerPlant();
          break;
        case StandardProject.ASTEROID:
          void globalWebSocketManager.launchAsteroid();
          break;
        case StandardProject.AQUIFER:
          void globalWebSocketManager.buildAquifer();
          break;
        case StandardProject.GREENERY:
          void globalWebSocketManager.plantGreenery();
          break;
        case StandardProject.CITY:
          void globalWebSocketManager.buildCity();
          break;
      }
    },
    [currentPlayer?.pendingTileSelection],
  );

  // Resource conversion handlers
  const handleConvertPlantsToGreenery = useCallback(() => {
    // Block if tile selection is already pending
    if (currentPlayer?.pendingTileSelection) {
      return;
    }

    // Initiate plant conversion (backend creates pending tile selection)
    void globalWebSocketManager.convertPlantsToGreenery();
  }, [currentPlayer?.pendingTileSelection]);

  const handleConvertHeatToTemperature = useCallback(() => {
    // Convert heat to temperature directly (no tile selection needed)
    void globalWebSocketManager.convertHeatToTemperature();
  }, []);

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
          // Start in-place reconnection instead of redirecting
          setIsReconnecting(true);
          void attemptReconnection();
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
    setCardDetails(cards);
  }, []);

  // Show/hide starting card selection overlay based on backend state
  useEffect(() => {
    const cards = game?.currentPlayer?.selectStartingCardsPhase?.availableCards;
    const hasCardSelection = cards && cards.length > 0;

    if (
      game?.currentPhase === GamePhaseStartingCardSelection &&
      game?.status === GameStatusActive &&
      hasCardSelection &&
      !showCardSelection
    ) {
      extractCardDetails(cards);
      setShowCardSelection(true);
    } else if (showCardSelection && !hasCardSelection) {
      setShowCardSelection(false);
    }
  }, [
    game?.currentPhase,
    game?.status,
    game?.currentPlayer?.selectStartingCardsPhase?.availableCards,
    showCardSelection,
    extractCardDetails,
  ]);

  // Show/hide pending card selection overlay (sell patents, etc.)
  useEffect(() => {
    const pendingSelection = game?.currentPlayer?.pendingCardSelection;

    if (pendingSelection && !showPendingCardSelection) {
      setShowPendingCardSelection(true);
    } else if (!pendingSelection && showPendingCardSelection) {
      setShowPendingCardSelection(false);
    }
  }, [game?.currentPlayer?.pendingCardSelection, showPendingCardSelection]);

  // Show/hide card draw selection overlay
  useEffect(() => {
    const pendingDrawSelection = game?.currentPlayer?.pendingCardDrawSelection;

    if (pendingDrawSelection && !showCardDrawSelection) {
      setShowCardDrawSelection(true);
    } else if (!pendingDrawSelection && showCardDrawSelection) {
      setShowCardDrawSelection(false);
    }
  }, [game?.currentPlayer?.pendingCardDrawSelection, showCardDrawSelection]);

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
            setShowVictoryPointsModal(true);
            break;
          case "4":
            event.preventDefault();
            // Actions are now handled via popover in BottomResourceBar
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
    let loadingMessage = "Connecting to game...";

    if (isReconnecting && reconnectionStep) {
      if (reconnectionStep === "game") {
        loadingMessage = "Reconnecting to game...";
      } else if (reconnectionStep === "environment") {
        loadingMessage = "Loading 3D environment...";
      }
    }

    return (
      <div
        style={{
          position: "fixed",
          top: 0,
          left: 0,
          width: "100%",
          height: "100%",
          background: "#000011",
          zIndex: 9999,
        }}
      >
        <LoadingSpinner message={loadingMessage} />
      </div>
    );
  }

  // Check if any modal is currently open
  const isAnyModalOpen =
    showCorporationModal ||
    showCardsPlayedModal ||
    showVictoryPointsModal ||
    showCardEffectsModal ||
    showActionsModal ||
    showProductionPhaseModal;

  // Check if game is in lobby phase
  const isLobbyPhase = game?.status === GameStatusLobby;

  // Check if we need the persistent backdrop (during overlay transitions)
  const shouldShowBackdrop = isLobbyPhase || showCardSelection;

  return (
    <>
      {/* Dev Mode Chip - Always visible in dev mode */}
      {game?.settings?.developmentMode && <DevModeChip />}

      {/* Persistent backdrop for overlays to prevent blink during transitions */}
      {shouldShowBackdrop && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-[999] animate-[backdropFadeIn_0.3s_ease-out]">
          <style>{`
            @keyframes backdropFadeIn {
              0% {
                opacity: 0;
              }
              100% {
                opacity: 1;
              }
            }
          `}</style>
        </div>
      )}

      <GameLayout
        gameState={game}
        currentPlayer={currentPlayer}
        playedCards={playedCards}
        corporationCard={corporationData}
        isAnyModalOpen={isAnyModalOpen}
        isLobbyPhase={isLobbyPhase}
        showCardSelection={showCardSelection}
        changedPaths={changedPaths}
        onOpenCardEffectsModal={() => setShowCardEffectsModal(true)}
        onOpenCardsPlayedModal={() => setShowCardsPlayedModal(true)}
        onOpenVictoryPointsModal={() => setShowVictoryPointsModal(true)}
        onOpenActionsModal={() => setShowActionsModal(true)}
        onActionSelect={handleActionSelect}
        onConvertPlantsToGreenery={handleConvertPlantsToGreenery}
        onConvertHeatToTemperature={handleConvertHeatToTemperature}
        showStandardProjectsPopover={showStandardProjectsPopover}
        onToggleStandardProjectsPopover={() =>
          setShowStandardProjectsPopover(!showStandardProjectsPopover)
        }
        standardProjectsButtonRef={standardProjectsButtonRef}
      />

      <CardsPlayedModal
        isVisible={showCardsPlayedModal}
        onClose={() => setShowCardsPlayedModal(false)}
        cards={playedCards}
      />

      <VictoryPointsModal
        isVisible={showVictoryPointsModal}
        onClose={() => setShowVictoryPointsModal(false)}
        cards={[]}
        terraformRating={currentPlayer?.terraformRating}
      />

      <EffectsModal
        isVisible={showCardEffectsModal}
        onClose={() => setShowCardEffectsModal(false)}
        effects={currentPlayer?.effects || []}
      />

      <ActionsModal
        isVisible={showActionsModal}
        onClose={() => setShowActionsModal(false)}
        actions={currentPlayer?.actions || []}
        onActionSelect={handleActionSelect}
        gameState={game}
      />

      <StandardProjectPopover
        isVisible={showStandardProjectsPopover}
        onClose={() => setShowStandardProjectsPopover(false)}
        onProjectSelect={handleStandardProjectSelect}
        gameState={game}
        anchorRef={standardProjectsButtonRef}
      />

      <ProductionPhaseModal
        isOpen={showProductionPhaseModal && !isProductionModalHidden}
        gameState={game}
        onClose={() => {
          setShowProductionPhaseModal(false);
          setIsProductionModalHidden(false);
          setOpenProductionToCardSelection(false);
        }}
        onHide={() => {
          setIsProductionModalHidden(true);
          setOpenProductionToCardSelection(false);
        }}
        openDirectlyToCardSelection={openProductionToCardSelection}
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
        availableCorporations={
          game?.currentPlayer?.selectStartingCardsPhase
            ?.availableCorporations || []
        }
        playerCredits={currentPlayer?.resources?.credits || 40}
        onSelectCards={handleCardSelection}
      />

      {/* Pending card selection overlay (sell patents, etc.) */}
      {game?.currentPlayer?.pendingCardSelection && (
        <PendingCardSelectionOverlay
          isOpen={showPendingCardSelection}
          selection={game.currentPlayer.pendingCardSelection}
          playerCredits={currentPlayer?.resources?.credits || 0}
          onSelectCards={handlePendingCardSelection}
        />
      )}

      {/* Card draw selection overlay (card-draw/peek/take/buy effects) */}
      {game?.currentPlayer?.pendingCardDrawSelection && (
        <CardDrawSelectionOverlay
          isOpen={showCardDrawSelection}
          selection={game.currentPlayer.pendingCardDrawSelection}
          playerCredits={currentPlayer?.resources?.credits || 0}
          onConfirm={handleCardDrawConfirm}
        />
      )}

      {/* Card fan overlay for hand cards */}
      {game && currentPlayer && (
        <CardFanOverlay
          cards={currentPlayer.cards || []}
          game={game}
          player={currentPlayer}
          hideWhenModalOpen={
            showCardSelection ||
            showPendingCardSelection ||
            showCardDrawSelection ||
            isLobbyPhase
          }
          onCardSelect={(_cardId) => {
            // TODO: Implement card selection logic (view details, etc.)
          }}
          onPlayCard={handlePlayCard}
          onUnplayableCard={handleUnplayableCard}
        />
      )}

      {/* Hexagonal shield overlay */}
      <HexagonalShieldOverlay
        card={unplayableCard}
        reason={unplayableReason}
        isVisible={unplayableCard !== null}
      />

      {/* Choice selection popover for card play */}
      {cardPendingChoice && (
        <ChoiceSelectionPopover
          cardId={cardPendingChoice.id}
          cardName={cardPendingChoice.name}
          behaviors={cardPendingChoice.behaviors || []}
          behaviorIndex={pendingCardBehaviorIndex}
          onChoiceSelect={handleChoiceSelect}
          onCancel={handleChoiceCancel}
          isVisible={showChoiceSelection}
          playerResources={currentPlayer?.resources}
          resourceStorage={currentPlayer?.resourceStorage}
        />
      )}

      {/* Choice selection popover for actions */}
      {actionPendingChoice && (
        <ChoiceSelectionPopover
          cardId={actionPendingChoice.cardId}
          cardName={actionPendingChoice.cardName}
          behaviors={[actionPendingChoice.behavior]}
          behaviorIndex={0}
          onChoiceSelect={handleActionChoiceSelect}
          onCancel={handleActionChoiceCancel}
          isVisible={showActionChoiceSelection}
          isAction={true}
          playerResources={currentPlayer?.resources}
          resourceStorage={currentPlayer?.resourceStorage}
        />
      )}

      {/* Card storage selection popover */}
      {pendingCardStorage && (
        <CardStorageSelectionPopover
          resourceType={pendingCardStorage.resourceType}
          amount={pendingCardStorage.amount}
          playedCards={playedCards}
          resourceStorage={currentPlayer?.resourceStorage}
          onCardSelect={handleCardStorageSelect}
          onCancel={handleCardStorageCancel}
          isVisible={showCardStorageSelection}
        />
      )}

      {/* Payment selection popover */}
      {pendingCardPayment && game && currentPlayer && (
        <PaymentSelectionPopover
          cardId={pendingCardPayment.card.id}
          card={pendingCardPayment.card}
          playerResources={currentPlayer.resources}
          paymentConstants={game.paymentConstants}
          playerPaymentSubstitutes={currentPlayer.paymentSubstitutes}
          onConfirm={handlePaymentConfirm}
          onCancel={handlePaymentCancel}
          isVisible={showPaymentSelection}
        />
      )}

      {/* Action storage selection popover */}
      {pendingActionStorage && (
        <CardStorageSelectionPopover
          resourceType={pendingActionStorage.resourceType}
          amount={pendingActionStorage.amount}
          playedCards={playedCards}
          resourceStorage={currentPlayer?.resourceStorage}
          onCardSelect={handleActionStorageSelect}
          onCancel={handleActionStorageCancel}
          isVisible={showActionStorageSelection}
        />
      )}

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

      {/* Return to Production Modal button */}
      {showProductionPhaseModal && isProductionModalHidden && (
        <button
          className="fixed top-[80px] left-[70%] bg-space-black-darker/95 border-2 border-space-blue-400 rounded-xl text-white text-base font-semibold py-3.5 px-7 cursor-pointer transition-all duration-300 text-shadow-glow shadow-[0_4px_15px_rgba(0,0,0,0.5),0_0_20px_rgba(30,60,150,0.4)] backdrop-blur-space z-[1000] whitespace-nowrap hover:bg-space-black-darker hover:border-space-blue-500 hover:-translate-y-0.5 hover:shadow-[0_6px_20px_rgba(0,0,0,0.6),0_0_35px_rgba(30,60,150,0.6)] active:translate-y-0 active:shadow-[0_2px_10px_rgba(0,0,0,0.4),0_0_20px_rgba(30,60,150,0.4)]"
          onClick={() => {
            setIsProductionModalHidden(false);
            setOpenProductionToCardSelection(true);
          }}
        >
          Return to Production
        </button>
      )}
    </>
  );
}
