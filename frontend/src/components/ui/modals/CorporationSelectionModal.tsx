import React, { useState, useRef } from "react";
import CorporationCard from "../cards/CorporationCard.tsx";
import { CardBehaviorDto } from "@/types/generated/api-types.ts";

interface Corporation {
  id: string;
  name: string;
  description: string;
  startingMegaCredits: number;
  startingProduction?: {
    credits?: number;
    steel?: number;
    titanium?: number;
    plants?: number;
    energy?: number;
    heat?: number;
  };
  startingResources?: {
    credits?: number;
    steel?: number;
    titanium?: number;
    plants?: number;
    energy?: number;
    heat?: number;
  };
  behaviors?: CardBehaviorDto[];
  expansion?: string;
  logoPath?: string;
}

interface CorporationSelectionModalProps {
  corporations: Corporation[];
  onSelectCorporation: (corporationId: string) => void;
  isVisible: boolean;
}

const CorporationSelectionModal: React.FC<CorporationSelectionModalProps> = ({
  corporations,
  onSelectCorporation,
  isVisible,
}) => {
  const [selectedCorporation, setSelectedCorporation] = useState<string | null>(
    null,
  );
  const [isFlashing, setIsFlashing] = useState(false);
  const modalRef = useRef<HTMLDivElement>(null);

  if (!isVisible) return null;

  const handleOverlayClick = (e: React.MouseEvent) => {
    // Prevent click from reaching overlay if clicking on modal content
    if (modalRef.current && modalRef.current.contains(e.target as Node)) {
      return;
    }

    // Flash animation when trying to dismiss
    setIsFlashing(true);
    setTimeout(() => setIsFlashing(false), 600);
  };

  const handleCorporationSelect = (corporationId: string) => {
    setSelectedCorporation(corporationId);
  };

  const handleConfirmSelection = () => {
    if (selectedCorporation) {
      onSelectCorporation(selectedCorporation);
    }
  };

  return (
    <div
      className={`fixed inset-0 bg-[rgba(0,0,0,0.9)] backdrop-blur-[10px] flex items-center justify-center isolate p-5 transition-all duration-300 ease-in-out ${isFlashing ? "animate-[flashBorder_0.6s_ease]" : ""}`}
      onClick={handleOverlayClick}
    >
      <div
        className="bg-gradient-to-br from-[rgba(10,20,40,0.98)] via-[rgba(20,30,50,0.96)] to-[rgba(15,25,45,0.98)] border-2 border-[rgba(100,150,255,0.5)] rounded-[20px] max-w-[1200px] max-h-[90vh] w-full overflow-y-auto backdrop-blur-[20px] shadow-[0_20px_60px_rgba(0,0,0,0.8),0_0_40px_rgba(100,150,255,0.3)] animate-[modalPulse_2s_ease-in-out_infinite]"
        ref={modalRef}
      >
        <div className="text-center px-[30px] pt-[30px] pb-5 border-b border-[rgba(100,150,255,0.3)]">
          <h2 className="text-[32px] text-white mb-2 shadow-[0_2px_4px_rgba(0,0,0,0.8)] animate-[headerPulse_3s_ease-in-out_infinite] max-[800px]:text-2xl">
            Choose Your Corporation
          </h2>
          <p className="text-base text-[rgba(255,255,255,0.7)] m-0">
            Select a corporation to begin your Mars terraforming journey
          </p>
        </div>

        <div className="grid grid-cols-[repeat(auto-fit,minmax(350px,1fr))] gap-5 p-[30px] max-[800px]:grid-cols-1 max-[800px]:p-5">
          {corporations.map((corp) => (
            <CorporationCard
              key={corp.id}
              corporation={corp}
              isSelected={selectedCorporation === corp.id}
              onSelect={handleCorporationSelect}
            />
          ))}
        </div>

        <div className="px-[30px] pt-5 pb-[30px] text-center border-t border-[rgba(100,150,255,0.3)]">
          <button
            className={`bg-gradient-to-br from-[#4a90e2] to-[#5ba0f2] text-white border-none rounded-lg px-[30px] py-3 text-base font-bold cursor-pointer transition-all duration-200 ease-in-out relative overflow-hidden ${
              !selectedCorporation
                ? "bg-[rgba(100,100,100,0.5)] text-[rgba(255,255,255,0.5)] cursor-not-allowed transform-none"
                : "animate-[buttonPulse_2.5s_ease-in-out_infinite] hover:bg-gradient-to-br hover:from-[#357abd] hover:to-[#4a90e2] hover:-translate-y-px"
            }`}
            disabled={!selectedCorporation}
            onClick={handleConfirmSelection}
          >
            Confirm Selection
          </button>
        </div>
      </div>
    </div>
  );
};

export default CorporationSelectionModal;
