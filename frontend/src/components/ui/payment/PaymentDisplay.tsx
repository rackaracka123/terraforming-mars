import React from "react";
import type {
  PaymentCostDto,
  CardTag,
} from "../../../types/generated/api-types";
import CostDisplay from "../display/CostDisplay";

interface PaymentDisplayProps {
  cost: PaymentCostDto;
  tags?: CardTag[];
  size?: "small" | "medium" | "large";
  className?: string;
}

const PaymentDisplay: React.FC<PaymentDisplayProps> = ({
  cost,
  tags = [],
  size = "medium",
  className = "",
}) => {
  return (
    <div className={`payment-display ${className}`}>
      <CostDisplay cost={cost.baseCost} size={size} />

      {(cost.canUseSteel || cost.canUseTitanium) && (
        <div className="payment-alternatives">
          {cost.canUseSteel && (
            <div className="alternative-payment">
              <img
                src="/assets/resources/steel.png"
                alt="Can use Steel"
                width={size === "small" ? 16 : size === "medium" ? 20 : 24}
                height={size === "small" ? 16 : size === "medium" ? 20 : 24}
                title="Can pay with Steel (2 MC per Steel)"
              />
            </div>
          )}

          {cost.canUseTitanium && (
            <div className="alternative-payment">
              <img
                src="/assets/resources/titanium.png"
                alt="Can use Titanium"
                width={size === "small" ? 16 : size === "medium" ? 20 : 24}
                height={size === "small" ? 16 : size === "medium" ? 20 : 24}
                title="Can pay with Titanium (3 MC per Titanium)"
              />
            </div>
          )}
        </div>
      )}

      <style jsx>{`
        .payment-display {
          display: flex;
          align-items: center;
          gap: 4px;
        }

        .payment-alternatives {
          display: flex;
          gap: 2px;
          margin-left: 4px;
        }

        .alternative-payment {
          display: flex;
          align-items: center;
          opacity: 0.8;
        }

        .alternative-payment img {
          filter: drop-shadow(0 0 2px rgba(255, 255, 255, 0.3));
        }
      `}</style>
    </div>
  );
};

export default PaymentDisplay;
