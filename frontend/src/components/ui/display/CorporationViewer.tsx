import React, { useState, useEffect, useRef } from "react";
import { CardDto } from "@/types/generated/api-types.ts";
import { getCorporationLogo } from "@/utils/corporationLogos.tsx";
import { getCorporationBorderColor } from "@/utils/corporationColors.ts";
import CorporationCard from "../cards/CorporationCard.tsx";

interface CorporationViewerProps {
  corporation: CardDto;
}

const CorporationViewer: React.FC<CorporationViewerProps> = ({ corporation }) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const [showExpanded, setShowExpanded] = useState(false);
  const [isHovered, setIsHovered] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const borderColor = getCorporationBorderColor(corporation.name);

  const handleOpen = (e: React.MouseEvent) => {
    e.stopPropagation();
    setShowExpanded(true);
    // Small delay to allow DOM to render before animating
    requestAnimationFrame(() => {
      setIsExpanded(true);
    });
  };

  const handleClose = () => {
    setIsExpanded(false);
    // Wait for animation to complete before hiding
    setTimeout(() => {
      setShowExpanded(false);
    }, 200);
  };

  // Close when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        handleClose();
      }
    };

    if (isExpanded) {
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isExpanded]);

  return (
    <div
      ref={containerRef}
      className="fixed bottom-[150px] left-[30px] z-[999] pointer-events-auto"
    >
      {/* Collapsed: Logo Only */}
      <div
        className={`bg-black/95 rounded-lg backdrop-blur-space transition-all duration-200 cursor-pointer p-2 origin-bottom-left ${
          showExpanded ? "opacity-0 scale-75 pointer-events-none" : "opacity-100 scale-100"
        }`}
        style={{
          boxShadow: isHovered ? `0 0 20px ${borderColor}80` : `0 0 12px ${borderColor}50`,
        }}
        onClick={handleOpen}
        onMouseEnter={() => setIsHovered(true)}
        onMouseLeave={() => setIsHovered(false)}
      >
        <div className="rounded-lg p-2 bg-black/30 flex items-center justify-center min-w-[120px] min-h-[60px] [&>*]:box-content [&>*]:[filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.8))] [&>*]:rounded-[4px] [&>*>*]:rounded-[4px] [&>*>*]:[filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.8))]">
          {getCorporationLogo(corporation.name.toLowerCase())}
        </div>
      </div>

      {/* Expanded: Full Corporation Card */}
      {showExpanded && (
        <div
          className={`absolute bottom-0 left-0 origin-bottom-left transition-all duration-200 ${
            isExpanded ? "opacity-100 scale-100" : "opacity-0 scale-90"
          }`}
        >
          {/* Close button - inside card, top right */}
          <button
            onClick={(e) => {
              e.stopPropagation();
              handleClose();
            }}
            className="absolute top-4 right-4 text-white/70 hover:text-white text-xl leading-none transition-colors z-10 cursor-pointer"
          >
            ×
          </button>
          <CorporationCard
            corporation={{
              id: corporation.id,
              name: corporation.name,
              description: corporation.description,
              startingMegaCredits: corporation.startingResources?.credits || 0,
              startingProduction: corporation.startingProduction,
              startingResources: corporation.startingResources,
              behaviors: corporation.behaviors,
            }}
            isSelected={false}
            onSelect={() => {}} // No-op for in-game view
            showCheckbox={false}
            borderColor={borderColor}
            disableInteraction={true}
          />
        </div>
      )}
    </div>
  );
};

export default CorporationViewer;
