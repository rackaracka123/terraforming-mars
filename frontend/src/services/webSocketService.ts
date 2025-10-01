import { v4 as uuidv4 } from "uuid";
import {
  ErrorPayload,
  FullStatePayload,
  GameUpdatedPayload,
  MessageType,
  MessageTypeError,
  MessageTypeFullState,
  MessageTypeGameUpdated,
  MessageTypePlayerConnect,
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
  MessageTypeActionSelectCards,
  MessageTypeActionSelectStartingCard,
  MessageTypeActionTileSelected,
  // Payload types
  PlayerDisconnectedPayload,
  WebSocketMessage,
  HexPositionDto,
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
  private isConnecting = false;
  private shouldReconnect = true;

  constructor(url: string = "ws://localhost:3001/ws") {
    this.url = url;
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        // If already connected, resolve immediately
        if (
          this.isConnected &&
          this.ws &&
          this.ws.readyState === WebSocket.OPEN
        ) {
          resolve();
          return;
        }

        // Prevent multiple concurrent connection attempts
        if (this.isConnecting) {
          resolve();
          return;
        }

        this.isConnecting = true;

        // Close existing connection if it exists
        if (this.ws) {
          this.ws.close();
        }

        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
          this.isConnected = true;
          this.isConnecting = false;
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
          this.isConnecting = false;
          this.emit("disconnect");

          // Only attempt reconnect if it was an unexpected closure and we should reconnect
          if (this.shouldReconnect && event.code !== 1000) {
            this.attemptReconnect();
          }
        };

        this.ws.onerror = (error) => {
          console.error("WebSocket error:", error);
          this.isConnecting = false;
          this.emit("error", error);
          if (!this.isConnected) {
            reject(error);
          }
        };
      } catch (error) {
        reject(error);
      }
    });
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

  playerConnect(
    playerName: string,
    gameId: string,
    playerId?: string,
  ): Promise<any> {
    return new Promise((resolve, reject) => {
      // Send the connect message with playerId if available (for reconnection)
      const payload: any = { playerName, gameId };
      if (playerId) {
        payload.playerId = playerId;
      }

      this.send(MessageTypePlayerConnect, payload, gameId);
      this.currentGameId = gameId;

      // Set up timeout
      const timeout = setTimeout(() => {
        this.off("game-updated", gameUpdatedHandler);
        this.off("error", errorHandler);
        reject(new Error("Timeout waiting for player connection confirmation"));
      }, 10000); // 10 second timeout

      // Handler for game updates (which indicate successful connection)
      const gameUpdatedHandler = (payload: any) => {
        // Extract the actual game data from the payload
        const gameData = payload.game || payload;

        // GameDto has currentPlayer and otherPlayers instead of players array
        const allPlayers = [];
        if (gameData.currentPlayer) {
          allPlayers.push(gameData.currentPlayer);
        }
        if (gameData.otherPlayers) {
          allPlayers.push(...gameData.otherPlayers);
        }

        const connectedPlayer = allPlayers.find(
          (p: any) => p.name === playerName,
        );

        if (connectedPlayer) {
          clearTimeout(timeout);
          this.off("game-updated", gameUpdatedHandler);
          this.off("error", errorHandler);
          this.currentPlayerId = connectedPlayer.id;

          // Return data similar to the old PlayerConnectedPayload format
          resolve({
            playerId: connectedPlayer.id,
            playerName: connectedPlayer.name,
            gameId: gameId,
            game: gameData, // Use 'game' instead of 'gameData' for consistency
          });
        }
      };

      // Error handler
      const errorHandler = (errorPayload: ErrorPayload) => {
        clearTimeout(timeout);
        this.off("game-updated", gameUpdatedHandler);
        this.off("error", errorHandler);
        reject(new Error(errorPayload.message || "Connection failed"));
      };

      // Listen for game updates and errors
      this.on("game-updated", gameUpdatedHandler);
      this.on("error", errorHandler);
    });
  }

  // Standard project actions
  sellPatents(cardCount: number): string {
    return this.send(MessageTypeActionSellPatents, { cardCount });
  }

  launchAsteroid(): string {
    return this.send(MessageTypeActionLaunchAsteroid, {});
  }

  buildPowerPlant(): string {
    return this.send(MessageTypeActionBuildPowerPlant, {});
  }

  buildAquifer(hexPosition: HexPositionDto): string {
    return this.send(MessageTypeActionBuildAquifer, { hexPosition });
  }

  plantGreenery(hexPosition: HexPositionDto): string {
    return this.send(MessageTypeActionPlantGreenery, { hexPosition });
  }

  buildCity(hexPosition: HexPositionDto): string {
    return this.send(MessageTypeActionBuildCity, { hexPosition });
  }

  // Game management actions
  startGame(): string {
    return this.send(MessageTypeActionStartGame, {});
  }

  skipAction(): string {
    return this.send(MessageTypeActionSkipAction, {});
  }

  // Card actions
  playCard(cardId: string, choiceIndex?: number): string {
    return this.send(MessageTypeActionPlayCard, {
      type: "play-card",
      cardId,
      ...(choiceIndex !== undefined && { choiceIndex }),
    });
  }

  playCardAction(
    cardId: string,
    behaviorIndex: number,
    choiceIndex?: number,
  ): string {
    return this.send(MessageTypeActionCardAction, {
      type: "card-action",
      cardId,
      behaviorIndex,
      ...(choiceIndex !== undefined && { choiceIndex }),
    });
  }

  selectStartingCard(cardIds: string[]): string {
    return this.send(MessageTypeActionSelectStartingCard, { cardIds });
  }

  selectCards(cardIds: string[]): string {
    return this.send(MessageTypeActionSelectCards, { cardIds });
  }

  // Tile selection actions
  selectTile(coordinate: { q: number; r: number; s: number }): string {
    return this.send(MessageTypeActionTileSelected, { coordinate });
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
    this.isConnecting = false;
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
