import React, { useState, useEffect } from 'react';
import type { PaymentDto, PaymentCostDto, ResourcesDto } from '../../../types/generated/api-types';
import CostDisplay from '../display/CostDisplay';

interface PaymentSelectorProps {
  cost: PaymentCostDto;
  playerResources: ResourcesDto;
  onPaymentChange: (payment: PaymentDto) => void;
  className?: string;
}

const PaymentSelector: React.FC<PaymentSelectorProps> = ({
  cost,
  playerResources,
  onPaymentChange,
  className = '',
}) => {
  const [payment, setPayment] = useState<PaymentDto>({
    credits: 0,
    steel: 0,
    titanium: 0,
  });

  // Calculate remaining cost after applying discounts
  const calculateRemainingCost = (currentPayment: PaymentDto): number => {
    let remaining = cost.baseCost;
    
    if (cost.canUseSteel) {
      remaining -= currentPayment.steel * 2; // Steel provides 2 MC discount
    }
    
    if (cost.canUseTitanium) {
      remaining -= currentPayment.titanium * 3; // Titanium provides 3 MC discount
    }
    
    return Math.max(0, remaining);
  };

  // Check if current payment is valid
  const isPaymentValid = (currentPayment: PaymentDto): boolean => {
    // Check if player has enough resources
    if (currentPayment.credits > playerResources.credits) return false;
    if (currentPayment.steel > playerResources.steel) return false;
    if (currentPayment.titanium > playerResources.titanium) return false;
    
    // Check if remaining cost is covered by credits
    const remainingCost = calculateRemainingCost(currentPayment);
    return currentPayment.credits >= remainingCost;
  };

  // Auto-calculate minimum required credits when steel/titanium changes
  useEffect(() => {
    const remainingCost = calculateRemainingCost(payment);
    const newPayment = {
      ...payment,
      credits: remainingCost,
    };
    
    if (newPayment.credits !== payment.credits) {
      setPayment(newPayment);
    }
  }, [payment.steel, payment.titanium, cost]);

  // Notify parent when payment changes
  useEffect(() => {
    if (isPaymentValid(payment)) {
      onPaymentChange(payment);
    }
  }, [payment, onPaymentChange]);

  const updatePayment = (field: keyof PaymentDto, value: number) => {
    setPayment(prev => ({
      ...prev,
      [field]: Math.max(0, value),
    }));
  };

  const remainingCost = calculateRemainingCost(payment);
  const valid = isPaymentValid(payment);

  return (
    <div className={`payment-selector ${className}`}>
      <div className="payment-header">
        <h3>Choose Payment Method</h3>
        <div className="total-cost">
          <CostDisplay cost={cost.baseCost} size="medium" />
        </div>
      </div>

      <div className="payment-options">
        {/* MegaCredits */}
        <div className="payment-option">
          <div className="payment-label">
            <img src="/assets/resources/megacredit.png" alt="MegaCredits" width="24" height="24" />
            <span>MegaCredits</span>
          </div>
          <div className="payment-controls">
            <button 
              type="button"
              onClick={() => updatePayment('credits', payment.credits - 1)}
              disabled={payment.credits <= remainingCost}
              className="payment-button"
            >
              -
            </button>
            <input
              type="number"
              value={payment.credits}
              onChange={(e) => updatePayment('credits', parseInt(e.target.value) || 0)}
              min={remainingCost}
              max={playerResources.credits}
              className="payment-input"
            />
            <button
              type="button"
              onClick={() => updatePayment('credits', payment.credits + 1)}
              disabled={payment.credits >= playerResources.credits}
              className="payment-button"
            >
              +
            </button>
          </div>
          <div className="payment-available">
            Available: {playerResources.credits}
          </div>
        </div>

        {/* Steel */}
        {cost.canUseSteel && (
          <div className="payment-option">
            <div className="payment-label">
              <img src="/assets/resources/steel.png" alt="Steel" width="24" height="24" />
              <span>Steel</span>
              <span className="discount-info">(2 MC each)</span>
            </div>
            <div className="payment-controls">
              <button
                type="button"
                onClick={() => updatePayment('steel', payment.steel - 1)}
                disabled={payment.steel <= 0}
                className="payment-button"
              >
                -
              </button>
              <input
                type="number"
                value={payment.steel}
                onChange={(e) => updatePayment('steel', parseInt(e.target.value) || 0)}
                min={0}
                max={Math.min(playerResources.steel, Math.ceil(cost.baseCost / 2))}
                className="payment-input"
              />
              <button
                type="button"
                onClick={() => updatePayment('steel', payment.steel + 1)}
                disabled={payment.steel >= playerResources.steel || payment.steel >= Math.ceil(cost.baseCost / 2)}
                className="payment-button"
              >
                +
              </button>
            </div>
            <div className="payment-available">
              Available: {playerResources.steel}
            </div>
          </div>
        )}

        {/* Titanium */}
        {cost.canUseTitanium && (
          <div className="payment-option">
            <div className="payment-label">
              <img src="/assets/resources/titanium.png" alt="Titanium" width="24" height="24" />
              <span>Titanium</span>
              <span className="discount-info">(3 MC each)</span>
            </div>
            <div className="payment-controls">
              <button
                type="button"
                onClick={() => updatePayment('titanium', payment.titanium - 1)}
                disabled={payment.titanium <= 0}
                className="payment-button"
              >
                -
              </button>
              <input
                type="number"
                value={payment.titanium}
                onChange={(e) => updatePayment('titanium', parseInt(e.target.value) || 0)}
                min={0}
                max={Math.min(playerResources.titanium, Math.ceil(cost.baseCost / 3))}
                className="payment-input"
              />
              <button
                type="button"
                onClick={() => updatePayment('titanium', payment.titanium + 1)}
                disabled={payment.titanium >= playerResources.titanium || payment.titanium >= Math.ceil(cost.baseCost / 3)}
                className="payment-button"
              >
                +
              </button>
            </div>
            <div className="payment-available">
              Available: {playerResources.titanium}
            </div>
          </div>
        )}
      </div>

      <div className="payment-summary">
        <div className={`payment-status ${valid ? 'valid' : 'invalid'}`}>
          {valid ? '✓ Payment Valid' : '✗ Invalid Payment'}
        </div>
        <div className="remaining-cost">
          Remaining Cost: {remainingCost} MC
        </div>
      </div>

      <style jsx>{`
        .payment-selector {
          background: rgba(0, 0, 0, 0.8);
          border: 1px solid #444;
          border-radius: 8px;
          padding: 16px;
          color: #fff;
          font-family: Arial, sans-serif;
        }

        .payment-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 16px;
        }

        .payment-header h3 {
          margin: 0;
          font-size: 16px;
          color: #fff;
        }

        .payment-options {
          display: flex;
          flex-direction: column;
          gap: 12px;
          margin-bottom: 16px;
        }

        .payment-option {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 8px;
          background: rgba(255, 255, 255, 0.1);
          border-radius: 4px;
        }

        .payment-label {
          display: flex;
          align-items: center;
          gap: 8px;
          flex: 1;
        }

        .discount-info {
          font-size: 12px;
          color: #ccc;
        }

        .payment-controls {
          display: flex;
          align-items: center;
          gap: 4px;
        }

        .payment-button {
          width: 24px;
          height: 24px;
          border: 1px solid #666;
          background: rgba(255, 255, 255, 0.1);
          color: #fff;
          border-radius: 2px;
          cursor: pointer;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 14px;
        }

        .payment-button:hover:not(:disabled) {
          background: rgba(255, 255, 255, 0.2);
        }

        .payment-button:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }

        .payment-input {
          width: 60px;
          text-align: center;
          background: rgba(0, 0, 0, 0.5);
          border: 1px solid #666;
          color: #fff;
          padding: 4px;
          border-radius: 2px;
        }

        .payment-available {
          font-size: 12px;
          color: #ccc;
          min-width: 80px;
          text-align: right;
        }

        .payment-summary {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding-top: 12px;
          border-top: 1px solid #444;
        }

        .payment-status.valid {
          color: #4CAF50;
        }

        .payment-status.invalid {
          color: #f44336;
        }

        .remaining-cost {
          font-size: 14px;
          color: #fff;
        }
      `}</style>
    </div>
  );
};

export default PaymentSelector;