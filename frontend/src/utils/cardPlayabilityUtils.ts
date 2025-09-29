import {
  CardDto,
  GameDto,
  GamePhaseAction,
  PlayerDto,
} from "../types/generated/api-types.ts";

// Cache for all cards fetched from the backend
let cardCache: Map<string, CardDto> | null = null;

/**
 * Fetches all cards from the backend and caches them
 */
export async function fetchAllCards(): Promise<Map<string, CardDto>> {
  if (cardCache !== null) {
    return cardCache;
  }

  try {
    const response = await fetch("/api/v1/cards?limit=500");
    if (!response.ok) {
      throw new Error(`Failed to fetch cards: ${response.statusText}`);
    }

    const data = await response.json();
    const cards = data.cards || [];

    cardCache = new Map();
    cards.forEach((card: CardDto) => {
      cardCache!.set(card.id, card);
    });

    return cardCache;
  } catch (error) {
    console.error("Failed to fetch cards:", error);
    // Return empty map on error to prevent crashes
    return new Map();
  }
}

/**
 * Gets a card by its ID from the cache
 */
export async function getCardById(cardId: string): Promise<CardDto | null> {
  const cards = await fetchAllCards();
  return cards.get(cardId) || null;
}

/**
 * Reasons why a card might be unplayable
 */
export interface UnplayableReason {
  type:
    | "cost"
    | "global-param"
    | "tag"
    | "production"
    | "resource"
    | "phase"
    | "multiple";
  requirement: any;
  message: string;
  currentValue?: number | string;
  requiredValue?: number | string;
  requirementType?: string; // e.g., "min", "max"
  failedRequirements?: UnplayableReason[]; // For multiple failed requirements
}

/**
 * Counts tags from a player's played cards
 */
async function countPlayerTags(
  player: PlayerDto,
  tagType?: string,
): Promise<number> {
  if (!player.playedCards || player.playedCards.length === 0) {
    return 0;
  }

  const cards = await fetchAllCards();
  let count = 0;

  for (const cardId of player.playedCards) {
    const card = cards.get(cardId);
    if (card && card.tags) {
      if (tagType) {
        count += card.tags.filter(
          (tag) => tag.toLowerCase() === tagType.toLowerCase(),
        ).length;
      } else {
        count += card.tags.length;
      }
    }
  }

  // Also count corporation tags if the player has a corporation
  // TODO: Add corporation tag counting when corporation data is available

  return count;
}

/**
 * Checks if a specific requirement is met
 */
async function checkRequirement(
  requirement: any,
  game: GameDto,
  player: PlayerDto,
): Promise<{ met: boolean; reason?: UnplayableReason }> {
  const { type, min, max, tag, resource } = requirement;

  switch (type) {
    case "temperature": {
      const currentTemp = game.globalParameters.temperature;
      if (min !== undefined && min !== null && currentTemp < min) {
        return {
          met: false,
          reason: {
            type: "global-param",
            requirement,
            message: `Temperature must be at least ${min}°C (current: ${currentTemp}°C)`,
            currentValue: currentTemp,
            requiredValue: min,
            requirementType: "min",
          },
        };
      }
      if (max !== undefined && max !== null && currentTemp > max) {
        return {
          met: false,
          reason: {
            type: "global-param",
            requirement,
            message: `Temperature must be at most ${max}°C (current: ${currentTemp}°C)`,
            currentValue: currentTemp,
            requiredValue: max,
            requirementType: "max",
          },
        };
      }
      break;
    }

    case "oxygen": {
      const currentOxygen = game.globalParameters.oxygen;
      if (min !== undefined && min !== null && currentOxygen < min) {
        return {
          met: false,
          reason: {
            type: "global-param",
            requirement,
            message: `Oxygen must be at least ${min}% (current: ${currentOxygen}%)`,
            currentValue: currentOxygen,
            requiredValue: min,
            requirementType: "min",
          },
        };
      }
      if (max !== undefined && max !== null && currentOxygen > max) {
        return {
          met: false,
          reason: {
            type: "global-param",
            requirement,
            message: `Oxygen must be at most ${max}% (current: ${currentOxygen}%)`,
            currentValue: currentOxygen,
            requiredValue: max,
            requirementType: "max",
          },
        };
      }
      break;
    }

    case "oceans": {
      const currentOceans = game.globalParameters.oceans;
      if (min !== undefined && min !== null && currentOceans < min) {
        return {
          met: false,
          reason: {
            type: "global-param",
            requirement,
            message: `Must have at least ${min} ocean${min !== 1 ? "s" : ""} (current: ${currentOceans})`,
            currentValue: currentOceans,
            requiredValue: min,
            requirementType: "min",
          },
        };
      }
      if (max !== undefined && max !== null && currentOceans > max) {
        return {
          met: false,
          reason: {
            type: "global-param",
            requirement,
            message: `Must have at most ${max} ocean${max !== 1 ? "s" : ""} (current: ${currentOceans})`,
            currentValue: currentOceans,
            requiredValue: max,
            requirementType: "max",
          },
        };
      }
      break;
    }

    case "tags": {
      if (tag) {
        const tagCount = await countPlayerTags(player, tag);
        const requiredCount = min || 1;
        if (tagCount < requiredCount) {
          return {
            met: false,
            reason: {
              type: "tag",
              requirement,
              message: `Need ${requiredCount} ${tag} tag${requiredCount !== 1 ? "s" : ""} (have ${tagCount})`,
              currentValue: tagCount,
              requiredValue: requiredCount,
              requirementType: "min",
            },
          };
        }
      }
      break;
    }

    case "production": {
      if (resource) {
        let currentProduction = 0;
        const prod = player.resourceProduction;

        switch (resource.toLowerCase()) {
          case "credits":
          case "megacredits":
            currentProduction = prod?.credits || 0;
            break;
          case "steel":
            currentProduction = prod?.steel || 0;
            break;
          case "titanium":
            currentProduction = prod?.titanium || 0;
            break;
          case "plants":
          case "plant":
            currentProduction = prod?.plants || 0;
            break;
          case "energy":
          case "power":
            currentProduction = prod?.energy || 0;
            break;
          case "heat":
            currentProduction = prod?.heat || 0;
            break;
        }

        const requiredProduction = min || 1;
        if (currentProduction < requiredProduction) {
          return {
            met: false,
            reason: {
              type: "production",
              requirement,
              message: `Need ${requiredProduction} ${resource} production (have ${currentProduction})`,
              currentValue: currentProduction,
              requiredValue: requiredProduction,
              requirementType: "min",
            },
          };
        }
      }
      break;
    }

    // TODO: Add more requirement types as needed (cities, greeneries, TR, etc.)
  }

  return { met: true };
}

/**
 * Checks if a card is playable by the current player
 */
export async function checkCardPlayability(
  card: CardDto,
  game: GameDto,
  player: PlayerDto,
): Promise<{ playable: boolean; reason?: UnplayableReason }> {
  const failedRequirements: UnplayableReason[] = [];

  // First check: Game phase (must be action phase to play cards)
  if (game.currentPhase !== GamePhaseAction) {
    return {
      playable: false,
      reason: {
        type: "phase",
        requirement: { phase: game.currentPhase },
        message: "Cards can only be played during the action phase",
        currentValue: game.currentPhase,
        requiredValue: GamePhaseAction,
      },
    };
  }

  // Second check: Cost (highest priority after phase)
  if (player.resources.credits < card.cost) {
    failedRequirements.push({
      type: "cost",
      requirement: { cost: card.cost },
      message: `Not enough credits (need ${card.cost}, have ${player.resources.credits})`,
      currentValue: player.resources.credits,
      requiredValue: card.cost,
    });
  }

  // Check all requirements
  if (card.requirements && card.requirements.length > 0) {
    for (const requirement of card.requirements) {
      const { met, reason } = await checkRequirement(requirement, game, player);
      if (!met && reason) {
        failedRequirements.push(reason);
      }
    }
  }

  // If there are failed requirements, return them
  if (failedRequirements.length > 0) {
    // If only one failed requirement, return it directly
    if (failedRequirements.length === 1) {
      return { playable: false, reason: failedRequirements[0] };
    }

    // Multiple failed requirements - create a combined reason
    const message = "Multiple requirements not met:";

    return {
      playable: false,
      reason: {
        type: "multiple",
        requirement: { multiple: true },
        message: message,
        failedRequirements: failedRequirements,
      },
    };
  }

  return { playable: true };
}

/**
 * Get the reason why a card is unplayable (if any)
 */
export async function getUnplayableReason(
  card: CardDto,
  game: GameDto,
  player: PlayerDto,
): Promise<UnplayableReason | null> {
  const { playable, reason } = await checkCardPlayability(card, game, player);
  return playable ? null : reason || null;
}

/**
 * Checks multiple cards and returns a map of card ID to playability
 */
export async function checkMultipleCardsPlayability(
  cards: CardDto[],
  game: GameDto,
  player: PlayerDto,
): Promise<Map<string, { playable: boolean; reason?: UnplayableReason }>> {
  const results = new Map();

  for (const card of cards) {
    const result = await checkCardPlayability(card, game, player);
    results.set(card.id, result);
  }

  return results;
}
