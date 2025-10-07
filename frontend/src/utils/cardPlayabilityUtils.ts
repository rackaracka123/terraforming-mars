import {
  CardDto,
  GameDto,
  GamePhaseAction,
  PlayerDto,
  ResourceTriggerAuto,
  ResourceTriggerManual,
} from "../types/generated/api-types.ts";

// Cache for all cards fetched from the backend
let cardCache: Map<string, CardDto> | null = null;

// Cache for all corporations fetched from the backend
let corporationCache: CardDto[] | null = null;

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
 * Fetches all corporations from the backend and caches them
 */
export async function fetchCorporations(): Promise<CardDto[]> {
  if (corporationCache !== null) {
    return corporationCache;
  }

  try {
    const response = await fetch("/api/v1/corporations");
    if (!response.ok) {
      throw new Error(`Failed to fetch corporations: ${response.statusText}`);
    }

    const corporations: CardDto[] = await response.json();
    corporationCache = corporations || [];

    return corporationCache;
  } catch (error) {
    console.error("Failed to fetch corporations:", error);
    // Return empty array on error to prevent crashes
    return [];
  }
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
 * Checks if player can afford negative resource/production costs in card behaviors
 */
function checkBehaviorCosts(
  card: CardDto,
  player: PlayerDto,
  game: GameDto,
): UnplayableReason[] {
  const failedReasons: UnplayableReason[] = [];

  if (!card.behaviors || card.behaviors.length === 0) {
    return failedReasons;
  }

  for (const behavior of card.behaviors) {
    if (!behavior.outputs || behavior.outputs.length === 0) {
      continue;
    }

    // Skip behaviors that shouldn't be checked when playing the card
    // Only check auto-triggered behaviors without conditions
    if (behavior.triggers && behavior.triggers.length > 0) {
      const shouldSkip = behavior.triggers.some((trigger) => {
        // Skip manual triggers (actions checked separately when used)
        if (trigger.type === ResourceTriggerManual) {
          return true;
        }
        // Skip auto triggers with conditions (not yet supported)
        if (
          trigger.type === ResourceTriggerAuto &&
          trigger.condition !== undefined
        ) {
          return true;
        }
        return false;
      });

      if (shouldSkip) {
        continue;
      }
    }

    for (const output of behavior.outputs) {
      // Skip positive amounts
      if (output.amount >= 0) {
        continue;
      }

      // Only check self-player and any-player targets
      // (any-player means you can target yourself or opponents, but SOMEONE must have it)
      if (output.target !== "self-player" && output.target !== "any-player") {
        continue;
      }

      const absAmount = Math.abs(output.amount);
      const resourceType = output.type.toLowerCase();

      // Helper to get production value
      const getProduction = (prod: any, baseResource: string): number => {
        switch (baseResource) {
          case "credits":
          case "megacredits":
            return prod?.credits || 0;
          case "steel":
            return prod?.steel || 0;
          case "titanium":
            return prod?.titanium || 0;
          case "plants":
          case "plant":
            return prod?.plants || 0;
          case "energy":
          case "power":
            return prod?.energy || 0;
          case "heat":
            return prod?.heat || 0;
          default:
            return 0;
        }
      };

      // Helper to get resource amount
      const getResourceAmount = (res: any, resType: string): number => {
        switch (resType) {
          case "credits":
          case "megacredits":
            return res?.credits || 0;
          case "steel":
            return res?.steel || 0;
          case "titanium":
            return res?.titanium || 0;
          case "plants":
          case "plant":
            return res?.plants || 0;
          case "energy":
          case "power":
            return res?.energy || 0;
          case "heat":
            return res?.heat || 0;
          default:
            return 0;
        }
      };

      // Check if it's a production type (ends with -production)
      if (resourceType.endsWith("-production")) {
        const baseResource = resourceType.replace("-production", "");

        if (output.target === "any-player") {
          // Check if ANY player has enough
          const currentProduction = getProduction(
            player.production,
            baseResource,
          );
          let anyoneHasEnough = currentProduction >= absAmount;

          if (!anyoneHasEnough && game.otherPlayers) {
            for (const other of game.otherPlayers) {
              if (getProduction(other.production, baseResource) >= absAmount) {
                anyoneHasEnough = true;
                break;
              }
            }
          }

          if (!anyoneHasEnough) {
            failedReasons.push({
              type: "production",
              requirement: { resource: resourceType, amount: absAmount },
              message: `${absAmount}`,
              currentValue: currentProduction,
              requiredValue: absAmount,
            });
          }
        } else {
          // self-player only
          const currentProduction = getProduction(
            player.production,
            baseResource,
          );

          if (currentProduction < absAmount) {
            failedReasons.push({
              type: "production",
              requirement: { resource: resourceType, amount: absAmount },
              message: `${absAmount}`,
              currentValue: currentProduction,
              requiredValue: absAmount,
            });
          }
        }
      } else {
        // Regular resource
        if (output.target === "any-player") {
          const currentAmount = getResourceAmount(
            player.resources,
            resourceType,
          );
          let anyoneHasEnough = currentAmount >= absAmount;

          if (!anyoneHasEnough && game.otherPlayers) {
            for (const other of game.otherPlayers) {
              if (
                getResourceAmount(other.resources, resourceType) >= absAmount
              ) {
                anyoneHasEnough = true;
                break;
              }
            }
          }

          // Skip special resource types stored on cards
          const isStandardResource = [
            "credits",
            "megacredits",
            "steel",
            "titanium",
            "plants",
            "plant",
            "energy",
            "power",
            "heat",
          ].includes(resourceType);
          if (!isStandardResource) {
            continue;
          }

          if (!anyoneHasEnough) {
            failedReasons.push({
              type: "resource",
              requirement: { resource: resourceType, amount: absAmount },
              message: `${absAmount}`,
              currentValue: currentAmount,
              requiredValue: absAmount,
            });
          }
        } else {
          // self-player only
          const currentAmount = getResourceAmount(
            player.resources,
            resourceType,
          );

          // Skip special resource types
          const isStandardResource = [
            "credits",
            "megacredits",
            "steel",
            "titanium",
            "plants",
            "plant",
            "energy",
            "power",
            "heat",
          ].includes(resourceType);
          if (!isStandardResource) {
            continue;
          }

          if (currentAmount < absAmount) {
            failedReasons.push({
              type: "resource",
              requirement: { resource: resourceType, amount: absAmount },
              message: `${absAmount}`,
              currentValue: currentAmount,
              requiredValue: absAmount,
            });
          }
        }
      }
    }
  }

  return failedReasons;
}

/**
 * Deduplicates requirements - keeps only the highest amount for each resource type
 */
function deduplicateRequirements(
  requirements: UnplayableReason[],
): UnplayableReason[] {
  const deduplicatedRequirements: UnplayableReason[] = [];
  const resourceMap = new Map<string, UnplayableReason>();

  for (const reason of requirements) {
    // Create a unique key for this requirement type + resource
    const key = `${reason.type}:${reason.requirement?.resource || ""}`;
    const existing = resourceMap.get(key);

    if (!existing) {
      // First time seeing this resource requirement
      resourceMap.set(key, reason);
    } else {
      // Already have a requirement for this resource - keep the higher one
      const existingAmount =
        typeof existing.requiredValue === "number" ? existing.requiredValue : 0;
      const currentAmount =
        typeof reason.requiredValue === "number" ? reason.requiredValue : 0;

      if (currentAmount > existingAmount) {
        resourceMap.set(key, reason);
      }
    }
  }

  // Convert map back to array
  resourceMap.forEach((reason) => deduplicatedRequirements.push(reason));

  return deduplicatedRequirements;
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
            message: `${min}°C`,
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
            message: `Max ${max}°C`,
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
            message: `${min}%`,
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
            message: `Max ${max}%`,
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
            message: `${min}`,
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
            message: `Max ${max}`,
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
              message: `${requiredCount}`,
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
        const prod = player.production;

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
              message: `${requiredProduction}`,
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

  // First check: Pending tile selection (highest priority - blocks everything)
  if (player.pendingTileSelection) {
    return {
      playable: false,
      reason: {
        type: "phase",
        requirement: { pendingTileSelection: true },
        message: "Pending tile selection",
        currentValue: "tile_selection_pending",
        requiredValue: "no_pending_tiles",
      },
    };
  }

  // Second check: Game phase (must be action phase to play cards)
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

  // Third check: Cost (highest priority after phase)
  if (player.resources.credits < card.cost) {
    failedRequirements.push({
      type: "cost",
      requirement: { cost: card.cost },
      message: `${card.cost}`,
      currentValue: player.resources.credits,
      requiredValue: card.cost,
    });
  }

  // Fourth check: Behavior costs (negative resource/production costs)
  const behaviorCostFailures = checkBehaviorCosts(card, player, game);
  failedRequirements.push(...behaviorCostFailures);

  // Check all requirements
  if (card.requirements && card.requirements.length > 0) {
    for (const requirement of card.requirements) {
      const { met, reason } = await checkRequirement(requirement, game, player);
      if (!met && reason) {
        failedRequirements.push(reason);
      }
    }
  }

  // Deduplicate requirements - keep only the highest amount for each resource type
  const deduplicatedRequirements = deduplicateRequirements(failedRequirements);

  // If there are failed requirements, return them
  if (deduplicatedRequirements.length > 0) {
    // If only one failed requirement, return it directly
    if (deduplicatedRequirements.length === 1) {
      return { playable: false, reason: deduplicatedRequirements[0] };
    }

    // Multiple failed requirements - create a combined reason
    const message = "Multiple requirements not met:";

    return {
      playable: false,
      reason: {
        type: "multiple",
        requirement: { multiple: true },
        message: message,
        failedRequirements: deduplicatedRequirements,
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
