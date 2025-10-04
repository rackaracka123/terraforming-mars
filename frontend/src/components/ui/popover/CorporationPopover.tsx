import React, { useEffect, useRef, useState } from "react";
import { CardDto } from "../../../types/generated/api-types.ts";

interface CorporationPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  corporation: CardDto;
  anchorRef: React.RefObject<HTMLElement>;
}

const CorporationPopover: React.FC<CorporationPopoverProps> = ({
  isVisible,
  onClose,
  corporation,
  anchorRef,
}) => {
  const popoverRef = useRef<HTMLDivElement>(null);
  const [position, setPosition] = useState({ bottom: 85, left: 30 });

  useEffect(() => {
    if (isVisible && anchorRef.current) {
      const rect = anchorRef.current.getBoundingClientRect();
      const padding = 30;

      const bottom = window.innerHeight - rect.top + 15;
      const left = rect.left;

      setPosition({ bottom, left: Math.max(padding, left) });
    }
  }, [isVisible, anchorRef]);

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };

    const handleClickOutside = (event: MouseEvent) => {
      if (
        popoverRef.current &&
        !popoverRef.current.contains(event.target as Node) &&
        anchorRef.current &&
        !anchorRef.current.contains(event.target as Node)
      ) {
        onClose();
      }
    };

    if (isVisible) {
      document.addEventListener("keydown", handleEscape);
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("keydown", handleEscape);
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isVisible, onClose, anchorRef]);

  if (!isVisible) return null;

  return (
    <div
      className="fixed w-[380px] bg-space-black-darker/95 border-2 border-space-blue-400 rounded-xl shadow-glow-lg backdrop-blur-space z-[10001] animate-[popoverSlideUp_0.3s_ease-out] flex flex-col overflow-hidden isolate pointer-events-auto max-[768px]:w-[320px]"
      ref={popoverRef}
      style={{ bottom: `${position.bottom}px`, left: `${position.left}px` }}
    >
      {/* Arrow pointing down to the corporation display */}
      <div className="absolute -bottom-2 left-[30px] w-0 h-0 border-l-[8px] border-l-transparent border-r-[8px] border-r-transparent border-t-[8px] border-t-space-blue-400" />

      {/* Header */}
      <div className="flex items-center justify-between py-[15px] px-5 bg-black/40 border-b border-b-space-blue-400/60">
        <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
          {corporation.name}
        </h3>
        <div className="text-white/80 text-xs bg-space-blue-400/20 py-1 px-2 rounded-md border border-space-blue-400/30 uppercase tracking-wider">
          Corporation
        </div>
      </div>

      {/* Corporation Card Image */}
      <div className="flex justify-center p-4 bg-black/20">
        <div className="w-48 h-64 rounded-lg overflow-hidden shadow-[0_8px_24px_rgba(0,0,0,0.8)]">
          <img
            src={`/assets/cards/${corporation.id}.webp`}
            alt={corporation.name}
            className="w-full h-full object-cover"
            onError={(e) => {
              e.currentTarget.src = "/assets/cards/001.webp";
            }}
          />
        </div>
      </div>

      {/* Description */}
      <div className="px-5 pb-4">
        <div className="text-sm text-white/90 leading-relaxed">
          {corporation.description}
        </div>
      </div>

      {/* Tags (if any) */}
      {corporation.tags && corporation.tags.length > 0 && (
        <div className="px-5 pb-4 border-t border-white/10 pt-3">
          <div className="text-xs text-white/70 uppercase tracking-wider mb-2">
            Tags
          </div>
          <div className="flex flex-wrap gap-2">
            {corporation.tags.map((tag, index) => (
              <div
                key={index}
                className="bg-space-blue-400/20 text-white/90 text-xs py-1 px-2.5 rounded-md border border-space-blue-400/30"
              >
                {tag}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Close hint */}
      <div className="px-5 pb-3 text-center text-xs text-white/50">
        Click outside or press ESC to close
      </div>
    </div>
  );
};

export default CorporationPopover;
