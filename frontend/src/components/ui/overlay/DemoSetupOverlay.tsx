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
} from "../../../types/generated/api-types.ts";
import { apiService } from "../../../services/apiService.ts";
import { globalWebSocketManager } from "../../../services/globalWebSocketManager.ts";
import GameIcon from "../display/GameIcon.tsx";
import {
  OVERLAY_CONTAINER_CLASS,
  OVERLAY_HEADER_CLASS,
  OVERLAY_TITLE_CLASS,
  OVERLAY_DESCRIPTION_CLASS,
  OVERLAY_FOOTER_CLASS,
  PRIMARY_BUTTON_CLASS,
  INPUT_CLASS,
  INPUT_SMALL_CLASS,
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

const DemoSetupOverlay: React.FC<DemoSetupOverlayProps> = ({
  game,
  playerId,
}) => {
  const isHost = game.hostPlayerId === playerId;

  // Global parameters (host only)
  const [globalParams, setGlobalParams] = useState<GlobalParametersDto>({
    temperature: game.globalParameters?.temperature ?? TEMP_MIN,
    oxygen: game.globalParameters?.oxygen ?? OXYGEN_MIN,
    oceans: game.globalParameters?.oceans ?? OCEANS_MIN,
  });
  const [generation, setGeneration] = useState(game.generation ?? 1);

  // Player setup
  const [availableCorporations, setAvailableCorporations] = useState<CardDto[]>(
    [],
  );
  const [availableCards, setAvailableCards] = useState<CardDto[]>([]);
  const [selectedCorporationId, setSelectedCorporationId] =
    useState<string>("");
  const [selectedCardIds, setSelectedCardIds] = useState<string[]>([]);
  const [cardSearchTerm, setCardSearchTerm] = useState("");
  const [showCardSelection, setShowCardSelection] = useState(false);

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
        // Single pass to categorize cards
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

    const corp = availableCorporations.find(
      (c) => c.id === selectedCorporationId,
    );
    if (corp) {
      // Apply starting resources
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
      // Apply starting production
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
      prev.includes(cardId)
        ? prev.filter((id) => id !== cardId)
        : [...prev, cardId],
    );
  };

  const filteredCards = availableCards.filter(
    (card) =>
      card.name.toLowerCase().includes(cardSearchTerm.toLowerCase()) ||
      card.id.toLowerCase().includes(cardSearchTerm.toLowerCase()),
  );

  const selectedCorporation = availableCorporations.find(
    (c) => c.id === selectedCorporationId,
  );

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

  const handleResourceChange = (
    resource: keyof ResourcesDto,
    value: number,
  ) => {
    setResources((prev) => ({ ...prev, [resource]: Math.max(0, value) }));
  };

  const handleProductionChange = (
    resource: keyof ProductionDto,
    value: number,
  ) => {
    // Production can be negative for credits
    const minValue = resource === "credits" ? -5 : 0;
    setProduction((prev) => ({
      ...prev,
      [resource]: Math.max(minValue, value),
    }));
  };

  const resourceTypes = [
    { key: "credits" as const, icon: ResourceTypeCredit, label: "Credits" },
    { key: "steel" as const, icon: ResourceTypeSteel, label: "Steel" },
    { key: "titanium" as const, icon: ResourceTypeTitanium, label: "Titanium" },
    { key: "plants" as const, icon: ResourceTypePlant, label: "Plants" },
    { key: "energy" as const, icon: ResourceTypeEnergy, label: "Energy" },
    { key: "heat" as const, icon: ResourceTypeHeat, label: "Heat" },
  ];

  // Select all text on focus for better UX with number inputs
  const handleSelectAll = (e: React.FocusEvent<HTMLInputElement>) =>
    e.target.select();

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center bg-black/60 backdrop-blur-sm animate-[fadeIn_0.3s_ease]">
      <div className={`${OVERLAY_CONTAINER_CLASS} max-w-[1000px]`}>
        {/* Header */}
        <div className={OVERLAY_HEADER_CLASS}>
          <h2 className={OVERLAY_TITLE_CLASS}>Demo Game Setup</h2>
          <p className={OVERLAY_DESCRIPTION_CLASS}>
            Configure your starting setup.{" "}
            {isHost ? "As host, you can also set global parameters." : ""}
          </p>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Left Column: Global Parameters (Host only) + Corporation */}
            <div className="space-y-6">
              {/* Global Parameters (Host only) */}
              {isHost && (
                <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-4">
                  <h3 className="text-white font-semibold mb-4 uppercase tracking-wide text-sm">
                    Global Parameters
                  </h3>
                  <div className="grid grid-cols-2 gap-4">
                    {/* Temperature */}
                    <div>
                      <label className="text-white/60 text-xs block mb-1">
                        Temperature ({TEMP_MIN} to {TEMP_MAX}°C)
                      </label>
                      <input
                        type="number"
                        min={TEMP_MIN}
                        max={TEMP_MAX}
                        step={2}
                        value={globalParams.temperature}
                        onChange={(e) =>
                          setGlobalParams((prev) => ({
                            ...prev,
                            temperature: Math.max(
                              TEMP_MIN,
                              Math.min(
                                TEMP_MAX,
                                parseInt(e.target.value) || TEMP_MIN,
                              ),
                            ),
                          }))
                        }
                        onFocus={handleSelectAll}
                        className={INPUT_CLASS}
                      />
                    </div>

                    {/* Oxygen */}
                    <div>
                      <label className="text-white/60 text-xs block mb-1">
                        Oxygen ({OXYGEN_MIN} to {OXYGEN_MAX}%)
                      </label>
                      <input
                        type="number"
                        min={OXYGEN_MIN}
                        max={OXYGEN_MAX}
                        value={globalParams.oxygen}
                        onChange={(e) =>
                          setGlobalParams((prev) => ({
                            ...prev,
                            oxygen: Math.max(
                              OXYGEN_MIN,
                              Math.min(
                                OXYGEN_MAX,
                                parseInt(e.target.value) || OXYGEN_MIN,
                              ),
                            ),
                          }))
                        }
                        onFocus={handleSelectAll}
                        className={INPUT_CLASS}
                      />
                    </div>

                    {/* Oceans */}
                    <div>
                      <label className="text-white/60 text-xs block mb-1">
                        Oceans ({OCEANS_MIN} to {OCEANS_MAX})
                      </label>
                      <input
                        type="number"
                        min={OCEANS_MIN}
                        max={OCEANS_MAX}
                        value={globalParams.oceans}
                        onChange={(e) =>
                          setGlobalParams((prev) => ({
                            ...prev,
                            oceans: Math.max(
                              OCEANS_MIN,
                              Math.min(
                                OCEANS_MAX,
                                parseInt(e.target.value) || OCEANS_MIN,
                              ),
                            ),
                          }))
                        }
                        onFocus={handleSelectAll}
                        className={INPUT_CLASS}
                      />
                    </div>

                    {/* Generation */}
                    <div>
                      <label className="text-white/60 text-xs block mb-1">
                        Generation ({GENERATION_MIN} to {GENERATION_MAX})
                      </label>
                      <input
                        type="number"
                        min={GENERATION_MIN}
                        max={GENERATION_MAX}
                        value={generation}
                        onChange={(e) =>
                          setGeneration(
                            Math.max(
                              GENERATION_MIN,
                              Math.min(
                                GENERATION_MAX,
                                parseInt(e.target.value) || GENERATION_MIN,
                              ),
                            ),
                          )
                        }
                        onFocus={handleSelectAll}
                        className={INPUT_CLASS}
                      />
                    </div>
                  </div>
                </div>
              )}

              {/* Corporation Selection */}
              <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-4">
                <h3 className="text-white font-semibold mb-3 uppercase tracking-wide text-sm">
                  Corporation
                </h3>
                <select
                  value={selectedCorporationId}
                  onChange={(e) => setSelectedCorporationId(e.target.value)}
                  className={`${INPUT_CLASS} transition-all`}
                >
                  <option value="">Random corporation</option>
                  {availableCorporations.map((corp) => (
                    <option key={corp.id} value={corp.id}>
                      {corp.name}
                    </option>
                  ))}
                </select>
                {selectedCorporation && (
                  <p className="text-white/50 text-xs mt-2">
                    {selectedCorporation.description}
                  </p>
                )}
              </div>

              {/* Starting Cards */}
              <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-4">
                <h3 className="text-white font-semibold mb-3 uppercase tracking-wide text-sm">
                  Starting Cards
                </h3>
                <button
                  onClick={() => setShowCardSelection(!showCardSelection)}
                  className="w-full bg-black/60 border border-space-blue-400/50 rounded-lg py-2 px-3 text-white text-sm text-left hover:border-space-blue-400 transition-all"
                >
                  {selectedCardIds.length > 0
                    ? `${selectedCardIds.length} card${selectedCardIds.length !== 1 ? "s" : ""} selected`
                    : "No cards selected (click to choose)"}
                </button>

                {showCardSelection && (
                  <div className="mt-3 bg-black/30 border border-space-blue-400/30 rounded-lg p-3">
                    <input
                      type="text"
                      placeholder="Search cards..."
                      value={cardSearchTerm}
                      onChange={(e) => setCardSearchTerm(e.target.value)}
                      className="w-full bg-black/60 border border-space-blue-400/30 rounded py-1.5 px-2 text-white text-xs outline-none focus:border-space-blue-400 mb-2"
                    />

                    {selectedCardIds.length > 0 && (
                      <div className="flex flex-wrap gap-1 mb-2">
                        {selectedCardIds.map((cardId) => {
                          const card = availableCards.find(
                            (c) => c.id === cardId,
                          );
                          return (
                            <span
                              key={cardId}
                              className="inline-flex items-center gap-1 bg-space-blue-600/50 text-white text-[10px] px-1.5 py-0.5 rounded"
                            >
                              {card?.name || cardId}
                              <button
                                onClick={() => toggleCardSelection(cardId)}
                                className="text-white/70 hover:text-white"
                              >
                                ×
                              </button>
                            </span>
                          );
                        })}
                      </div>
                    )}

                    <div className="max-h-[200px] overflow-y-auto">
                      {filteredCards.slice(0, 50).map((card) => {
                        const isSelected = selectedCardIds.includes(card.id);
                        return (
                          <button
                            key={card.id}
                            onClick={() => toggleCardSelection(card.id)}
                            className={`w-full text-left px-2 py-1 text-xs rounded transition-all ${
                              isSelected
                                ? "bg-space-blue-600/30 text-white"
                                : "text-white/70 hover:bg-space-blue-400/10 hover:text-white"
                            }`}
                          >
                            <span className="font-medium">{card.name}</span>
                            <span className="text-white/40 ml-1">
                              ({card.cost})
                            </span>
                          </button>
                        );
                      })}
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* Right Column: Resources and Production */}
            <div className="space-y-6">
              {/* Terraform Rating */}
              <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-4">
                <h3 className="text-white font-semibold mb-3 uppercase tracking-wide text-sm">
                  Terraform Rating
                </h3>
                <div className="flex items-center gap-3">
                  <button
                    onClick={() =>
                      setTerraformRating((prev) => Math.max(0, prev - 1))
                    }
                    className="w-8 h-8 bg-black/60 border border-space-blue-400/50 rounded text-white hover:border-space-blue-400 transition-all"
                  >
                    -
                  </button>
                  <input
                    type="number"
                    min={0}
                    value={terraformRating}
                    onChange={(e) =>
                      setTerraformRating(
                        Math.max(0, parseInt(e.target.value) || 0),
                      )
                    }
                    onFocus={handleSelectAll}
                    className={`${INPUT_SMALL_CLASS} w-20 rounded-lg`}
                  />
                  <button
                    onClick={() => setTerraformRating((prev) => prev + 1)}
                    className="w-8 h-8 bg-black/60 border border-space-blue-400/50 rounded text-white hover:border-space-blue-400 transition-all"
                  >
                    +
                  </button>
                </div>
              </div>

              {/* Resources */}
              <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-4">
                <h3 className="text-white font-semibold mb-3 uppercase tracking-wide text-sm">
                  Resources
                </h3>
                <div className="grid grid-cols-3 gap-3">
                  {resourceTypes.map(({ key, icon, label }) => (
                    <div key={key} className="flex items-center gap-2">
                      <GameIcon iconType={icon} size="small" />
                      <input
                        type="number"
                        min={0}
                        value={resources[key]}
                        onChange={(e) =>
                          handleResourceChange(
                            key,
                            parseInt(e.target.value) || 0,
                          )
                        }
                        onFocus={handleSelectAll}
                        className={INPUT_SMALL_CLASS}
                        title={label}
                      />
                    </div>
                  ))}
                </div>
              </div>

              {/* Production */}
              <div className="bg-black/40 border border-space-blue-600/50 rounded-xl p-4">
                <h3 className="text-white font-semibold mb-3 uppercase tracking-wide text-sm">
                  Production
                </h3>
                <div className="grid grid-cols-3 gap-3">
                  {resourceTypes.map(({ key, icon, label }) => (
                    <div key={key} className="flex items-center gap-2">
                      <div className="relative">
                        <GameIcon iconType={icon} size="small" />
                        <div className="absolute -bottom-0.5 -right-0.5 w-2 h-2 bg-amber-600 rounded-full" />
                      </div>
                      <input
                        type="number"
                        min={key === "credits" ? -5 : 0}
                        value={production[key]}
                        onChange={(e) =>
                          handleProductionChange(
                            key,
                            parseInt(e.target.value) || 0,
                          )
                        }
                        onFocus={handleSelectAll}
                        className={INPUT_SMALL_CLASS}
                        title={`${label} Production`}
                      />
                    </div>
                  ))}
                </div>
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
