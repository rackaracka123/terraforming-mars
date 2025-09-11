import { v4 as uuidv4 } from "uuid";
import {
  ErrorPayload,
  FullStatePayload,
  GameUpdatedPayload,
  MessageType,
  MessageTypeError,
  MessageTypeFullState,
  MessageTypeGameUpdated,
  MessageTypePlayAction,
  MessageTypePlayerConnect,
  MessageTypePlayerConnected,
  MessageTypePlayerReconnected,
  MessageTypePlayerDisconnected,
  MessageTypeProductionPhaseStarted,
  PlayerConnectedPayload,
  PlayerReconnectedPayload,
  PlayerDisconnectedPayload,
  ProductionPhaseStartedPayload,
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

  private handleMessage(message: WebSocketMessage) {
    switch (message.type) {
      case MessageTypeGameUpdated: {
        const gamePayload = message.payload as GameUpdatedPayload;
        this.emit("game-updated", gamePayload.game);
        break;
      }
      case MessageTypePlayerConnected: {
        const playerPayload = message.payload as PlayerConnectedPayload;
        this.currentPlayerId = playerPayload.playerId;
        this.emit("player-connected", playerPayload);
        break;
      }
      case MessageTypePlayerReconnected: {
        const reconnectedPayload = message.payload as PlayerReconnectedPayload;
        this.emit("player-reconnected", reconnectedPayload);
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
      case MessageTypeProductionPhaseStarted: {
        const productionPayload =
          message.payload as ProductionPhaseStartedPayload;
        this.emit("production-phase-started", productionPayload);
        break;
      }
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
  ): Promise<PlayerConnectedPayload | PlayerReconnectedPayload> {
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
        this.off("player-connected", connectedHandler);
        this.off("player-reconnected", reconnectedHandler);
        this.off("error", errorHandler);
        reject(new Error("Timeout waiting for player connection confirmation"));
      }, 10000); // 10 second timeout

      // Handler for new connections
      const connectedHandler = (payload: PlayerConnectedPayload) => {
        if (payload.playerName === playerName) {
          clearTimeout(timeout);
          this.off("player-connected", connectedHandler);
          this.off("player-reconnected", reconnectedHandler);
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
          this.off("error", errorHandler);
          this.currentPlayerId = payload.playerId;
          resolve(payload);
        }
      };

      // Error handler
      const errorHandler = (errorPayload: ErrorPayload) => {
        clearTimeout(timeout);
        this.off("player-connected", connectedHandler);
        this.off("player-reconnected", reconnectedHandler);
        this.off("error", errorHandler);
        reject(new Error(errorPayload.message || "Connection failed"));
      };

      // Listen for both types of responses
      this.on("player-connected", connectedHandler);
      this.on("player-reconnected", reconnectedHandler);
      this.on("error", errorHandler);
    });
  }

  playAction(actionPayload: object): string {
    return this.send(MessageTypePlayAction, { actionRequest: actionPayload });
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
