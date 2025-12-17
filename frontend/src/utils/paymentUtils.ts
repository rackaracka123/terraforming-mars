import {
  CardDto,
  CardPaymentDto,
  PaymentConstantsDto,
  PaymentSubstituteDto,
  ResourcesDto,
  TagBuilding,
  TagSpace,
} from "../types/generated/api-types.ts";

/**
 * Calculates the total MC value of a payment
 * This is for UI display only - backend validates actual payment
 */
export function calculatePaymentValue(
  payment: CardPaymentDto,
  constants: PaymentConstantsDto,
  playerSubstitutes?: PaymentSubstituteDto[],
): number {
  let total =
    payment.credits +
    payment.steel * constants.steelValue +
    payment.titanium * constants.titaniumValue;

  // Add value from payment substitutes
  if (payment.substitutes && playerSubstitutes) {
    for (const [resourceType, amount] of Object.entries(payment.substitutes)) {
      const substitute = playerSubstitutes.find((sub) => sub.resourceType === resourceType);
      if (substitute) {
        total += amount * substitute.conversionRate;
      }
    }
  }

  return total;
}

/**
 * Creates a default all-credits payment for a card cost
 * Backend will validate if player can afford this
 */
export function createDefaultPayment(cardCost: number): CardPaymentDto {
  return {
    credits: cardCost,
    steel: 0,
    titanium: 0,
    substitutes: undefined,
  };
}

/**
 * Determines if the payment modal should be shown
 * Show if:
 * 1. Card can use steel/titanium AND player has them, OR
 * 2. Player has any payment substitutes
 */
export function shouldShowPaymentModal(
  card: CardDto,
  playerResources: ResourcesDto,
  playerSubstitutes?: PaymentSubstituteDto[],
): boolean {
  // If card costs 0, no payment needed
  if (card.cost === 0) {
    return false;
  }

  // Check if card can use steel (has building tag) and player has steel
  const canUseSteel = (card.tags?.includes(TagBuilding) ?? false) && playerResources.steel > 0;

  // Check if card can use titanium (has space tag) and player has titanium
  const canUseTitanium = (card.tags?.includes(TagSpace) ?? false) && playerResources.titanium > 0;

  // Check if player has any payment substitutes with available resources
  const hasUsableSubstitutes =
    playerSubstitutes &&
    playerSubstitutes.some((sub) => {
      const resourceType = sub.resourceType;
      // Check if player has this resource available
      switch (resourceType) {
        case "heat":
          return playerResources.heat > 0;
        case "energy":
          return playerResources.energy > 0;
        case "plant":
          return playerResources.plants > 0;
        default:
          return false;
      }
    });

  // Show modal if at least one alternative payment option is available
  return canUseSteel || canUseTitanium || !!hasUsableSubstitutes;
}
