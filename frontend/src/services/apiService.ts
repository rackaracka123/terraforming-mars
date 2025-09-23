import {
  CreateGameRequest,
  CreateGameResponse,
  GameDto,
  GameSettingsDto,
  GetGameResponse,
  ListGamesResponse,
  ListCardsResponse,
} from "../types/generated/api-types.ts";

export class ApiService {
  private baseUrl: string;

  constructor(baseUrl: string = "http://localhost:3001/api/v1") {
    this.baseUrl = baseUrl;
  }

  async createGame(settings: GameSettingsDto): Promise<GameDto> {
    try {
      const request: CreateGameRequest = {
        maxPlayers: settings.maxPlayers,
        developmentMode: settings.developmentMode,
      };

      const response = await fetch(`${this.baseUrl}/games`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(
          errorData.error || `HTTP error! status: ${response.status}`,
        );
      }

      const gameResponse: CreateGameResponse = await response.json();
      return gameResponse.game;
    } catch (error) {
      console.error("Failed to create game:", error);
      throw error;
    }
  }

  async getGame(gameId: string): Promise<GameDto | null> {
    try {
      const response = await fetch(`${this.baseUrl}/games/${gameId}`);

      if (response.status === 404) {
        return null;
      }

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(
          errorData.error || `HTTP error! status: ${response.status}`,
        );
      }

      const gameResponse: GetGameResponse = await response.json();
      return gameResponse.game;
    } catch (error) {
      console.error("Failed to get game:", error);
      throw error;
    }
  }

  async listGames(status?: string): Promise<GameDto[]> {
    try {
      const url = new URL(`${this.baseUrl}/games`);
      if (status) {
        url.searchParams.set("status", status);
      }

      const response = await fetch(url.toString());

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(
          errorData.error || `HTTP error! status: ${response.status}`,
        );
      }

      const data: ListGamesResponse = await response.json();
      return data.games || [];
    } catch (error) {
      console.error("Failed to list games:", error);
      throw error;
    }
  }

  async listCards(
    offset: number = 0,
    limit: number = 50,
  ): Promise<ListCardsResponse> {
    try {
      const url = new URL(`${this.baseUrl}/cards`);
      url.searchParams.set("offset", offset.toString());
      url.searchParams.set("limit", limit.toString());

      const response = await fetch(url.toString());

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(
          errorData.error || `HTTP error! status: ${response.status}`,
        );
      }

      const data: ListCardsResponse = await response.json();
      return data;
    } catch (error) {
      console.error("Failed to list cards:", error);
      throw error;
    }
  }
}

// Singleton instance
export const apiService = new ApiService();
