import type { PaymentCostDto, CardDto, CardTag, TagBuilding, TagSpace } from '../types/generated/api-types';

// Re-export tag constants for easier use
export { TagBuilding, TagSpace } from '../types/generated/api-types';

/**
 * Determines the payment cost structure for a card based on its tags
 */
export function getCardPaymentCost(card: CardDto): PaymentCostDto {
  const canUseSteel = card.tags.includes('building' as CardTag);
  const canUseTitanium = card.tags.includes('space' as CardTag);

  return {
    baseCost: card.cost,
    canUseSteel,
    canUseTitanium,
  };
}

/**
 * Creates a payment cost for standard projects (MegaCredits only)
 */
export function getStandardProjectCost(baseCost: number): PaymentCostDto {
  return {
    baseCost,
    canUseSteel: false,
    canUseTitanium: false,
  };
}

/**
 * Calculates the effective cost of a payment after discounts
 */
export function calculateEffectiveCost(payment: { credits: number; steel: number; titanium: number }, cost: PaymentCostDto): number {
  let effectiveCost = cost.baseCost;
  
  if (cost.canUseSteel) {
    effectiveCost -= payment.steel * 2; // Steel provides 2 MC discount
  }
  
  if (cost.canUseTitanium) {
    effectiveCost -= payment.titanium * 3; // Titanium provides 3 MC discount
  }
  
  return Math.max(0, effectiveCost);
}

/**
 * Checks if a player can afford a payment with their resources
 */
export function canAffordPayment(
  payment: { credits: number; steel: number; titanium: number },
  playerResources: { credits: number; steel: number; titanium: number }
): boolean {
  return (
    payment.credits <= playerResources.credits &&
    payment.steel <= playerResources.steel &&
    payment.titanium <= playerResources.titanium
  );
}

/**
 * Validates if a payment is valid for a specific cost
 */
export function isValidPayment(
  payment: { credits: number; steel: number; titanium: number },
  cost: PaymentCostDto
): boolean {
  // Check if trying to use steel when not allowed
  if (payment.steel > 0 && !cost.canUseSteel) {
    return false;
  }
  
  // Check if trying to use titanium when not allowed
  if (payment.titanium > 0 && !cost.canUseTitanium) {
    return false;
  }
  
  // Check if the effective cost is covered by credits
  const effectiveCost = calculateEffectiveCost(payment, cost);
  return payment.credits >= effectiveCost;
}

/**
 * Suggests an optimal payment for a given cost and player resources
 * Prioritizes titanium over steel for better efficiency
 */
export function suggestOptimalPayment(
  cost: PaymentCostDto,
  playerResources: { credits: number; steel: number; titanium: number }
): { credits: number; steel: number; titanium: number } {
  let remainingCost = cost.baseCost;
  let titaniumToUse = 0;
  let steelToUse = 0;
  
  // Use titanium first (better discount rate of 3 MC per unit)
  if (cost.canUseTitanium) {
    titaniumToUse = Math.min(playerResources.titanium, Math.ceil(remainingCost / 3));
    remainingCost -= titaniumToUse * 3;
  }
  
  // Use steel next (2 MC discount per unit)
  if (cost.canUseSteel && remainingCost > 0) {
    steelToUse = Math.min(playerResources.steel, Math.ceil(remainingCost / 2));
    remainingCost -= steelToUse * 2;
  }
  
  // Cover remaining cost with credits
  const creditsToUse = Math.max(0, remainingCost);
  
  return {
    credits: creditsToUse,
    steel: steelToUse,
    titanium: titaniumToUse,
  };
}