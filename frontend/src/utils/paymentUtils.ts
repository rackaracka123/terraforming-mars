import {
  CardDto,
  CardPaymentDto,
  PaymentConstantsDto,
  ResourcesDto,
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
 * Skip modal ONLY if player has ZERO steel AND ZERO titanium
 * Let backend decide if tags allow usage - we just check if resources exist
 */
export function shouldShowPaymentModal(
  card: CardDto,
  playerResources: ResourcesDto,
): boolean {
  // If player has no alternative resources, skip the modal
  if (playerResources.steel === 0 && playerResources.titanium === 0) {
    return false;
  }

  // If card costs 0, no payment needed
  if (card.cost === 0) {
    return false;
  }

  // Show modal if player has at least one alternative payment resource
  return true;
}
