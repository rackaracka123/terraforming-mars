import React from "react";

interface ProductionIconProps {
  resourceIcon: string;
  value?: string | number;
  size?: "small" | "medium" | "large";
}

const ProductionIcon: React.FC<ProductionIconProps> = ({
  resourceIcon,
  value,
  size = "medium",
}) => {
  const sizeClasses = {
    small: "w-4 h-4",
    medium: "w-5 h-5",
    large: "w-6 h-6",
  };

  const valueClasses = {
    small: "text-[6px] px-0.5 py-0",
    medium: "text-[7px] px-0.5 py-px",
    large: "text-[8px] px-[3px] py-px",
  };

  return (
    <div
      className={`relative inline-flex items-center justify-center ${sizeClasses[size]}`}
    >
      <img
        src="/assets/misc/production.png"
        alt="Production"
        className="w-full h-full object-contain [filter:drop-shadow(0_1px_3px_rgba(0,0,0,0.3))]"
      />
      <img
        src={resourceIcon}
        alt="Resource"
        className="absolute w-[70%] h-[70%] object-contain top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2"
      />
      {value && (
        <span
          className={`absolute -bottom-[2px] -right-[2px] bg-black/70 text-white font-bold rounded-[2px] font-[Prototype,Arial_Black,Arial,sans-serif] leading-none ${valueClasses[size]}`}
        >
          {value}
        </span>
      )}
    </div>
  );
};

export default ProductionIcon;
