import { v4 as uuidv4 } from "uuid";
import {
  CardPaymentDto,
  ConfirmDemoSetupRequest,
  ErrorPayload,
  FullStatePayload,
  GameUpdatedPayload,
  MessageType,
  MessageTypeError,
  MessageTypeFullState,
  MessageTypeGameUpdated,
  MessageTypePlayerConnect,
  MessageTypePlayerConnected,
  MessageTypePlayerDisconnected,
  // New message types
  MessageTypeActionSellPatents,
  MessageTypeActionLaunchAsteroid,
  MessageTypeActionBuildPowerPlant,
  MessageTypeActionBuildAquifer,
  MessageTypeActionPlantGreenery,
  MessageTypeActionBuildCity,
  MessageTypeActionStartGame,
  MessageTypeActionSkipAction,
  MessageTypeActionPlayCard,
  MessageTypeActionCardAction,
  MessageTypeActionSelectStartingCard,
  MessageTypeActionConfirmSellPatents,
  MessageTypeActionConfirmProductionCards,
  MessageTypeActionCardDrawConfirmed,
  MessageTypeActionTileSelected,
  MessageTypeActionConvertPlantsToGreenery,
  MessageTypeActionConvertHeatToTemperature,
  MessageTypeActionConfirmDemoSetup,
  MessageTypeActionClaimMilestone,
  MessageTypeActionFundAward,
  // Payload types
  PlayerConnectedPayload,
  PlayerDisconnectedPayload,
  WebSocketMessage,
} from "../types/generated/api-types.ts";

type EventCallback = (data: any) => void;

export class WebSocketService {
  private ws: WebSocket | null = null;
  private readonly url: string;
  private listeners: { [event: string]: EventCallback[] } = {};
  private isConnected = false;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private currentGameId: string | null = null;
  private currentPlayerId: string | null = null;
  private pendingConnection: Promise<void> | null = null;
  private shouldReconnect = true;

  constructor(url: string = "ws://localhost:3001/ws") {
    this.url = url;
  }

  connect(): Promise<void> {
    // If already connected, resolve immediately
    if (this.isConnected && this.ws && this.ws.readyState === WebSocket.OPEN) {
      return Promise.resolve();
    }

    // If already connecting, return the existing pending promise
    if (this.pendingConnection) {
      return this.pendingConnection;
    }

    // Create new connection promise
    this.pendingConnection = new Promise((resolve, reject) => {
      try {
        // Close existing connection if it exists
        if (this.ws) {
          this.ws.close();
        }

        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
          this.isConnected = true;
          this.pendingConnection = null;
          this.reconnectAttempts = 0;
          this.emit("connect");
          resolve();
        };

        this.ws.onmessage = (event) => {
          let message: any;
          try {
            message = JSON.parse(event.data);
          } catch (error) {
            console.error("Failed to parse WebSocket message:", error);
            return;
          }

          try {
            this.handleMessage(message);
          } catch (error) {
            console.error("Error handling WebSocket message:", error);
          }
        };

        this.ws.onclose = (event) => {
          // WebSocket connection closed
          this.isConnected = false;
          this.emit("disconnect");

          // Only attempt reconnect if it was an unexpected closure and we should reconnect
          if (this.shouldReconnect && event.code !== 1000) {
            this.attemptReconnect();
          }
        };

        this.ws.onerror = (error) => {
          console.error("WebSocket error:", error);
          this.pendingConnection = null;
          this.emit("error", error);
          if (!this.isConnected) {
            reject(error);
          }
        };
      } catch (error) {
        this.pendingConnection = null;
        reject(error);
      }
    });

    return this.pendingConnection;
  }

  private handleMessage(message: WebSocketMessage) {
    switch (message.type) {
      case MessageTypeGameUpdated: {
        const gamePayload = message.payload as GameUpdatedPayload;
        // Handle both direct game data and nested structure
        const gameData = gamePayload.game || gamePayload;
        this.emit("game-updated", gameData);
        break;
      }
      case MessageTypePlayerConnected: {
        const connectedPayload = message.payload as PlayerConnectedPayload;
        // This is a confirmation that player joined successfully
        // The full game state will arrive via game-updated from broadcaster
        this.emit("player-connected", connectedPayload);
        break;
      }
      case MessageTypePlayerDisconnected: {
        const disconnectedPayload =
          message.payload as PlayerDisconnectedPayload;
        this.emit("player-disconnected", disconnectedPayload);
        break;
      }
      case MessageTypeError: {
        const errorPayload = message.payload as ErrorPayload;
        this.emit("error", errorPayload);
        break;
      }
      case MessageTypeFullState: {
        const statePayload = message.payload as FullStatePayload;
        this.currentPlayerId = statePayload.playerId;
        this.emit("full-state", statePayload);
        break;
      }
      // Note: production-phase-started is now handled via game state updates
      // The production phase data is available in player.productionSelection
      default:
        console.warn("Unknown message type:", message.type);
    }
  }

  send(type: MessageType, payload: unknown, gameId?: string): string {
    const reqId = uuidv4();

    if (!this.isConnected || !this.ws) {
      throw new Error("WebSocket is not connected");
    }

    const message: WebSocketMessage = {
      type,
      payload,
      gameId: gameId || this.currentGameId || undefined,
    };

    this.ws.send(JSON.stringify(message));

    return reqId;
  }

  playerConnect(playerName: string, gameId: string, playerId?: string): void {
    // Simple send-and-forget pattern - WebSocket guarantees delivery
    // UI will update reactively when backend sends game-updated event
    const payload: any = { playerName, gameId };
    if (playerId) {
      payload.playerId = playerId;
    }

    this.send(MessageTypePlayerConnect, payload, gameId);
    this.currentGameId = gameId;
  }

  // Standard project actions
  sellPatents(): string {
    return this.send(MessageTypeActionSellPatents, {});
  }

  launchAsteroid(): string {
    return this.send(MessageTypeActionLaunchAsteroid, {});
  }

  buildPowerPlant(): string {
    return this.send(MessageTypeActionBuildPowerPlant, {});
  }

  buildAquifer(): string {
    return this.send(MessageTypeActionBuildAquifer, {});
  }

  plantGreenery(): string {
    return this.send(MessageTypeActionPlantGreenery, {});
  }

  buildCity(): string {
    return this.send(MessageTypeActionBuildCity, {});
  }

  // Resource conversion actions
  convertPlantsToGreenery(): string {
    return this.send(MessageTypeActionConvertPlantsToGreenery, {
      type: "convert-plants-to-greenery",
    });
  }

  convertHeatToTemperature(): string {
    return this.send(MessageTypeActionConvertHeatToTemperature, {
      type: "convert-heat-to-temperature",
    });
  }

  // Game management actions
  startGame(): string {
    return this.send(MessageTypeActionStartGame, {});
  }

  skipAction(): string {
    return this.send(MessageTypeActionSkipAction, {});
  }

  // Card actions
  playCard(
    cardId: string,
    payment: CardPaymentDto,
    choiceIndex?: number,
    cardStorageTarget?: string,
  ): string {
    return this.send(MessageTypeActionPlayCard, {
      type: "play-card",
      cardId,
      payment,
      ...(choiceIndex !== undefined && { choiceIndex }),
      ...(cardStorageTarget !== undefined && { cardStorageTarget }),
    });
  }

  playCardAction(
    cardId: string,
    behaviorIndex: number,
    choiceIndex?: number,
    cardStorageTarget?: string,
  ): string {
    return this.send(MessageTypeActionCardAction, {
      type: "card-action",
      cardId,
      behaviorIndex,
      ...(choiceIndex !== undefined && { choiceIndex }),
      ...(cardStorageTarget !== undefined && { cardStorageTarget }),
    });
  }

  selectStartingCard(cardIds: string[], corporationId: string): string {
    return this.send(MessageTypeActionSelectStartingCard, {
      cardIds,
      corporationId,
    });
  }

  selectCards(cardIds: string[]): string {
    return this.send(MessageTypeActionConfirmSellPatents, {
      selectedCardIds: cardIds,
    });
  }

  confirmProductionCards(cardIds: string[]): string {
    return this.send(MessageTypeActionConfirmProductionCards, { cardIds });
  }

  confirmCardDraw(cardsToTake: string[], cardsToBuy: string[]): string {
    return this.send(MessageTypeActionCardDrawConfirmed, {
      cardsToTake,
      cardsToBuy,
    });
  }

  // Tile selection actions
  selectTile(coordinate: { q: number; r: number; s: number }): string {
    // Convert coordinate object to "q,r,s" string format expected by backend
    const hex = `${coordinate.q},${coordinate.r},${coordinate.s}`;
    return this.send(MessageTypeActionTileSelected, { hex });
  }

  // Demo setup
  confirmDemoSetup(request: ConfirmDemoSetupRequest): string {
    return this.send(MessageTypeActionConfirmDemoSetup, request);
  }

  // Milestone and award actions
  claimMilestone(milestoneType: string): string {
    return this.send(MessageTypeActionClaimMilestone, { milestoneType });
  }

  fundAward(awardType: string): string {
    return this.send(MessageTypeActionFundAward, { awardType });
  }

  on(event: string, callback: EventCallback) {
    if (!this.listeners[event]) {
      this.listeners[event] = [];
    }
    this.listeners[event].push(callback);
  }

  off(event: string, callback: EventCallback) {
    if (this.listeners[event]) {
      this.listeners[event] = this.listeners[event].filter(
        (cb) => cb !== callback,
      );
    }
  }

  private emit(event: string, data?: unknown) {
    if (this.listeners[event]) {
      this.listeners[event].forEach((callback) => {
        try {
          callback(data);
        } catch (error) {
          console.error(`Error in event listener for ${event}:`, error);
        }
      });
    }
  }

  private attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;

      setTimeout(() => {
        this.connect().catch((error) => {
          console.error("Reconnection failed:", error);
        });
      }, this.reconnectDelay * this.reconnectAttempts);
    } else {
      console.error("Max reconnection attempts reached");
      this.emit("max-reconnects-reached");
    }
  }

  disconnect() {
    this.shouldReconnect = false;
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.isConnected = false;
    this.currentGameId = null;
    this.currentPlayerId = null;
  }

  get connected() {
    return this.isConnected;
  }

  get playerId() {
    return this.currentPlayerId;
  }

  get gameId() {
    return this.currentGameId;
  }
}

// Singleton instance for application-wide use
export const webSocketService = new WebSocketService();
