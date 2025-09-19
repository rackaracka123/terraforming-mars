import React from "react";
import { Z_INDEX } from "../../../constants/zIndex.ts";

interface CreditIconWithAmountProps {
  amount: number;
  size?: "small" | "medium" | "large";
  className?: string;
  showMinus?: boolean;
}

const CreditIconWithAmount: React.FC<CreditIconWithAmountProps> = ({
  amount,
  size = "medium",
  className = "",
  showMinus = false,
}) => {
  const sizeMap = {
    small: { container: 24, icon: 24, fontSize: "10px" },
    medium: { container: 32, icon: 32, fontSize: "12px" },
    large: { container: 40, icon: 40, fontSize: "14px" },
  };

  const dimensions = sizeMap[size];
  const displayAmount = showMinus ? `-${amount}` : amount;

  return (
    <div
      className={`credit-icon-with-amount ${className}`}
      style={{
        position: "relative",
        display: "inline-flex",
        alignItems: "center",
        justifyContent: "center",
        width: `${dimensions.container}px`,
        height: `${dimensions.container}px`,
        borderRadius: "4px",
      }}
    >
      <img
        src="/assets/resources/megacredit.png"
        alt="Megacredits"
        style={{
          width: `${dimensions.icon}px`,
          height: `${dimensions.icon}px`,
          display: "block",
        }}
      />
      <span
        style={{
          position: "absolute",
          top: "50%",
          left: "50%",
          transform: "translate(-50%, -50%)",
          zIndex: Z_INDEX.COST_DISPLAY,
          color: showMinus ? "#ffcdd2" : "#000000",
          fontWeight: "bold",
          fontSize: dimensions.fontSize,
          textAlign: "center",
          fontFamily: "Arial, sans-serif",
          lineHeight: "1",
          textShadow: showMinus
            ? "1px 1px 2px rgba(0, 0, 0, 0.8), 0 0 3px rgba(0, 0, 0, 0.9)"
            : "0.5px 0.5px 1px rgba(255, 255, 255, 0.8)",
          whiteSpace: "nowrap",
        }}
      >
        {displayAmount}
      </span>
    </div>
  );
};

export default CreditIconWithAmount;
