// Player service - business logic layer
// Uses repository for data access, can be easily swapped between mock and real data

import { PlayerRepository, MockPlayerRepository, Player } from '../repositories/playerRepository';

class PlayerService {
  private repository: PlayerRepository;

  constructor(repository: PlayerRepository) {
    this.repository = repository;
  }

  getAllPlayers(): Player[] {
    return this.repository.getAllPlayers();
  }

  getPlayerById(id: string): Player | undefined {
    return this.repository.getPlayerById(id);
  }

  getPassedPlayers(): Player[] {
    return this.repository.getPassedPlayers();
  }

  getActivePlayers(): Player[] {
    return this.repository.getActivePlayers();
  }

  // Business logic methods can be added here
  getPlayersSortedByScore(): Player[] {
    return this.getAllPlayers().sort((a, b) => b.score - a.score);
  }

  getTopPlayer(): Player | undefined {
    const sorted = this.getPlayersSortedByScore();
    return sorted[0];
  }
}

// Configuration - easily switch between mock and real data
const USE_MOCK_DATA = true;

// Dependency injection - inject the appropriate repository
const repository = USE_MOCK_DATA 
  ? new MockPlayerRepository() 
  : new MockPlayerRepository(); // Would be ApiPlayerRepository() in real app

export const playerService = new PlayerService(repository);

// Convenience exports for components
export const getAllPlayers = () => playerService.getAllPlayers();
export const getActivePlayers = () => playerService.getActivePlayers();
export const getPassedPlayers = () => playerService.getPassedPlayers();

export type { Player };