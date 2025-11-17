import { ResourcesDto } from "@/types/generated/api-types.ts";
import { CARD_STORAGE_RESOURCE_TYPES } from "../constants.ts";

/**
 * Checks if a resource is affordable by the player
 * @param resource - The resource to check
 * @param isInput - Whether this is an input (cost) or output (gain)
 * @param playerResources - The player's current resources
 * @param resourceStorage - Card-specific resource storage
 * @param cardId - ID of the card these behaviors belong to
 * @param greyOutAll - Whether to grey out all resources regardless
 * @returns true if the resource is affordable
 */
export const isResourceAffordable = (
  resource: any,
  isInput: boolean,
  playerResources?: ResourcesDto,
  resourceStorage?: { [cardId: string]: number },
  cardId?: string,
  greyOutAll: boolean = false,
): boolean => {
  if (greyOutAll) return false;
  if (!playerResources) return true;
  if (!isInput) return true;

  const resourceType = resource.resourceType || resource.type;
  const amount = resource.amount || 1;
  const target = resource.target;

  switch (resourceType) {
    case "credits":
      return playerResources.credits >= amount;
    case "steel":
      return playerResources.steel >= amount;
    case "titanium":
      return playerResources.titanium >= amount;
    case "plants":
      return playerResources.plants >= amount;
    case "energy":
      return playerResources.energy >= amount;
    case "heat":
      return playerResources.heat >= amount;
  }

  // Check card storage resources
  if (
    CARD_STORAGE_RESOURCE_TYPES.includes(resourceType) &&
    target === "self-card"
  ) {
    if (!resourceStorage || !cardId) return true;
    const currentStorage = resourceStorage[cardId] || 0;
    return currentStorage >= amount;
  }

  return true;
};
