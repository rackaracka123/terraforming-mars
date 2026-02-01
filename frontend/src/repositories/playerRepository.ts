// Player repository - handles data storage and retrieval
// This layer abstracts the data source (mock, API, etc.)

interface Player {
  id: string;
  name: string;
  score: number;
  passed: boolean;
  corporation: string;
  terraformRating: number;
}

// Mock data storage - in a real app this would be API calls, database queries, etc.
const mockPlayersData: Player[] = [
  {
    id: "1",
    name: "Alice Chen",
    score: 76,
    passed: true,
    corporation: "mars-direct",
    terraformRating: 35,
  },
  {
    id: "2",
    name: "Bob Martinez",
    score: 76,
    passed: true,
    corporation: "habitat-marte",
    terraformRating: 34,
  },
  {
    id: "3",
    name: "Carol Kim",
    score: 28,
    passed: false,
    corporation: "aurorai",
    terraformRating: 28,
  },
  {
    id: "4",
    name: "David Singh",
    score: 24,
    passed: false,
    corporation: "bio-sol",
    terraformRating: 24,
  },
  {
    id: "5",
    name: "Emma Wilson",
    score: 27,
    passed: false,
    corporation: "chimera",
    terraformRating: 27,
  },
  {
    id: "6",
    name: "Frank Lee",
    score: 19,
    passed: true,
    corporation: "odyssey",
    terraformRating: 19,
  },
];

// Repository interface - defines the contract for data access
export interface PlayerRepository {
  getAllPlayers(): Player[];
  getPlayerById(id: string): Player | undefined;
  getPassedPlayers(): Player[];
  getActivePlayers(): Player[];
}

// Mock implementation of the repository
export class MockPlayerRepository implements PlayerRepository {
  getAllPlayers(): Player[] {
    return [...mockPlayersData]; // Return copy to prevent mutation
  }

  getPlayerById(id: string): Player | undefined {
    return mockPlayersData.find((player) => player.id === id);
  }

  getPassedPlayers(): Player[] {
    return mockPlayersData.filter((player) => player.passed);
  }

  getActivePlayers(): Player[] {
    return mockPlayersData.filter((player) => !player.passed);
  }
}

// Future: RealPlayerRepository for actual API calls
// export class ApiPlayerRepository implements PlayerRepository {
//   async getAllPlayers(): Promise<Player[]> {
//     const response = await fetch('/api/players');
//     return response.json();
//   }
//   // ... other methods
// }

export type { Player };
