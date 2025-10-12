import React, { useState, useEffect, useCallback } from "react";
import {
  CardDto,
  CardPaymentDto,
  PaymentConstantsDto,
  ResourcesDto,
  TagBuilding,
  TagSpace,
} from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import { calculatePaymentValue } from "@/utils/paymentUtils.ts";

interface PaymentSelectionModalProps {
  cardId: string;
  card: CardDto;
  playerResources: ResourcesDto;
  paymentConstants: PaymentConstantsDto;
  onConfirm: (payment: CardPaymentDto) => void;
  onCancel: () => void;
  isVisible: boolean;
}

const PaymentSelectionModal: React.FC<PaymentSelectionModalProps> = ({
  cardId: _cardId,
  card,
  playerResources,
  paymentConstants,
  onConfirm,
  onCancel,
  isVisible,
}) => {
  // Payment state
  const [credits, setCredits] = useState(card.cost);
  const [steel, setSteel] = useState(0);
  const [titanium, setTitanium] = useState(0);

  // Reset payment when modal opens
  useEffect(() => {
    if (isVisible) {
      setCredits(card.cost);
      setSteel(0);
      setTitanium(0);
    }
  }, [isVisible, card.cost]);

  // Escape key handler
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === "Escape" && isVisible) {
        onCancel();
      }
    };

    window.addEventListener("keydown", handleEscape);
    return () => window.removeEventListener("keydown", handleEscape);
  }, [isVisible, onCancel]);

  const handleConfirm = useCallback(() => {
    // Payment validation is already checked via disabled state,
    // but double-check here for safety
    const hasResources =
      credits <= playerResources.credits &&
      steel <= playerResources.steel &&
      titanium <= playerResources.titanium;

    const paymentValue = calculatePaymentValue(
      { credits, steel, titanium },
      paymentConstants,
    );
    const paymentCoversCardCost = paymentValue >= card.cost;

    if (!hasResources || !paymentCoversCardCost) return;

    onConfirm({
      credits,
      steel,
      titanium,
    });
  }, [
    credits,
    steel,
    titanium,
    playerResources,
    paymentConstants,
    card.cost,
    onConfirm,
  ]);

  // Early return if modal is not visible
  if (!isVisible) return null;

  // Calculate current payment value
  const paymentValue = calculatePaymentValue(
    { credits, steel, titanium },
    paymentConstants,
  );

  // Check if user can afford to use selected resources and payment covers cost
  const hasResources =
    credits <= playerResources.credits &&
    steel <= playerResources.steel &&
    titanium <= playerResources.titanium;

  const paymentCoversCardCost = paymentValue >= card.cost;
  const canConfirm = hasResources && paymentCoversCardCost;

  // Increment/decrement handlers
  const incrementCredits = () => {
    if (credits < playerResources.credits) {
      setCredits(credits + 1);
    }
  };

  const decrementCredits = () => {
    if (credits > 0) {
      setCredits(credits - 1);
    }
  };

  const incrementSteel = () => {
    // Limit steel to available resources and maximum sensible payment
    const maxAllowed = Math.min(playerResources.steel, maxSteelUnits);
    if (steel < maxAllowed) {
      setSteel(steel + 1);
      // Automatically reduce credits by steel value
      const newCredits = Math.max(0, credits - paymentConstants.steelValue);
      setCredits(newCredits);
    }
  };

  const decrementSteel = () => {
    if (steel > 0) {
      setSteel(steel - 1);
      // Automatically increase credits by steel value (up to card cost)
      const currentPayment = calculatePaymentValue(
        { credits, steel: steel - 1, titanium },
        paymentConstants,
      );
      const needed = card.cost - currentPayment;
      const newCredits = Math.min(
        playerResources.credits,
        credits + Math.min(paymentConstants.steelValue, needed),
      );
      setCredits(newCredits);
    }
  };

  const incrementTitanium = () => {
    // Limit titanium to available resources and maximum sensible payment
    const maxAllowed = Math.min(playerResources.titanium, maxTitaniumUnits);
    if (titanium < maxAllowed) {
      setTitanium(titanium + 1);
      // Automatically reduce credits by titanium value
      const newCredits = Math.max(0, credits - paymentConstants.titaniumValue);
      setCredits(newCredits);
    }
  };

  const decrementTitanium = () => {
    if (titanium > 0) {
      setTitanium(titanium - 1);
      // Automatically increase credits by titanium value (up to card cost)
      const currentPayment = calculatePaymentValue(
        { credits, steel, titanium: titanium - 1 },
        paymentConstants,
      );
      const needed = card.cost - currentPayment;
      const newCredits = Math.min(
        playerResources.credits,
        credits + Math.min(paymentConstants.titaniumValue, needed),
      );
      setCredits(newCredits);
    }
  };

  // Calculate maximum allowed units to prevent excessive overpayment
  // Max units = ceil(card cost / resource value)
  const maxSteelUnits = Math.ceil(card.cost / paymentConstants.steelValue);
  const maxTitaniumUnits = Math.ceil(
    card.cost / paymentConstants.titaniumValue,
  );

  // Check which resources to show (only if player has resources AND card has appropriate tag)
  const showSteel =
    playerResources.steel > 0 && card.tags?.includes(TagBuilding);
  const showTitanium =
    playerResources.titanium > 0 && card.tags?.includes(TagSpace);

  return (
    <div className="fixed inset-0 z-[200] flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/80 backdrop-blur-sm"
        onClick={onCancel}
      />

      {/* Modal */}
      <div className="relative z-10 w-[480px] rounded-lg border border-space-blue-500 bg-space-black-light p-6 shadow-glow">
        {/* Header */}
        <div className="mb-6 border-b border-space-blue-500/30 pb-4">
          <h2 className="font-orbitron text-xl tracking-wider-2xl text-white">
            Select Payment Method
          </h2>
          <div className="mt-2 flex items-center justify-between">
            <span className="text-sm text-gray-400">Pay for: {card.name}</span>
            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-400">Cost:</span>
              <GameIcon iconType="credits" amount={card.cost} size="small" />
            </div>
          </div>
        </div>

        {/* Payment options */}
        <div className="space-y-4">
          {/* Credits - always shown */}
          <div className="flex items-center justify-between rounded-md bg-space-black/50 p-4">
            <div className="flex items-center gap-3">
              <GameIcon iconType="credits" size="medium" />
              <span className="text-white">Credits</span>
              <span className="text-xs text-gray-500">
                ({playerResources.credits} available)
              </span>
            </div>
            <div className="flex items-center gap-3">
              <button
                onClick={decrementCredits}
                disabled={credits === 0}
                className="h-8 w-8 rounded border border-space-blue-500 bg-space-black text-white hover:bg-space-blue-900 disabled:opacity-30 disabled:hover:bg-space-black"
              >
                −
              </button>
              <span className="w-12 text-center text-lg text-white">
                {credits}
              </span>
              <button
                onClick={incrementCredits}
                disabled={credits >= playerResources.credits}
                className="h-8 w-8 rounded border border-space-blue-500 bg-space-black text-white hover:bg-space-blue-900 disabled:opacity-30 disabled:hover:bg-space-black"
              >
                +
              </button>
            </div>
          </div>

          {/* Steel - shown only if player has steel */}
          {showSteel && (
            <div className="flex items-center justify-between rounded-md bg-space-black/50 p-4">
              <div className="flex items-center gap-3">
                <GameIcon iconType="steel" size="medium" />
                <span className="text-white">Steel</span>
                <span className="text-xs text-gray-400">
                  ({paymentConstants.steelValue} MC each)
                </span>
                <span className="text-xs text-gray-500">
                  ({playerResources.steel} available)
                </span>
              </div>
              <div className="flex items-center gap-3">
                <button
                  onClick={decrementSteel}
                  disabled={steel === 0}
                  className="h-8 w-8 rounded border border-space-blue-500 bg-space-black text-white hover:bg-space-blue-900 disabled:opacity-30 disabled:hover:bg-space-black"
                >
                  −
                </button>
                <span className="w-12 text-center text-lg text-white">
                  {steel}
                </span>
                <button
                  onClick={incrementSteel}
                  disabled={
                    steel >= Math.min(playerResources.steel, maxSteelUnits)
                  }
                  className="h-8 w-8 rounded border border-space-blue-500 bg-space-black text-white hover:bg-space-blue-900 disabled:opacity-30 disabled:hover:bg-space-black"
                >
                  +
                </button>
              </div>
            </div>
          )}

          {/* Titanium - shown only if player has titanium */}
          {showTitanium && (
            <div className="flex items-center justify-between rounded-md bg-space-black/50 p-4">
              <div className="flex items-center gap-3">
                <GameIcon iconType="titanium" size="medium" />
                <span className="text-white">Titanium</span>
                <span className="text-xs text-gray-400">
                  ({paymentConstants.titaniumValue} MC each)
                </span>
                <span className="text-xs text-gray-500">
                  ({playerResources.titanium} available)
                </span>
              </div>
              <div className="flex items-center gap-3">
                <button
                  onClick={decrementTitanium}
                  disabled={titanium === 0}
                  className="h-8 w-8 rounded border border-space-blue-500 bg-space-black text-white hover:bg-space-blue-900 disabled:opacity-30 disabled:hover:bg-space-black"
                >
                  −
                </button>
                <span className="w-12 text-center text-lg text-white">
                  {titanium}
                </span>
                <button
                  onClick={incrementTitanium}
                  disabled={
                    titanium >=
                    Math.min(playerResources.titanium, maxTitaniumUnits)
                  }
                  className="h-8 w-8 rounded border border-space-blue-500 bg-space-black text-white hover:bg-space-blue-900 disabled:opacity-30 disabled:hover:bg-space-black"
                >
                  +
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Payment summary */}
        <div className="mt-6 rounded-md border border-space-blue-500/30 bg-space-black/30 p-4">
          <div className="flex items-center justify-between">
            <span className="text-gray-400">Payment Value:</span>
            <div className="flex items-center gap-2">
              <span className="text-xl font-bold text-white">
                {paymentValue}
              </span>
              <GameIcon iconType="credits" size="small" />
            </div>
          </div>
          <div className="mt-2 flex items-center justify-between border-t border-space-blue-500/20 pt-2">
            <span className="text-gray-400">Card Cost:</span>
            <div className="flex items-center gap-2">
              <span className="text-xl font-bold text-white">{card.cost}</span>
              <GameIcon iconType="credits" size="small" />
            </div>
          </div>
          {!canConfirm && (
            <div className="mt-2 text-sm text-error-red">
              {!hasResources
                ? "Insufficient resources for this payment"
                : `Payment insufficient: need ${card.cost - paymentValue} more MC`}
            </div>
          )}
          {canConfirm && paymentValue > card.cost && (
            <div className="mt-2 text-sm text-yellow-500">
              Overpaying by {paymentValue - card.cost} MC (excess will be lost)
            </div>
          )}
        </div>

        {/* Action buttons */}
        <div className="mt-6 flex justify-end gap-3">
          <button
            onClick={onCancel}
            className="rounded-md border border-space-blue-500 bg-space-black px-6 py-2 text-white hover:bg-space-blue-900"
          >
            Cancel
          </button>
          <button
            onClick={handleConfirm}
            disabled={!canConfirm}
            className="rounded-md border border-space-blue-500 bg-space-blue-600 px-6 py-2 text-white hover:bg-space-blue-500 disabled:opacity-50 disabled:hover:bg-space-blue-600"
          >
            Confirm
          </button>
        </div>
      </div>
    </div>
  );
};

export default PaymentSelectionModal;
