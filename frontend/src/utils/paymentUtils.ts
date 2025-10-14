import {
  CardDto,
  CardPaymentDto,
  PaymentConstantsDto,
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
): number {
  return (
    payment.credits +
    payment.steel * constants.steelValue +
    payment.titanium * constants.titaniumValue
  );
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
  };
}

/**
 * Determines if the payment modal should be shown
 * Only show if card can actually use alternative payment resources
 */
export function shouldShowPaymentModal(
  card: CardDto,
  playerResources: ResourcesDto,
): boolean {
  // If card costs 0, no payment needed
  if (card.cost === 0) {
    return false;
  }

  // Check if card can use steel (has building tag) and player has steel
  const canUseSteel =
    (card.tags?.includes(TagBuilding) ?? false) && playerResources.steel > 0;

  // Check if card can use titanium (has space tag) and player has titanium
  const canUseTitanium =
    (card.tags?.includes(TagSpace) ?? false) && playerResources.titanium > 0;

  // Show modal only if at least one alternative payment option is available
  return canUseSteel || canUseTitanium;
}
