// Award definitions for Terraforming Mars

import { Player } from './player';
import { GameState } from './game';

export interface Award {
  id: string;
  name: string;
  description: string;
  firstPlacePoints: number;
  secondPlacePoints: number;
  fundedBy?: string;
  cost: number;
  scoreFunction: (player: Player, gameState: GameState) => number;
  expansion?: string;
}

export enum AwardType {
  // Base game awards
  LANDLORD = 'landlord',
  BANKER = 'banker',
  SCIENTIST = 'scientist',
  THERMALIST = 'thermalist',
  MINER = 'miner',
  
  // Venus Next awards
  VENUPHILE = 'venuphile',
  
  // Prelude awards
  CULTIVATOR = 'cultivator',
  MAGNATE = 'magnate',
  SPACE_BARON = 'space_baron',
  ECCENTRIC = 'eccentric',
  CONTRACTOR = 'contractor',
  ENTREPRENEUR = 'entrepreneur',
  
  // Colonies awards
  COLLECTOR = 'collector',
  COLONIZER = 'colonizer',
  
  // Turmoil awards
  POLITICIAN = 'politician',
  URBANIST = 'urbanist',
  WARMONGER = 'warmonger',
  
  // The Moon awards
  LUNARCHITECT = 'lunarchitect',
  LUNAR_MAGNATE = 'lunar_magnate',
  
  // Pathfinders awards
  CURATOR = 'curator',
  ENGINEER = 'engineer',
  TOURIST = 'tourist',
  A_ZOOLOGIST = 'a_zoologist',
  
  // Underworld awards
  KINGPIN = 'kingpin',
  EDGEDANCER = 'edgedancer'
}

// Award definitions
export const AWARD_DEFINITIONS: { [key in AwardType]: Omit<Award, 'fundedBy'> } = {
  [AwardType.LANDLORD]: {
    id: 'landlord',
    name: 'Landlord',
    description: 'Most tiles in play',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    scoreFunction: (player: Player) => {
      // Count all tiles owned by player (implementation needed)
      return 0; // TODO: Implement tile counting
    }
  },
  
  [AwardType.BANKER]: {
    id: 'banker',
    name: 'Banker',
    description: 'Highest M€ production',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    scoreFunction: (player: Player) => player.production.credits
  },
  
  [AwardType.SCIENTIST]: {
    id: 'scientist',
    name: 'Scientist',
    description: 'Most science tags',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    scoreFunction: (player: Player) => {
      return player.tags.filter(tag => tag.toString() === 'science').length;
    }
  },
  
  [AwardType.THERMALIST]: {
    id: 'thermalist',
    name: 'Thermalist',
    description: 'Most heat resources',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    scoreFunction: (player: Player) => player.resources.heat
  },
  
  [AwardType.MINER]: {
    id: 'miner',
    name: 'Miner',
    description: 'Most steel and titanium resources',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    scoreFunction: (player: Player) => player.resources.steel + player.resources.titanium
  },
  
  // Venus Next
  [AwardType.VENUPHILE]: {
    id: 'venuphile',
    name: 'Venuphile',
    description: 'Most Venus tags',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'venus',
    scoreFunction: (player: Player) => {
      return player.tags.filter(tag => tag.toString() === 'venus').length;
    }
  },
  
  // Prelude
  [AwardType.CULTIVATOR]: {
    id: 'cultivator',
    name: 'Cultivator',
    description: 'Most greenery tiles',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'prelude',
    scoreFunction: (player: Player) => {
      // Count greenery tiles (implementation needed)
      return 0; // TODO: Implement greenery counting
    }
  },
  
  [AwardType.MAGNATE]: {
    id: 'magnate',
    name: 'Magnate',
    description: 'Most automated cards',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'prelude',
    scoreFunction: (player: Player) => {
      return player.playedCards.filter(card => card.definition.type.toString() === 'automated').length;
    }
  },
  
  [AwardType.SPACE_BARON]: {
    id: 'space_baron',
    name: 'Space Baron',
    description: 'Most space event cards',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'prelude',
    scoreFunction: (player: Player) => {
      return player.playedCards.filter(card => 
        card.definition.type.toString() === 'event' && 
        card.definition.tags.some(tag => tag.toString() === 'space')
      ).length;
    }
  },
  
  [AwardType.ECCENTRIC]: {
    id: 'eccentric',
    name: 'Eccentric',
    description: 'Most resources on cards',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'prelude',
    scoreFunction: (player: Player) => {
      // Count all resources on cards (implementation needed)
      return 0; // TODO: Implement resource counting on cards
    }
  },
  
  [AwardType.CONTRACTOR]: {
    id: 'contractor',
    name: 'Contractor',
    description: 'Most building tags',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'prelude',
    scoreFunction: (player: Player) => {
      return player.tags.filter(tag => tag.toString() === 'building').length;
    }
  },
  
  [AwardType.ENTREPRENEUR]: {
    id: 'entrepreneur',
    name: 'Entrepreneur',
    description: 'Most energy and heat production',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'prelude',
    scoreFunction: (player: Player) => player.production.energy + player.production.heat
  },
  
  // Colonies
  [AwardType.COLLECTOR]: {
    id: 'collector',
    name: 'Collector',
    description: 'Most cards in hand',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'colonies',
    scoreFunction: (player: Player) => player.hand.length
  },
  
  [AwardType.COLONIZER]: {
    id: 'colonizer',
    name: 'Colonizer',
    description: 'Most colonies',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'colonies',
    scoreFunction: (player: Player) => {
      // Count colonies (implementation needed)
      return 0; // TODO: Implement colony counting
    }
  },
  
  // Turmoil
  [AwardType.POLITICIAN]: {
    id: 'politician',
    name: 'Politician',
    description: 'Most party leaders and delegates',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'turmoil',
    scoreFunction: (player: Player) => {
      // Count party influence (implementation needed)
      return 0; // TODO: Implement party influence counting
    }
  },
  
  [AwardType.URBANIST]: {
    id: 'urbanist',
    name: 'Urbanist',
    description: 'Most city tiles',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'turmoil',
    scoreFunction: (player: Player) => {
      // Count city tiles (implementation needed)
      return 0; // TODO: Implement city counting
    }
  },
  
  [AwardType.WARMONGER]: {
    id: 'warmonger',
    name: 'Warmonger',
    description: 'Most cards that decrease opponent\'s production or resources',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'turmoil',
    scoreFunction: (player: Player) => {
      // Count attack cards (implementation needed)
      return 0; // TODO: Implement attack card counting
    }
  },
  
  // The Moon
  [AwardType.LUNARCHITECT]: {
    id: 'lunarchitect_award',
    name: 'Lunarchitect',
    description: 'Most Moon building tags',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'moon',
    scoreFunction: (player: Player) => {
      // Count Moon building tags (implementation needed)
      return 0; // TODO: Implement Moon building tag counting
    }
  },
  
  [AwardType.LUNAR_MAGNATE]: {
    id: 'lunar_magnate_award',
    name: 'Lunar Magnate',
    description: 'Most Moon tiles',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'moon',
    scoreFunction: (player: Player) => {
      // Count Moon tiles (implementation needed)
      return 0; // TODO: Implement Moon tile counting
    }
  },
  
  // Pathfinders
  [AwardType.CURATOR]: {
    id: 'curator',
    name: 'Curator',
    description: 'Most diverse tags',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'pathfinders',
    scoreFunction: (player: Player) => {
      const uniqueTags = new Set(player.tags.map(tag => tag.toString()));
      return uniqueTags.size;
    }
  },
  
  [AwardType.ENGINEER]: {
    id: 'engineer',
    name: 'Engineer',
    description: 'Most building tags',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'pathfinders',
    scoreFunction: (player: Player) => {
      return player.tags.filter(tag => tag.toString() === 'building').length;
    }
  },
  
  [AwardType.TOURIST]: {
    id: 'tourist',
    name: 'Tourist',
    description: 'Most Earth tags',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'pathfinders',
    scoreFunction: (player: Player) => {
      return player.tags.filter(tag => tag.toString() === 'earth').length;
    }
  },
  
  [AwardType.A_ZOOLOGIST]: {
    id: 'a_zoologist',
    name: 'A. Zoologist',
    description: 'Most animal resources',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'pathfinders',
    scoreFunction: (player: Player) => {
      // Count animal resources on cards (implementation needed)
      return 0; // TODO: Implement animal resource counting
    }
  },
  
  // Underworld
  [AwardType.KINGPIN]: {
    id: 'kingpin',
    name: 'Kingpin',
    description: 'Most corruption resources',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'underworld',
    scoreFunction: (player: Player) => {
      // Count corruption resources (implementation needed)
      return 0; // TODO: Implement corruption resource counting
    }
  },
  
  [AwardType.EDGEDANCER]: {
    id: 'edgedancer',
    name: 'Edgedancer',
    description: 'Most tiles on the edges',
    firstPlacePoints: 5,
    secondPlacePoints: 2,
    cost: 8,
    expansion: 'underworld',
    scoreFunction: (player: Player) => {
      // Count edge tiles (implementation needed)
      return 0; // TODO: Implement edge tile counting
    }
  }
};