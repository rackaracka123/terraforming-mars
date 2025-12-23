import React, { useState, useEffect } from "react";
import {
  GameDto,
  CardDto,
  ResourcesDto,
  ProductionDto,
  GlobalParametersDto,
  ResourceTypeCredit,
  ResourceTypeSteel,
  ResourceTypeTitanium,
  ResourceTypePlant,
  ResourceTypeEnergy,
  ResourceTypeHeat,
} from "@/types/generated/api-types.ts";
import { apiService } from "../../../services/apiService.ts";
import { globalWebSocketManager } from "../../../services/globalWebSocketManager.ts";
import GameIcon from "../display/GameIcon.tsx";
import SimpleGameCard from "../cards/SimpleGameCard.tsx";
import CorporationCard from "../cards/CorporationCard.tsx";
import {
  OVERLAY_CONTAINER_CLASS,
  OVERLAY_HEADER_CLASS,
  OVERLAY_TITLE_CLASS,
  OVERLAY_DESCRIPTION_CLASS,
  OVERLAY_FOOTER_CLASS,
  PRIMARY_BUTTON_CLASS,
} from "./overlayStyles.ts";

interface DemoSetupOverlayProps {
  game: GameDto;
  playerId: string;
}

const TEMP_MIN = -30;
const TEMP_MAX = 8;
const OXYGEN_MIN = 0;
const OXYGEN_MAX = 14;
const OCEANS_MIN = 0;
const OCEANS_MAX = 9;
const GENERATION_MIN = 1;
const GENERATION_MAX = 14;

// Stepper component - just +/- buttons and a value display
interface StepperProps {
  value: number;
  onChange: (value: number) => void;
  min?: number;
  max?: number;
  step?: number;
  label?: string;
}

const Stepper: React.FC<StepperProps> = ({
  value,
  onChange,
  min = 0,
  max = 999,
  step = 1,
  label,
}) => {
  const canDecrease = value > min;
  const canIncrease = value < max;

  return (
    <div className="flex items-center gap-1">
      {label && <span className="text-white/60 text-xs mr-2 min-w-[60px]">{label}</span>}
      <button
        onClick={() => canDecrease && onChange(Math.max(min, value - step))}
        disabled={!canDecrease}
        className={`w-8 h-8 text-lg rounded border transition-all ${
          canDecrease
            ? "bg-black/60 border-space-blue-400/50 text-white hover:border-space-blue-400 hover:bg-space-blue-600/30"
            : "bg-black/30 border-white/10 text-white/30 cursor-not-allowed"
        }`}
      >
        -
      </button>
      <span className="w-12 text-center text-white font-medium text-sm">{value}</span>
      <button
        onClick={() => canIncrease && onChange(Math.min(max, value + step))}
        disabled={!canIncrease}
        className={`w-8 h-8 text-lg rounded border transition-all ${
          canIncrease
            ? "bg-black/60 border-space-blue-400/50 text-white hover:border-space-blue-400 hover:bg-space-blue-600/30"
            : "bg-black/30 border-white/10 text-white/30 cursor-not-allowed"
        }`}
      >
        +
      </button>
    </div>
  );
};

// Resource stepper with icon
interface ResourceStepperProps {
  icon: string;
  value: number;
  onChange: (value: number) => void;
  min?: number;
  isProduction?: boolean;
}

const ResourceStepper: React.FC<ResourceStepperProps> = ({
  icon,
  value,
  onChange,
  min = 0,
  isProduction = false,
}) => {
  return (
    <div className="flex items-center gap-2 bg-black/30 rounded-lg p-2">
      <div className="relative">
        <GameIcon iconType={icon} size="small" />
        {isProduction && (
          <div className="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 bg-amber-600 rounded-full border border-black/50" />
        )}
      </div>
      <button
        onClick={() => onChange(Math.max(min, value - 1))}
        className="w-6 h-6 text-sm rounded bg-black/40 border border-space-blue-400/30 text-white hover:border-space-blue-400 transition-all"
      >
        -
      </button>
      <span className="w-8 text-center text-white font-medium text-sm">{value}</span>
      <button
        onClick={() => onChange(value + 1)}
        className="w-6 h-6 text-sm rounded bg-black/40 border border-space-blue-400/30 text-white hover:border-space-blue-400 transition-all"
      >
        +
      </button>
    </div>
  );
};

const DemoSetupOverlay: React.FC<DemoSetupOverlayProps> = ({ game, playerId }) => {
  const isHost = game.hostPlayerId === playerId;

  // Global parameters (host only)
  const [globalParams, setGlobalParams] = useState<GlobalParametersDto>({
    temperature: game.globalParameters?.temperature ?? TEMP_MIN,
    oxygen: game.globalParameters?.oxygen ?? OXYGEN_MIN,
    oceans: game.globalParameters?.oceans ?? OCEANS_MIN,
  });
  const [generation, setGeneration] = useState(game.generation ?? 1);

  // Player setup
  const [availableCorporations, setAvailableCorporations] = useState<CardDto[]>([]);
  const [availableCards, setAvailableCards] = useState<CardDto[]>([]);
  const [selectedCorporationId, setSelectedCorporationId] = useState<string>("");
  const [selectedCardIds, setSelectedCardIds] = useState<string[]>([]);
  const [cardSearchTerm, setCardSearchTerm] = useState("");
  const [corpSearchTerm, setCorpSearchTerm] = useState("");

  // Resources
  const [resources, setResources] = useState<ResourcesDto>({
    credits: 0,
    steel: 0,
    titanium: 0,
    plants: 0,
    energy: 0,
    heat: 0,
  });

  // Production
  const [production, setProduction] = useState<ProductionDto>({
    credits: 0,
    steel: 0,
    titanium: 0,
    plants: 0,
    energy: 0,
    heat: 0,
  });

  // Terraform rating
  const [terraformRating, setTerraformRating] = useState(20);

  // Loading state
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Load corporations and cards on mount
  useEffect(() => {
    const loadCardsData = async () => {
      try {
        const response = await apiService.listCards(0, 1000);
        const corps: CardDto[] = [];
        const projectCards: CardDto[] = [];
        for (const card of response.cards) {
          if (card.type === "corporation") {
            corps.push(card);
          } else if (card.type !== "prelude") {
            projectCards.push(card);
          }
        }
        setAvailableCorporations(corps);
        setAvailableCards(projectCards);
      } catch (err) {
        console.error("Failed to load cards:", err);
      }
    };

    void loadCardsData();
  }, []);

  // When a corporation is selected, apply its starting resources and production
  useEffect(() => {
    if (!selectedCorporationId) return;

    const corp = availableCorporations.find((c) => c.id === selectedCorporationId);
    if (corp) {
      if (corp.startingResources) {
        setResources({
          credits: corp.startingResources.credits ?? 0,
          steel: corp.startingResources.steel ?? 0,
          titanium: corp.startingResources.titanium ?? 0,
          plants: corp.startingResources.plants ?? 0,
          energy: corp.startingResources.energy ?? 0,
          heat: corp.startingResources.heat ?? 0,
        });
      }
      if (corp.startingProduction) {
        setProduction({
          credits: corp.startingProduction.credits ?? 0,
          steel: corp.startingProduction.steel ?? 0,
          titanium: corp.startingProduction.titanium ?? 0,
          plants: corp.startingProduction.plants ?? 0,
          energy: corp.startingProduction.energy ?? 0,
          heat: corp.startingProduction.heat ?? 0,
        });
      }
    }
  }, [selectedCorporationId, availableCorporations]);

  const toggleCardSelection = (cardId: string) => {
    setSelectedCardIds((prev) =>
      prev.includes(cardId) ? prev.filter((id) => id !== cardId) : [...prev, cardId],
    );
  };

  const filteredCards = availableCards.filter(
    (card) =>
      card.name.toLowerCase().includes(cardSearchTerm.toLowerCase()) ||
      card.id.toLowerCase().includes(cardSearchTerm.toLowerCase()),
  );

  const filteredCorporations = availableCorporations.filter(
    (corp) =>
      corp.name.toLowerCase().includes(corpSearchTerm.toLowerCase()) ||
      corp.id.toLowerCase().includes(corpSearchTerm.toLowerCase()),
  );

  const selectedCorporation = availableCorporations.find((c) => c.id === selectedCorporationId);

  const handleConfirm = async () => {
    if (isSubmitting) return;

    setIsSubmitting(true);
    try {
      await globalWebSocketManager.confirmDemoSetup({
        corporationId: selectedCorporationId || undefined,
        cardIds: selectedCardIds,
        resources,
        production,
        terraformRating,
        globalParameters: isHost ? globalParams : undefined,
        generation: isHost ? generation : undefined,
      });
    } catch (err) {
      console.error("Failed to confirm demo setup:", err);
      setIsSubmitting(false);
    }
  };

  const resourceTypes = [
    { key: "credits" as const, icon: ResourceTypeCredit },
    { key: "steel" as const, icon: ResourceTypeSteel },
    { key: "titanium" as const, icon: ResourceTypeTitanium },
    { key: "plants" as const, icon: ResourceTypePlant },
    { key: "energy" as const, icon: ResourceTypeEnergy },
    { key: "heat" as const, icon: ResourceTypeHeat },
  ];

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center bg-black/70 backdrop-blur-sm animate-[fadeIn_0.3s_ease]">
      <div className={`${OVERLAY_CONTAINER_CLASS} max-w-[1200px] max-h-[90vh]`}>
        {/* Header */}
        <div className={OVERLAY_HEADER_CLASS}>
          <h2 className={OVERLAY_TITLE_CLASS}>Demo Game Setup</h2>
          <p className={OVERLAY_DESCRIPTION_CLASS}>
            Configure your starting setup.{" "}
            {isHost ? "As host, you can also set global parameters." : ""}
          </p>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-4">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
            {/* Column 1: Global Params + Resources */}
            <div className="space-y-4">
              {/* Global Parameters (Host only) */}
              {isHost && (
                <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-3">
                  <h3 className="text-white font-semibold mb-3 uppercase tracking-wide text-xs">
                    Global Parameters
                  </h3>
                  <div className="space-y-2">
                    <Stepper
                      label="Temp"
                      value={globalParams.temperature}
                      onChange={(v) => setGlobalParams((p) => ({ ...p, temperature: v }))}
                      min={TEMP_MIN}
                      max={TEMP_MAX}
                      step={2}
                    />
                    <Stepper
                      label="Oxygen"
                      value={globalParams.oxygen}
                      onChange={(v) => setGlobalParams((p) => ({ ...p, oxygen: v }))}
                      min={OXYGEN_MIN}
                      max={OXYGEN_MAX}
                    />
                    <Stepper
                      label="Oceans"
                      value={globalParams.oceans}
                      onChange={(v) => setGlobalParams((p) => ({ ...p, oceans: v }))}
                      min={OCEANS_MIN}
                      max={OCEANS_MAX}
                    />
                    <Stepper
                      label="Gen"
                      value={generation}
                      onChange={setGeneration}
                      min={GENERATION_MIN}
                      max={GENERATION_MAX}
                    />
                  </div>
                </div>
              )}

              {/* Terraform Rating */}
              <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-3">
                <h3 className="text-white font-semibold mb-3 uppercase tracking-wide text-xs">
                  Terraform Rating
                </h3>
                <Stepper value={terraformRating} onChange={setTerraformRating} min={0} max={100} />
              </div>

              {/* Resources */}
              <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-3">
                <h3 className="text-white font-semibold mb-3 uppercase tracking-wide text-xs">
                  Resources
                </h3>
                <div className="grid grid-cols-2 gap-2">
                  {resourceTypes.map(({ key, icon }) => (
                    <ResourceStepper
                      key={key}
                      icon={icon}
                      value={resources[key]}
                      onChange={(v) => setResources((p) => ({ ...p, [key]: v }))}
                    />
                  ))}
                </div>
              </div>

              {/* Production */}
              <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-3">
                <h3 className="text-white font-semibold mb-3 uppercase tracking-wide text-xs">
                  Production
                </h3>
                <div className="grid grid-cols-2 gap-2">
                  {resourceTypes.map(({ key, icon }) => (
                    <ResourceStepper
                      key={key}
                      icon={icon}
                      value={production[key]}
                      onChange={(v) => setProduction((p) => ({ ...p, [key]: v }))}
                      min={key === "credits" ? -5 : 0}
                      isProduction
                    />
                  ))}
                </div>
              </div>
            </div>

            {/* Column 2: Corporation Selection */}
            <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-3 flex flex-col max-h-[70vh]">
              <h3 className="text-white font-semibold mb-2 uppercase tracking-wide text-xs shrink-0">
                Corporation{" "}
                <span className="text-white/50 font-normal normal-case">
                  ({selectedCorporationId ? "1 selected" : "Random"})
                </span>
              </h3>
              <input
                type="text"
                placeholder="Search corporations..."
                value={corpSearchTerm}
                onChange={(e) => setCorpSearchTerm(e.target.value)}
                className="w-full bg-black/60 border border-space-blue-400/30 rounded-lg py-2 px-3 text-white text-sm outline-none focus:border-space-blue-400 mb-3 shrink-0"
              />
              <div className="flex flex-col gap-4 items-center flex-1 min-h-0 overflow-y-auto">
                {/* Random option */}
                <button
                  onClick={() => setSelectedCorporationId("")}
                  className={`w-[400px] h-[60px] rounded-lg border-2 p-2 transition-all text-center flex items-center justify-center shrink-0 ${
                    !selectedCorporationId
                      ? "border-yellow-400 bg-yellow-900/30"
                      : "border-white/20 bg-black/30 hover:border-yellow-400/50"
                  }`}
                >
                  <span className="text-white/80 text-sm font-medium">Random Corporation</span>
                </button>
                {filteredCorporations.map((corp) => (
                  <div key={corp.id} className="shrink-0">
                    <CorporationCard
                      corporation={{
                        id: corp.id,
                        name: corp.name,
                        description: corp.description ?? "",
                        startingMegaCredits: corp.startingResources?.credits ?? 0,
                        startingProduction: corp.startingProduction
                          ? {
                              credits: corp.startingProduction.credits,
                              steel: corp.startingProduction.steel,
                              titanium: corp.startingProduction.titanium,
                              plants: corp.startingProduction.plants,
                              energy: corp.startingProduction.energy,
                              heat: corp.startingProduction.heat,
                            }
                          : undefined,
                        startingResources: corp.startingResources
                          ? {
                              credits: corp.startingResources.credits,
                              steel: corp.startingResources.steel,
                              titanium: corp.startingResources.titanium,
                              plants: corp.startingResources.plants,
                              energy: corp.startingResources.energy,
                              heat: corp.startingResources.heat,
                            }
                          : undefined,
                        behaviors: corp.behaviors,
                        logoPath: undefined,
                      }}
                      isSelected={selectedCorporationId === corp.id}
                      onSelect={() =>
                        setSelectedCorporationId(selectedCorporationId === corp.id ? "" : corp.id)
                      }
                      showCheckbox
                    />
                  </div>
                ))}
              </div>
            </div>

            {/* Column 3: Card Selection */}
            <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-3 flex flex-col max-h-[70vh]">
              <h3 className="text-white font-semibold mb-2 uppercase tracking-wide text-xs shrink-0">
                Starting Cards{" "}
                <span className="text-white/50 font-normal normal-case">
                  ({selectedCardIds.length} selected)
                </span>
              </h3>
              <input
                type="text"
                placeholder="Search cards..."
                value={cardSearchTerm}
                onChange={(e) => setCardSearchTerm(e.target.value)}
                className="w-full bg-black/60 border border-space-blue-400/30 rounded-lg py-2 px-3 text-white text-sm outline-none focus:border-space-blue-400 mb-3 shrink-0"
              />
              <div className="flex flex-wrap gap-2 justify-center content-start flex-1 min-h-0 overflow-y-auto">
                {filteredCards.slice(0, 50).map((card) => (
                  <div key={card.id} className="shrink-0">
                    <SimpleGameCard
                      card={card}
                      isSelected={selectedCardIds.includes(card.id)}
                      onSelect={() => toggleCardSelection(card.id)}
                      animationDelay={0}
                      showCheckbox
                    />
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className={OVERLAY_FOOTER_CLASS}>
          <div className="text-white/60 text-sm">
            {selectedCorporationId ? (
              <span>Corporation: {selectedCorporation?.name}</span>
            ) : (
              <span>Corporation: Random</span>
            )}
            {selectedCardIds.length > 0 && (
              <span className="ml-4">Cards: {selectedCardIds.length}</span>
            )}
          </div>
          <button
            className={PRIMARY_BUTTON_CLASS}
            onClick={() => void handleConfirm()}
            disabled={isSubmitting}
          >
            {isSubmitting ? "Confirming..." : "Confirm Setup"}
          </button>
        </div>
      </div>
    </div>
  );
};

export default DemoSetupOverlay;
