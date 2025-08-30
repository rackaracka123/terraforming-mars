// Milestone definitions for Terraforming Mars

import { Player } from './player';
import { GameState } from './game';

export interface Milestone {
  id: string;
  name: string;
  description: string;
  victoryPoints: number;
  claimedBy?: string;
  cost: number;
  checkFunction: (player: Player, gameState: GameState) => boolean;
  expansion?: string;
}

export enum MilestoneType {
  // Base game milestones
  TERRAFORMER = 'terraformer',
  MAYOR = 'mayor',
  GARDENER = 'gardener',
  BUILDER = 'builder',
  PLANNER = 'planner',
  
  // Venus Next milestones
  HOVERLORD = 'hoverlord',
  VENUPHILE = 'venuphile',
  
  // Prelude milestones
  ENERGIZER = 'energizer',
  RIM_SETTLER = 'rim_settler',
  
  // Colonies milestones
  COSMIC_SETTLER = 'cosmic_settler',
  BENEFACTOR = 'benefactor',
  
  // Turmoil milestones
  CELEBRITY = 'celebrity',
  INDUSTRIALIST = 'industrialist',
  DESERT_SETTLER = 'desert_settler',
  ESTATE_DEALER = 'estate_dealer',
  
  // The Moon milestones
  LUNAR_MAGNATE = 'lunar_magnate',
  FULL_MOON = 'full_moon',
  LUNARCHITECT = 'lunarchitect',
  
  // Pathfinders milestones
  MARTIAN = 'martian',
  TROPICALIST = 'tropicalist',
  POLAR_EXPLORER = 'polar_explorer',
  
  // Underworld milestones
  EXCAVATOR = 'excavator',
  TUNNELER = 'tunneler'
}

// Milestone definitions
export const MILESTONE_DEFINITIONS: { [key in MilestoneType]: Omit<Milestone, 'claimedBy'> } = {
  [MilestoneType.TERRAFORMER]: {
    id: 'terraformer',
    name: 'Terraformer',
    description: 'Have a terraform rating of at least 35',
    victoryPoints: 5,
    cost: 8,
    checkFunction: (player: Player) => player.terraformRating >= 35
  },
  
  [MilestoneType.MAYOR]: {
    id: 'mayor',
    name: 'Mayor',
    description: 'Own at least 3 cities',
    victoryPoints: 5,
    cost: 8,
    checkFunction: (player: Player) => {
      // Count city tiles owned by player (implementation needed)
      return false; // TODO: Implement city counting
    }
  },
  
  [MilestoneType.GARDENER]: {
    id: 'gardener',
    name: 'Gardener',
    description: 'Own at least 3 greenery tiles',
    victoryPoints: 5,
    cost: 8,
    checkFunction: (player: Player) => {
      // Count greenery tiles owned by player (implementation needed)
      return false; // TODO: Implement greenery counting
    }
  },
  
  [MilestoneType.BUILDER]: {
    id: 'builder',
    name: 'Builder',
    description: 'Have at least 8 building tags',
    victoryPoints: 5,
    cost: 8,
    checkFunction: (player: Player) => {
      return player.tags.filter(tag => tag.toString() === 'building').length >= 8;
    }
  },
  
  [MilestoneType.PLANNER]: {
    id: 'planner',
    name: 'Planner',
    description: 'Have at least 16 cards in your hand',
    victoryPoints: 5,
    cost: 8,
    checkFunction: (player: Player) => player.hand.length >= 16
  },
  
  // Venus Next
  [MilestoneType.HOVERLORD]: {
    id: 'hoverlord',
    name: 'Hoverlord',
    description: 'Have at least 7 floater resources',
    victoryPoints: 5,
    cost: 8,
    expansion: 'venus',
    checkFunction: (player: Player) => {
      // Count floater resources on cards (implementation needed)
      return false; // TODO: Implement floater counting
    }
  },
  
  [MilestoneType.VENUPHILE]: {
    id: 'venuphile',
    name: 'Venuphile',
    description: 'Have at least 3 Venus tags',
    victoryPoints: 5,
    cost: 8,
    expansion: 'venus',
    checkFunction: (player: Player) => {
      return player.tags.filter(tag => tag.toString() === 'venus').length >= 3;
    }
  },
  
  // Prelude
  [MilestoneType.ENERGIZER]: {
    id: 'energizer',
    name: 'Energizer',
    description: 'Have at least 6 energy production',
    victoryPoints: 5,
    cost: 8,
    expansion: 'prelude',
    checkFunction: (player: Player) => player.production.energy >= 6
  },
  
  [MilestoneType.RIM_SETTLER]: {
    id: 'rim_settler',
    name: 'Rim Settler',
    description: 'Have at least 3 Jovian tags',
    victoryPoints: 5,
    cost: 8,
    expansion: 'prelude',
    checkFunction: (player: Player) => {
      return player.tags.filter(tag => tag.toString() === 'jovian').length >= 3;
    }
  },
  
  // Colonies
  [MilestoneType.COSMIC_SETTLER]: {
    id: 'cosmic_settler',
    name: 'Cosmic Settler',
    description: 'Have at least 4 colonies',
    victoryPoints: 5,
    cost: 8,
    expansion: 'colonies',
    checkFunction: (player: Player) => {
      // Count colonies owned by player (implementation needed)
      return false; // TODO: Implement colony counting
    }
  },
  
  [MilestoneType.BENEFACTOR]: {
    id: 'benefactor',
    name: 'Benefactor',
    description: 'Have at least 20 TR from cards and awards',
    victoryPoints: 5,
    cost: 8,
    expansion: 'colonies',
    checkFunction: (player: Player) => {
      // Calculate TR from cards and awards (implementation needed)
      return false; // TODO: Implement TR from cards calculation
    }
  },
  
  // Turmoil
  [MilestoneType.CELEBRITY]: {
    id: 'celebrity',
    name: 'Celebrity',
    description: 'Have at least 5 cards with a cost of at least 20 MC',
    victoryPoints: 5,
    cost: 8,
    expansion: 'turmoil',
    checkFunction: (player: Player) => {
      return player.playedCards.filter(card => card.definition.cost >= 20).length >= 5;
    }
  },
  
  [MilestoneType.INDUSTRIALIST]: {
    id: 'industrialist',
    name: 'Industrialist',
    description: 'Have at least 6 steel/titanium production',
    victoryPoints: 5,
    cost: 8,
    expansion: 'turmoil',
    checkFunction: (player: Player) => {
      return (player.production.steel + player.production.titanium) >= 6;
    }
  },
  
  [MilestoneType.DESERT_SETTLER]: {
    id: 'desert_settler',
    name: 'Desert Settler',
    description: 'Have at least 3 tiles south of the equator',
    victoryPoints: 5,
    cost: 8,
    expansion: 'turmoil',
    checkFunction: (player: Player) => {
      // Count tiles south of equator (implementation needed)
      return false; // TODO: Implement equator tile counting
    }
  },
  
  [MilestoneType.ESTATE_DEALER]: {
    id: 'estate_dealer',
    name: 'Estate Dealer',
    description: 'Own at least 4 tiles adjacent to ocean tiles',
    victoryPoints: 5,
    cost: 8,
    expansion: 'turmoil',
    checkFunction: (player: Player) => {
      // Count tiles adjacent to oceans (implementation needed)
      return false; // TODO: Implement ocean adjacency counting
    }
  },
  
  // The Moon
  [MilestoneType.LUNAR_MAGNATE]: {
    id: 'lunar_magnate',
    name: 'Lunar Magnate',
    description: 'Have at least 3 tiles on The Moon',
    victoryPoints: 5,
    cost: 8,
    expansion: 'moon',
    checkFunction: (player: Player) => {
      // Count moon tiles (implementation needed)
      return false; // TODO: Implement moon tile counting
    }
  },
  
  [MilestoneType.FULL_MOON]: {
    id: 'full_moon',
    name: 'Full Moon',
    description: 'Have at least 5 Moon tags',
    victoryPoints: 5,
    cost: 8,
    expansion: 'moon',
    checkFunction: (player: Player) => {
      return player.tags.filter(tag => tag.toString() === 'moon').length >= 5;
    }
  },
  
  [MilestoneType.LUNARCHITECT]: {
    id: 'lunarchitect',
    name: 'Lunarchitect',
    description: 'Have at least 6 Moon colony rate',
    victoryPoints: 5,
    cost: 8,
    expansion: 'moon',
    checkFunction: (player: Player) => {
      // Check moon colony rate (implementation needed)
      return false; // TODO: Implement moon colony rate check
    }
  },
  
  // Pathfinders
  [MilestoneType.MARTIAN]: {
    id: 'martian',
    name: 'Martian',
    description: 'Have at least 5 Mars tags',
    victoryPoints: 5,
    cost: 8,
    expansion: 'pathfinders',
    checkFunction: (player: Player) => {
      return player.tags.filter(tag => tag.toString() === 'mars').length >= 5;
    }
  },
  
  [MilestoneType.TROPICALIST]: {
    id: 'tropicalist',
    name: 'Tropicalist',
    description: 'Have at least 3 tiles in the two bottom rows',
    victoryPoints: 5,
    cost: 8,
    expansion: 'pathfinders',
    checkFunction: (player: Player) => {
      // Count tiles in bottom rows (implementation needed)
      return false; // TODO: Implement bottom row tile counting
    }
  },
  
  [MilestoneType.POLAR_EXPLORER]: {
    id: 'polar_explorer',
    name: 'Polar Explorer',
    description: 'Have at least 3 tiles in the two top rows',
    victoryPoints: 5,
    cost: 8,
    expansion: 'pathfinders',
    checkFunction: (player: Player) => {
      // Count tiles in top rows (implementation needed)
      return false; // TODO: Implement top row tile counting
    }
  },
  
  // Underworld
  [MilestoneType.EXCAVATOR]: {
    id: 'excavator',
    name: 'Excavator',
    description: 'Have at least 4 cards with underground resources',
    victoryPoints: 5,
    cost: 8,
    expansion: 'underworld',
    checkFunction: (player: Player) => {
      // Count cards with underground resources (implementation needed)
      return false; // TODO: Implement underground resource counting
    }
  },
  
  [MilestoneType.TUNNELER]: {
    id: 'tunneler',
    name: 'Tunneler',
    description: 'Have at least 2 tiles under Mars',
    victoryPoints: 5,
    cost: 8,
    expansion: 'underworld',
    checkFunction: (player: Player) => {
      // Count underground tiles (implementation needed)
      return false; // TODO: Implement underground tile counting
    }
  }
};