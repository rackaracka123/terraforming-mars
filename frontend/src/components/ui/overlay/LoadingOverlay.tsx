import React, { useState, useEffect } from "react";

interface LoadingOverlayProps {
  isLoading: boolean;
  message?: string;
}

const LoadingOverlay: React.FC<LoadingOverlayProps> = ({
  isLoading,
  message = "Loading...",
}) => {
  const [isVisible, setIsVisible] = useState(isLoading);
  const [isFadingOut, setIsFadingOut] = useState(false);

  useEffect(() => {
    if (isLoading) {
      setIsVisible(true);
      setIsFadingOut(false);
    } else if (isVisible) {
      // Start fade out animation
      setIsFadingOut(true);
      // Remove overlay after animation completes
      const timeout = setTimeout(() => {
        setIsVisible(false);
      }, 300); // Match CSS transition duration (faster)
      return () => clearTimeout(timeout);
    }
    return undefined;
  }, [isLoading, isVisible]);

  if (!isVisible) return null;

  return (
    <div
      className={`absolute top-0 left-0 right-0 bottom-0 bg-black/70 backdrop-blur-space flex items-end justify-center pb-[200px] z-[9999] transition-all duration-300 ${
        isFadingOut
          ? "opacity-0 backdrop-blur-none pointer-events-none"
          : "opacity-100 pointer-events-auto"
      }`}
    >
      <div className="flex flex-col items-center gap-5">
        <div className="w-[60px] h-[60px] border-4 border-space-blue-200 border-t-space-blue-solid rounded-full animate-spin shadow-glow" />
        <div className="text-white text-lg font-medium text-shadow-glow font-orbitron tracking-wide">
          {message}
        </div>
      </div>
    </div>
  );
};

export default LoadingOverlay;
