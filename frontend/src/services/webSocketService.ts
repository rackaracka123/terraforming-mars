import { v4 as uuidv4 } from "uuid";
import {
  ErrorPayload,
  FullStatePayload,
  GameUpdatedPayload,
  MessageType,
  WebSocketMessage,
  // Payload types
  PlayerConnectedPayload,
  PlayerReconnectedPayload,
  PlayerDisconnectedPayload,
  ProductionPhaseStartedPayload,
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

        // Close existing connection if it exists
        if (this.ws) {
          this.ws.close();
        }

        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
          this.isConnected = true;
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
          }

          try {
            this.handleMessage(message);
          } catch (error) {
            console.error("Error handling WebSocket message:", error);
          }
        };

        this.ws.onclose = (_event) => {
          // WebSocket connection closed
          this.isConnected = false;
          this.emit("disconnect");
          this.attemptReconnect();
        };

        this.ws.onerror = (error) => {
          console.error("WebSocket error:", error);
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

  private handleMessage(message: any) {
    // Handle the backend's direct serialization format
    // Backend sends FullStateEvent directly as {game, playerId} without a type wrapper
    if (message.game && message.playerId && !message.type) {
      const statePayload: FullStatePayload = {
        game: message.game,
        playerId: message.playerId
      };
      this.currentPlayerId = statePayload.playerId;
      this.emit("full-state", statePayload);
      return;
    }

    // Handle the standard WebSocketMessage format
    if (message.type) {
      switch (message.type) {
        case MessageType.GAME_UPDATED: {
          const gamePayload = message.payload as GameUpdatedPayload;
          this.emit("game-updated", gamePayload.game);
          break;
        }
        case MessageType.PLAYER_CONNECTED: {
          const playerPayload = message.payload as PlayerConnectedPayload;
          this.currentPlayerId = playerPayload.playerId;
          this.emit("player-connected", playerPayload);
          break;
        }
        case MessageType.PLAYER_RECONNECTED: {
          const reconnectedPayload = message.payload as PlayerReconnectedPayload;
          this.emit("player-reconnected", reconnectedPayload);
          break;
        }
        case MessageType.PLAYER_DISCONNECTED: {
          const disconnectedPayload =
            message.payload as PlayerDisconnectedPayload;
          this.emit("player-disconnected", disconnectedPayload);
          break;
        }
        case MessageType.ERROR: {
          const errorPayload = message.payload as ErrorPayload;
          this.emit("error", errorPayload);
          break;
        }
        case MessageType.FULL_STATE: {
          const statePayload = message.payload as FullStatePayload;
          this.currentPlayerId = statePayload.playerId;
          this.emit("full-state", statePayload);
          break;
        }
        case MessageType.PRODUCTION_PHASE_STARTED: {
          const productionPayload =
            message.payload as ProductionPhaseStartedPayload;
          this.emit("production-phase-started", productionPayload);
          break;
        }
        default:
          console.warn("Unknown message type:", message.type);
      }
    } else {
      console.warn("Unknown message format:", message);
    }
  }

  sendCommand(command: any): string {
    const reqId = uuidv4();

    if (!this.isConnected || !this.ws) {
      throw new Error("WebSocket is not connected");
    }

    this.ws.send(JSON.stringify(command));

    return reqId;
  }

  playerConnect(
    playerName: string,
    gameId: string,
    playerId?: string,
  ): Promise<PlayerConnectedPayload | PlayerReconnectedPayload> {
    return new Promise((resolve, reject) => {
      // Send the connect command directly
      const command = {
        type: "player-connect",
        playerName,
        gameId,
        playerId: playerId || null
      };

      this.sendCommand(command);
      this.currentGameId = gameId;

      // Set up timeout
      const timeout = setTimeout(() => {
        this.off("player-connected", connectedHandler);
        this.off("player-reconnected", reconnectedHandler);
        this.off("full-state", fullStateHandler);
        this.off("error", errorHandler);
        reject(new Error("Timeout waiting for player connection confirmation"));
      }, 10000); // 10 second timeout

      // Handler for new connections
      const connectedHandler = (payload: PlayerConnectedPayload) => {
        if (payload.playerName === playerName) {
          clearTimeout(timeout);
          this.off("player-connected", connectedHandler);
          this.off("player-reconnected", reconnectedHandler);
          this.off("full-state", fullStateHandler);
          this.off("error", errorHandler);
          this.currentPlayerId = payload.playerId;
          resolve(payload);
        }
      };

      // Handler for reconnections
      const reconnectedHandler = (payload: PlayerReconnectedPayload) => {
        if (payload.playerName === playerName) {
          clearTimeout(timeout);
          this.off("player-connected", connectedHandler);
          this.off("player-reconnected", reconnectedHandler);
          this.off("full-state", fullStateHandler);
          this.off("error", errorHandler);
          this.currentPlayerId = payload.playerId;
          resolve(payload);
        }
      };

      // Handler for full state (which includes player connection info)
      const fullStateHandler = (payload: FullStatePayload) => {
        // Check if this is for our player by matching the gameId and that we have a playerId
        if (payload.game.id === gameId && payload.playerId) {
          clearTimeout(timeout);
          this.off("player-connected", connectedHandler);
          this.off("player-reconnected", reconnectedHandler);
          this.off("full-state", fullStateHandler);
          this.off("error", errorHandler);
          this.currentPlayerId = payload.playerId;

          // Convert FullStatePayload to PlayerConnectedPayload format for compatibility
          const connectedPayload: PlayerConnectedPayload = {
            playerId: payload.playerId,
            playerName: playerName,
            game: payload.game
          };
          resolve(connectedPayload);
        }
      };

      // Error handler
      const errorHandler = (errorPayload: ErrorPayload) => {
        clearTimeout(timeout);
        this.off("player-connected", connectedHandler);
        this.off("player-reconnected", reconnectedHandler);
        this.off("full-state", fullStateHandler);
        this.off("error", errorHandler);
        reject(new Error(errorPayload.message || "Connection failed"));
      };

      // Listen for all types of responses
      this.on("player-connected", connectedHandler);
      this.on("player-reconnected", reconnectedHandler);
      this.on("full-state", fullStateHandler);
      this.on("error", errorHandler);
    });
  }

  // Standard project actions
  sellPatents(cardCount: number): string {
    return this.sendCommand({
      type: "action.standard-project.sell-patents",
      gameId: this.currentGameId,
      playerId: this.currentPlayerId,
      cardCount
    });
  }

  launchAsteroid(): string {
    return this.sendCommand({
      type: "action.standard-project.launch-asteroid",
      gameId: this.currentGameId,
      playerId: this.currentPlayerId
    });
  }

  buildPowerPlant(): string {
    return this.sendCommand({
      type: "action.standard-project.build-power-plant",
      gameId: this.currentGameId,
      playerId: this.currentPlayerId
    });
  }

  buildAquifer(hexPosition: HexPositionDto): string {
    return this.sendCommand({
      type: "action.standard-project.build-aquifer",
      gameId: this.currentGameId,
      playerId: this.currentPlayerId,
      hexPosition
    });
  }

  plantGreenery(hexPosition: HexPositionDto): string {
    return this.sendCommand({
      type: "action.standard-project.plant-greenery",
      gameId: this.currentGameId,
      playerId: this.currentPlayerId,
      hexPosition
    });
  }

  buildCity(hexPosition: HexPositionDto): string {
    return this.sendCommand({
      type: "action.standard-project.build-city",
      gameId: this.currentGameId,
      playerId: this.currentPlayerId,
      hexPosition
    });
  }

  // Game management actions
  startGame(): string {
    return this.sendCommand({
      type: "action.game-management.start-game",
      gameId: this.currentGameId,
      playerId: this.currentPlayerId
    });
  }

  skipAction(): string {
    return this.sendCommand({
      type: "action.game-management.skip-action",
      gameId: this.currentGameId,
      playerId: this.currentPlayerId
    });
  }

  // Card actions
  playCard(cardId: string): string {
    return this.sendCommand({
      type: "action.card.play-card",
      gameId: this.currentGameId,
      playerId: this.currentPlayerId,
      cardId
    });
  }

  selectStartingCard(cardIds: string[]): string {
    return this.sendCommand({
      type: "action.card.select-starting-card",
      gameId: this.currentGameId,
      playerId: this.currentPlayerId,
      cardIds
    });
  }

  selectCards(cardIds: string[]): string {
    return this.sendCommand({
      type: "action.card.select-cards",
      gameId: this.currentGameId,
      playerId: this.currentPlayerId,
      cardIds
    });
  }

  // productionPhaseReady(): string {
  //   if (!this.currentPlayerId) {
  //     throw new Error("Cannot send production phase ready without player ID");
  //   }

  //   return this.send(MessageTypeProductionPhaseReady, {
  //     playerId: this.currentPlayerId,
  //   });
  // }

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
      this.listeners[event].forEach((callback) => callback(data));
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
