import React, { useState, useEffect, useRef } from "react";
import { CardDto } from "@/types/generated/api-types.ts";
import { getCorporationLogo } from "@/utils/corporationLogos.tsx";
import CorporationCard from "../cards/CorporationCard.tsx";

interface CorporationViewerProps {
  corporation: CardDto;
}

const CorporationViewer: React.FC<CorporationViewerProps> = ({
  corporation,
}) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const toggleExpanded = (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsExpanded(!isExpanded);
  };

  // Close when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setIsExpanded(false);
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
      title={
        isExpanded ? "" : `${corporation.name}\n${corporation.description}`
      }
    >
      {!isExpanded ? (
        /* Collapsed: Logo Only */
        <div
          className="bg-black/95 rounded-lg backdrop-blur-space transition-all duration-300 cursor-pointer p-1.5 shadow-[0_0_15px_rgba(30,60,150,0.39)] hover:-translate-y-1 hover:shadow-[0_0_25px_rgba(30,60,150,0.65)]"
          onClick={toggleExpanded}
        >
          <div className="rounded-lg p-1 bg-black/30 flex items-center justify-center [&>*]:box-content [&>*]:[filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.8))] [&>*]:rounded-[4px] [&>*>*]:rounded-[4px] [&>*>*]:[filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.8))]">
            {getCorporationLogo(corporation.name.toLowerCase())}
          </div>
        </div>
      ) : (
        /* Expanded: Full Corporation Card */
        <div onClick={toggleExpanded}>
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
          />
        </div>
      )}
    </div>
  );
};

export default CorporationViewer;
