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
  MessageTypePlayerReconnect,
  MessageTypePlayerReconnected,
  MessageTypePlayerDisconnected,
  MessageTypeProductionPhaseStarted,
  MessageTypeProductionPhaseReady,
  PlayerConnectedPayload,
  PlayerReconnectedPayload,
  PlayerDisconnectedPayload,
  ProductionPhaseStartedPayload,
  ProductionPhaseReadyPayload,
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
  private isPageReload = false;

  constructor(url: string = "ws://localhost:3001/ws") {
    this.url = url;

    // Detect if this is a page reload by checking performance navigation timing
    if (typeof window !== "undefined" && window.performance) {
      const navEntries = window.performance.getEntriesByType(
        "navigation",
      ) as PerformanceNavigationTiming[];
      if (navEntries.length > 0) {
        this.isPageReload = navEntries[0].type === "reload";
      }
    }
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        // On page reload, always force a fresh connection
        if (this.isPageReload) {
          console.log(
            "🔄 Page reload detected, forcing fresh WebSocket connection",
          );
          this.forceDisconnect();
        }

        // If already connected, resolve immediately
        if (
          this.isConnected &&
          this.ws &&
          this.ws.readyState === WebSocket.OPEN &&
          !this.isPageReload
        ) {
          console.log("🔗 WebSocket already connected, reusing connection");
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

          if (this.isPageReload) {
            console.log(
              "🔄 WebSocket connected after page reload - fresh connection established",
            );
            this.isPageReload = false; // Reset flag after successful connection
          } else {
            console.log(
              "🔗 WebSocket connected - normal connection established",
            );
          }

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
  ): Promise<PlayerConnectedPayload> {
    return new Promise((resolve, reject) => {
      this.send(MessageTypePlayerConnect, { playerName, gameId }, gameId);
      this.currentGameId = gameId;

      // Set up one-time listener for player-connected response
      const timeout = setTimeout(() => {
        this.off("player-connected", responseHandler);
        reject(new Error("Timeout waiting for player connection confirmation"));
      }, 10000); // 10 second timeout

      const responseHandler = (payload: PlayerConnectedPayload) => {
        clearTimeout(timeout);
        this.off("player-connected", responseHandler);
        resolve(payload);
      };

      this.on("player-connected", responseHandler);
    });
  }

  playerReconnect(
    playerName: string,
    gameId: string,
  ): Promise<PlayerReconnectedPayload> {
    return new Promise((resolve, reject) => {
      console.log("🔄 Sending player-reconnect message", {
        playerName,
        gameId,
      });
      this.send(MessageTypePlayerReconnect, { playerName, gameId }, gameId);
      this.currentGameId = gameId;

      // Set up one-time listener for player-reconnected response
      const timeout = setTimeout(() => {
        this.off("player-reconnected", responseHandler);
        this.off("error", errorHandler);
        reject(
          new Error("Timeout waiting for player reconnection confirmation"),
        );
      }, 10000); // 10 second timeout

      const responseHandler = (payload: PlayerReconnectedPayload) => {
        console.log("📨 Received player-reconnected message", {
          payloadPlayerName: payload.playerName,
          expectedPlayerName: playerName,
          playerId: payload.playerId,
        });

        // Only resolve if this is the reconnection for the current player
        if (payload.playerName === playerName) {
          console.log(
            "✅ Player reconnection confirmed - matching player name",
          );
          clearTimeout(timeout);
          this.off("player-reconnected", responseHandler);
          this.off("error", errorHandler);
          this.currentPlayerId = payload.playerId;
          resolve(payload);
        } else {
          console.log(
            "⚠️ Player reconnection received for different player, ignoring",
          );
        }
      };

      const errorHandler = (errorPayload: ErrorPayload) => {
        clearTimeout(timeout);
        this.off("player-reconnected", responseHandler);
        this.off("error", errorHandler);
        reject(new Error(errorPayload.message || "Reconnection failed"));
      };

      this.on("player-reconnected", responseHandler);
      this.on("error", errorHandler);
    });
  }

  playAction(actionPayload: object): string {
    return this.send(MessageTypePlayAction, { actionRequest: actionPayload });
  }

  productionPhaseReady(): string {
    if (!this.currentPlayerId) {
      throw new Error(
        "Cannot send production-phase-ready: no current player ID",
      );
    }

    const payload: ProductionPhaseReadyPayload = {
      playerId: this.currentPlayerId,
    };

    console.log("📦 Sending production-phase-ready message", {
      playerId: this.currentPlayerId,
      gameId: this.currentGameId,
    });

    return this.send(MessageTypeProductionPhaseReady, payload);
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

  forceDisconnect() {
    console.log("🚨 Force disconnecting WebSocket connection");
    this.isConnected = false;
    this.currentGameId = null;
    this.currentPlayerId = null;
    this.reconnectAttempts = 0;

    if (this.ws) {
      // Remove all event listeners to prevent them from firing
      this.ws.onopen = null;
      this.ws.onmessage = null;
      this.ws.onclose = null;
      this.ws.onerror = null;

      if (
        this.ws.readyState === WebSocket.OPEN ||
        this.ws.readyState === WebSocket.CONNECTING
      ) {
        this.ws.close();
      }
      this.ws = null;
    }
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
