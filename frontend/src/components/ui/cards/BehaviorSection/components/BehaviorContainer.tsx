import React from "react";
import { ClassifiedBehavior } from "../types.ts";

interface BehaviorContainerProps {
  classifiedBehavior: ClassifiedBehavior;
  index: number;
  children: React.ReactNode;
}

const BehaviorContainer: React.FC<BehaviorContainerProps> = ({
  classifiedBehavior,
  index,
  children,
}) => {
  const { type } = classifiedBehavior;

  if (type === "auto-no-background") {
    return (
      <div
        key={index}
        className="flex items-center justify-center my-px p-[3px] min-h-8 max-md:p-px max-md:my-px"
      >
        {children}
      </div>
    );
  } else {
    const typeStyles = {
      "manual-action":
        "bg-[linear-gradient(135deg,rgba(33,150,243,0.35)_0%,rgba(25,118,210,0.3)_100%)] border-[rgba(33,150,243,0.5)] shadow-[0_2px_4px_rgba(33,150,243,0.3)]",
      "triggered-effect": "bg-white/[0.08] border-white/20 shadow-[0_1px_3px_rgba(0,0,0,0.15)]",
      discount: "bg-white/[0.08] border-white/20 shadow-[0_1px_3px_rgba(0,0,0,0.15)]",
      "payment-substitute": "bg-white/[0.08] border-white/20 shadow-[0_1px_3px_rgba(0,0,0,0.15)]",
      "value-modifier": "bg-white/[0.08] border-white/20 shadow-[0_1px_3px_rgba(0,0,0,0.15)]",
      defense: "bg-white/[0.08] border-white/20 shadow-[0_1px_3px_rgba(0,0,0,0.15)]",
      "immediate-production":
        "bg-[linear-gradient(135deg,rgba(139,89,42,0.35)_0%,rgba(101,67,33,0.3)_100%)] border-[rgba(139,89,42,0.5)] shadow-[0_2px_4px_rgba(139,89,42,0.25)]",
      "immediate-effect": "bg-white/[0.08] border-white/20 shadow-[0_1px_3px_rgba(0,0,0,0.15)]",
    };

    const widthClass =
      type === "manual-action" ||
      type === "triggered-effect" ||
      type === "discount" ||
      type === "payment-substitute" ||
      type === "value-modifier" ||
      type === "defense"
        ? "w-fit"
        : "w-[calc(100%-20px)]";

    return (
      <div
        key={index}
        className={`rounded-[3px] px-2 py-1 min-h-8 my-px border border-white/10 backdrop-blur-[2px] flex items-center ${widthClass} ${typeStyles[type] || ""} max-md:px-1.5 max-md:py-[3px] max-md:min-h-7 max-md:my-px`}
      >
        <div className="flex items-center gap-1.5 flex-nowrap w-full justify-center max-md:gap-1">
          {children}
        </div>
      </div>
    );
  }
};

export default BehaviorContainer;
