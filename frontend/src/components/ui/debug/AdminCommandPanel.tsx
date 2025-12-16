import React, { useState, useEffect } from "react";
import { globalWebSocketManager } from "../../../services/globalWebSocketManager.ts";
import {
  GameDto,
  AdminCommandRequest,
  AdminCommandTypeGiveCard,
  AdminCommandTypeSetPhase,
  AdminCommandTypeSetResources,
  AdminCommandTypeSetProduction,
  AdminCommandTypeSetGlobalParams,
  AdminCommandTypeStartTileSelection,
  GiveCardAdminCommand,
  SetPhaseAdminCommand,
  SetResourcesAdminCommand,
  SetProductionAdminCommand,
  SetGlobalParamsAdminCommand,
  StartTileSelectionAdminCommand,
  GamePhaseWaitingForGameStart,
  GamePhaseStartingCardSelection,
  GamePhaseAction,
  GamePhaseProductionAndCardDraw,
  GamePhaseComplete,
} from "../../../types/generated/api-types.ts";

interface AdminCommandPanelProps {
  gameState: GameDto;
  onClose?: () => void;
}

const AdminCommandPanel: React.FC<AdminCommandPanelProps> = ({
  gameState,
  onClose,
}) => {
  const [selectedCommand, setSelectedCommand] = useState<string>("");
  const [validationErrors, setValidationErrors] = useState<
    Record<string, boolean>
  >({});

  // Shared styling functions

  const getInputStyle = (
    hasError: boolean = false,
    disabled: boolean = false,
  ) => ({
    width: "100%",
    padding: "6px 10px",
    marginTop: "2px",
    background: disabled ? "rgba(0, 0, 0, 0.4)" : "rgba(0, 0, 0, 0.8)",
    border: hasError
      ? "1px solid #ff4444"
      : "1px solid rgba(155, 89, 182, 0.3)",
    borderRadius: "4px",
    color: disabled ? "#666" : "white",
    fontSize: "12px",
    outline: "none",
    boxShadow: hasError ? "0 0 0 2px rgba(255, 68, 68, 0.2)" : "none",
    cursor: disabled ? "not-allowed" : "text",
  });

  const getSelectStyle = (hasError: boolean = false, customWidth?: string) => ({
    width: customWidth || "200px",
    maxWidth: "100%",
    padding: "6px 10px",
    marginTop: "2px",
    background: "rgba(0, 0, 0, 0.8)",
    border: hasError
      ? "1px solid #ff4444"
      : "1px solid rgba(155, 89, 182, 0.3)",
    borderRadius: "4px",
    color: "white",
    fontSize: "12px",
    outline: "none",
    cursor: "pointer",
    appearance: "none" as const,
    backgroundImage: `url("data:image/svg+xml;charset=US-ASCII,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 4 5'><path fill='%23abb2bf' d='M2 0L0 2h4zm0 5L0 3h4z'/></svg>")`,
    backgroundRepeat: "no-repeat",
    backgroundPosition: "right 6px center",
    backgroundSize: "10px",
    paddingRight: "28px",
    boxShadow: hasError ? "0 0 0 2px rgba(255, 68, 68, 0.2)" : "none",
  });

  const buttonStyle = {
    padding: "8px 16px",
    background:
      "linear-gradient(135deg, rgba(155, 89, 182, 0.8), rgba(155, 89, 182, 0.6))",
    border: "1px solid rgba(155, 89, 182, 0.5)",
    borderRadius: "6px",
    color: "white",
    fontSize: "12px",
    cursor: "pointer",
    transition: "all 0.2s ease",
    fontWeight: "500" as const,
  };
  const [giveCardForm, setGiveCardForm] = useState({
    playerId: "",
    cardId: "",
  });
  const [setPhaseForm, setSetPhaseForm] = useState({ phase: "" });
  const [resourcesForm, setResourcesForm] = useState({
    playerId: "",
    credit: "",
    steel: "",
    titanium: "",
    plant: "",
    energy: "",
    heat: "",
  });
  const [productionForm, setProductionForm] = useState({
    playerId: "",
    credit: "",
    steel: "",
    titanium: "",
    plant: "",
    energy: "",
    heat: "",
  });
  const [globalParamsForm, setGlobalParamsForm] = useState({
    temperature: gameState.globalParameters.temperature,
    oxygen: gameState.globalParameters.oxygen,
    oceans: gameState.globalParameters.oceans,
  });
  const [tileSelectionForm, setTileSelectionForm] = useState({
    playerId: "",
    tileType: "",
  });
  const [setCorporationForm, setSetCorporationForm] = useState({
    playerId: "",
    corporationId: "",
  });

  const allPlayers = [gameState.currentPlayer, ...gameState.otherPlayers];

  // Update forms when player selection changes
  useEffect(() => {
    if (resourcesForm.playerId) {
      const selectedPlayer = allPlayers.find(
        (p) => p.id === resourcesForm.playerId,
      );
      if (selectedPlayer && selectedPlayer.resources) {
        setResourcesForm((prev) => ({
          ...prev,
          credit: (selectedPlayer.resources.credits || 0).toString(),
          steel: (selectedPlayer.resources.steel || 0).toString(),
          titanium: (selectedPlayer.resources.titanium || 0).toString(),
          plant: (selectedPlayer.resources.plants || 0).toString(),
          energy: (selectedPlayer.resources.energy || 0).toString(),
          heat: (selectedPlayer.resources.heat || 0).toString(),
        }));
      }
    } else {
      // Reset all fields when no player is selected
      setResourcesForm((prev) => ({
        ...prev,
        credit: "",
        steel: "",
        titanium: "",
        plant: "",
        energy: "",
        heat: "",
      }));
    }
  }, [resourcesForm.playerId]);

  useEffect(() => {
    if (productionForm.playerId) {
      const selectedPlayer = allPlayers.find(
        (p) => p.id === productionForm.playerId,
      );
      if (selectedPlayer && selectedPlayer.production) {
        setProductionForm((prev) => ({
          ...prev,
          credit: (selectedPlayer.production.credits || 0).toString(),
          steel: (selectedPlayer.production.steel || 0).toString(),
          titanium: (selectedPlayer.production.titanium || 0).toString(),
          plant: (selectedPlayer.production.plants || 0).toString(),
          energy: (selectedPlayer.production.energy || 0).toString(),
          heat: (selectedPlayer.production.heat || 0).toString(),
        }));
      }
    } else {
      // Reset all fields when no player is selected
      setProductionForm((prev) => ({
        ...prev,
        credit: "",
        steel: "",
        titanium: "",
        plant: "",
        energy: "",
        heat: "",
      }));
    }
  }, [productionForm.playerId]);

  // Update global parameters form when game state changes
  useEffect(() => {
    setGlobalParamsForm({
      temperature: gameState.globalParameters.temperature,
      oxygen: gameState.globalParameters.oxygen,
      oceans: gameState.globalParameters.oceans,
    });
  }, [gameState.globalParameters]);

  // Keyboard event handlers for Enter key support
  const handleGiveCardKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      void handleGiveCard();
    }
  };

  const handleSetPhaseKeyDown = (e: React.KeyboardEvent<HTMLSelectElement>) => {
    if (e.key === "Enter") {
      void handleSetPhase();
    }
  };

  const handleResourcesKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      void handleSetResources();
    }
  };

  const handleProductionKeyDown = (
    e: React.KeyboardEvent<HTMLInputElement>,
  ) => {
    if (e.key === "Enter") {
      void handleSetProduction();
    }
  };

  const handleGlobalParamsKeyDown = (
    e: React.KeyboardEvent<HTMLInputElement>,
  ) => {
    if (e.key === "Enter") {
      void handleSetGlobalParams();
    }
  };

  const sendAdminCommand = async (commandType: string, payload: any) => {
    const adminRequest: AdminCommandRequest = {
      commandType: commandType as any,
      payload: payload,
    };

    try {
      await globalWebSocketManager.sendAdminCommand(adminRequest);
    } catch (error) {
      console.error("❌ Failed to send admin command:", error);
    }
  };

  const handleGiveCard = async () => {
    const errors: Record<string, boolean> = {};

    if (!giveCardForm.playerId) errors.giveCardPlayerId = true;
    if (!giveCardForm.cardId) errors.giveCardCardId = true;

    setValidationErrors(errors);

    if (Object.keys(errors).length > 0) {
      // Clear errors after 3 seconds
      setTimeout(() => setValidationErrors({}), 3000);
      return;
    }

    const command: GiveCardAdminCommand = {
      playerId: giveCardForm.playerId,
      cardId: giveCardForm.cardId,
    };

    await sendAdminCommand(AdminCommandTypeGiveCard, command);
    // Keep player selected, only clear card ID for next card
    setGiveCardForm({ ...giveCardForm, cardId: "" });
  };

  const handleSetPhase = async () => {
    const errors: Record<string, boolean> = {};

    if (!setPhaseForm.phase) errors.setPhase = true;

    setValidationErrors(errors);

    if (Object.keys(errors).length > 0) {
      setTimeout(() => setValidationErrors({}), 3000);
      return;
    }

    const command: SetPhaseAdminCommand = {
      phase: setPhaseForm.phase,
    };

    await sendAdminCommand(AdminCommandTypeSetPhase, command);
  };

  const parseStringToNumber = (value: string): number => {
    if (value === "" || value === undefined || value === null) {
      return 0;
    }
    const parsed = parseInt(value, 10);
    return isNaN(parsed) ? 0 : Math.max(0, parsed); // Ensure non-negative
  };

  const handleSetResources = async () => {
    const errors: Record<string, boolean> = {};

    if (!resourcesForm.playerId) errors.setResourcesPlayerId = true;

    // Validate that all resource values are valid numbers
    const resourceFields = [
      "credit",
      "steel",
      "titanium",
      "plant",
      "energy",
      "heat",
    ];
    for (const field of resourceFields) {
      const value = resourcesForm[
        field as keyof typeof resourcesForm
      ] as string;
      if (
        value !== "" &&
        (isNaN(parseInt(value, 10)) || parseInt(value, 10) < 0)
      ) {
        errors[
          `setResources${field.charAt(0).toUpperCase() + field.slice(1)}`
        ] = true;
      }
    }

    setValidationErrors(errors);

    if (Object.keys(errors).length > 0) {
      setTimeout(() => setValidationErrors({}), 3000);
      return;
    }

    const command: SetResourcesAdminCommand = {
      playerId: resourcesForm.playerId,
      resources: {
        credits: parseStringToNumber(resourcesForm.credit),
        steel: parseStringToNumber(resourcesForm.steel),
        titanium: parseStringToNumber(resourcesForm.titanium),
        plants: parseStringToNumber(resourcesForm.plant),
        energy: parseStringToNumber(resourcesForm.energy),
        heat: parseStringToNumber(resourcesForm.heat),
      },
    };

    await sendAdminCommand(AdminCommandTypeSetResources, command);
  };

  const handleSetProduction = async () => {
    const errors: Record<string, boolean> = {};

    if (!productionForm.playerId) errors.setProductionPlayerId = true;

    // Validate that all production values are valid numbers
    const productionFields = [
      "credit",
      "steel",
      "titanium",
      "plant",
      "energy",
      "heat",
    ];
    for (const field of productionFields) {
      const value = productionForm[
        field as keyof typeof productionForm
      ] as string;
      if (
        value !== "" &&
        (isNaN(parseInt(value, 10)) || parseInt(value, 10) < 0)
      ) {
        errors[
          `setProduction${field.charAt(0).toUpperCase() + field.slice(1)}`
        ] = true;
      }
    }

    setValidationErrors(errors);

    if (Object.keys(errors).length > 0) {
      setTimeout(() => setValidationErrors({}), 3000);
      return;
    }

    const command: SetProductionAdminCommand = {
      playerId: productionForm.playerId,
      production: {
        credits: parseStringToNumber(productionForm.credit),
        steel: parseStringToNumber(productionForm.steel),
        titanium: parseStringToNumber(productionForm.titanium),
        plants: parseStringToNumber(productionForm.plant),
        energy: parseStringToNumber(productionForm.energy),
        heat: parseStringToNumber(productionForm.heat),
      },
    };

    await sendAdminCommand(AdminCommandTypeSetProduction, command);
  };

  const handleSetGlobalParams = async () => {
    const command: SetGlobalParamsAdminCommand = {
      globalParameters: {
        temperature: globalParamsForm.temperature,
        oxygen: globalParamsForm.oxygen,
        oceans: globalParamsForm.oceans,
      },
    };

    await sendAdminCommand(AdminCommandTypeSetGlobalParams, command);
  };

  const handleStartTileSelection = async () => {
    const errors: Record<string, boolean> = {};

    if (!tileSelectionForm.playerId) errors.tileSelectionPlayerId = true;
    if (!tileSelectionForm.tileType) errors.tileSelectionTileType = true;

    setValidationErrors(errors);

    if (Object.keys(errors).length > 0) {
      setTimeout(() => setValidationErrors({}), 3000);
      return;
    }

    const command: StartTileSelectionAdminCommand = {
      playerId: tileSelectionForm.playerId,
      tileType: tileSelectionForm.tileType,
    };

    await sendAdminCommand(AdminCommandTypeStartTileSelection, command);
    setTileSelectionForm({ playerId: "", tileType: "" });

    // Close the admin panel after starting tile selection
    if (onClose) {
      onClose();
    }
  };

  const handleSetCorporation = async () => {
    const errors: Record<string, boolean> = {};

    if (!setCorporationForm.playerId) errors.setCorporationPlayerId = true;
    if (!setCorporationForm.corporationId)
      errors.setCorporationCorporationId = true;

    setValidationErrors(errors);

    if (Object.keys(errors).length > 0) {
      setTimeout(() => setValidationErrors({}), 3000);
      return;
    }

    const command = {
      playerId: setCorporationForm.playerId,
      corporationId: setCorporationForm.corporationId,
    };

    await sendAdminCommand("set-corporation" as any, command);
    setSetCorporationForm({ playerId: "", corporationId: "" });
  };

  const commandOptions = [
    { value: "give-card", label: "Give Card to Player" },
    { value: "set-phase", label: "Set Game Phase" },
    { value: "set-resources", label: "Set Player Resources" },
    { value: "set-production", label: "Set Player Production" },
    { value: "set-global-params", label: "Set Global Parameters" },
    { value: "start-tile-selection", label: "Start Tile Selection (Demo)" },
    { value: "set-corporation", label: "Set Player Corporation" },
  ];

  const phaseOptions = [
    { value: GamePhaseWaitingForGameStart, label: "Waiting for Game Start" },
    { value: GamePhaseStartingCardSelection, label: "Starting Card Selection" },
    { value: GamePhaseAction, label: "Action Phase" },
    {
      value: GamePhaseProductionAndCardDraw,
      label: "Production and Card Draw",
    },
    { value: GamePhaseComplete, label: "Game Complete" },
  ];

  return (
    <div
      className="debug-content-area"
      style={{
        flex: 1,
        overflow: "auto",
        background: "rgba(0, 0, 0, 0.5)",
        padding: "12px",
        borderRadius: "4px",
        border: "1px solid #222",
      }}
    >
      <div style={{ marginBottom: "16px" }}>
        <label
          style={{ color: "#9b59b6", fontSize: "12px", fontWeight: "bold" }}
        >
          Select Admin Command:
        </label>
        <select
          value={selectedCommand}
          onChange={(e) => setSelectedCommand(e.target.value)}
          style={{
            ...getSelectStyle(false, "100%"),
            fontSize: "13px",
            padding: "8px 12px",
            borderRadius: "6px",
            backgroundPosition: "right 8px center",
            backgroundSize: "12px",
            paddingRight: "32px",
          }}
        >
          <option value="">Choose a command...</option>
          {commandOptions.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      </div>

      {selectedCommand === "give-card" && (
        <div style={{ marginBottom: "16px" }}>
          <h4 style={{ color: "#9b59b6", margin: "0 0 12px 0" }}>
            Give Card to Player
          </h4>
          <div style={{ marginBottom: "8px" }}>
            <select
              value={giveCardForm.playerId}
              onChange={(e) =>
                setGiveCardForm({ ...giveCardForm, playerId: e.target.value })
              }
              style={getSelectStyle(validationErrors.giveCardPlayerId)}
            >
              <option value="">Select player...</option>
              {allPlayers.map((player) => (
                <option key={player.id} value={player.id}>
                  {player.name}
                </option>
              ))}
            </select>
          </div>
          <div style={{ marginBottom: "8px" }}>
            <input
              type="text"
              placeholder="Enter card ID"
              value={giveCardForm.cardId}
              onChange={(e) =>
                setGiveCardForm({ ...giveCardForm, cardId: e.target.value })
              }
              onKeyDown={handleGiveCardKeyDown}
              style={{
                ...getInputStyle(validationErrors.giveCardCardId),
                width: "200px",
                maxWidth: "100%",
              }}
            />
          </div>
          <button onClick={handleGiveCard} style={buttonStyle}>
            Give Card
          </button>
        </div>
      )}

      {selectedCommand === "set-phase" && (
        <div style={{ marginBottom: "16px" }}>
          <h4 style={{ color: "#9b59b6", margin: "0 0 12px 0" }}>
            Set Game Phase
          </h4>
          <div style={{ marginBottom: "8px" }}>
            <select
              value={setPhaseForm.phase}
              onChange={(e) => setSetPhaseForm({ phase: e.target.value })}
              onKeyDown={handleSetPhaseKeyDown}
              style={getSelectStyle(validationErrors.setPhase)}
            >
              <option value="">Select phase...</option>
              {phaseOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>
          <button onClick={handleSetPhase} style={buttonStyle}>
            Set Phase
          </button>
        </div>
      )}

      {selectedCommand === "set-resources" && (
        <div style={{ marginBottom: "16px" }}>
          <h4 style={{ color: "#9b59b6", margin: "0 0 12px 0" }}>
            Set Player Resources
          </h4>
          <div style={{ marginBottom: "8px" }}>
            <select
              value={resourcesForm.playerId}
              onChange={(e) =>
                setResourcesForm({ ...resourcesForm, playerId: e.target.value })
              }
              style={getSelectStyle(validationErrors.setResourcesPlayerId)}
            >
              <option value="">Select player...</option>
              {allPlayers.map((player) => (
                <option key={player.id} value={player.id}>
                  {player.name}
                </option>
              ))}
            </select>
          </div>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: "8px 16px",
              marginBottom: "8px",
              padding: "0 4px",
            }}
          >
            {["credit", "steel", "titanium", "plant", "energy", "heat"].map(
              (resource) => (
                <div key={resource} style={{ minWidth: 0 }}>
                  <label
                    style={{
                      color: "#abb2bf",
                      fontSize: "11px",
                      textTransform: "capitalize",
                      display: "block",
                      marginBottom: "4px",
                    }}
                  >
                    {resource}:
                  </label>
                  <input
                    type="text"
                    value={
                      resourcesForm[
                        resource as keyof typeof resourcesForm
                      ] as string
                    }
                    onChange={(e) =>
                      setResourcesForm({
                        ...resourcesForm,
                        [resource]: e.target.value,
                      })
                    }
                    onKeyDown={handleResourcesKeyDown}
                    disabled={!resourcesForm.playerId}
                    style={{
                      ...getInputStyle(false, !resourcesForm.playerId),
                      fontSize: "11px",
                      width: "100%",
                      minWidth: 0,
                      boxSizing: "border-box" as const,
                    }}
                  />
                </div>
              ),
            )}
          </div>
          <button onClick={handleSetResources} style={buttonStyle}>
            Set Resources
          </button>
        </div>
      )}

      {selectedCommand === "set-production" && (
        <div style={{ marginBottom: "16px" }}>
          <h4 style={{ color: "#9b59b6", margin: "0 0 12px 0" }}>
            Set Player Production
          </h4>
          <div style={{ marginBottom: "8px" }}>
            <select
              value={productionForm.playerId}
              onChange={(e) =>
                setProductionForm({
                  ...productionForm,
                  playerId: e.target.value,
                })
              }
              style={getSelectStyle(validationErrors.setProductionPlayerId)}
            >
              <option value="">Select player...</option>
              {allPlayers.map((player) => (
                <option key={player.id} value={player.id}>
                  {player.name}
                </option>
              ))}
            </select>
          </div>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: "8px 16px",
              marginBottom: "8px",
              padding: "0 4px",
            }}
          >
            {["credit", "steel", "titanium", "plant", "energy", "heat"].map(
              (resource) => (
                <div key={resource} style={{ minWidth: 0 }}>
                  <label
                    style={{
                      color: "#abb2bf",
                      fontSize: "11px",
                      textTransform: "capitalize",
                      display: "block",
                      marginBottom: "4px",
                    }}
                  >
                    {resource}:
                  </label>
                  <input
                    type="text"
                    value={
                      productionForm[
                        resource as keyof typeof productionForm
                      ] as string
                    }
                    onChange={(e) =>
                      setProductionForm({
                        ...productionForm,
                        [resource]: e.target.value,
                      })
                    }
                    onKeyDown={handleProductionKeyDown}
                    disabled={!productionForm.playerId}
                    style={{
                      ...getInputStyle(false, !productionForm.playerId),
                      fontSize: "11px",
                      width: "100%",
                      minWidth: 0,
                      boxSizing: "border-box" as const,
                    }}
                  />
                </div>
              ),
            )}
          </div>
          <button onClick={handleSetProduction} style={buttonStyle}>
            Set Production
          </button>
        </div>
      )}

      {selectedCommand === "set-global-params" && (
        <div style={{ marginBottom: "16px" }}>
          <h4 style={{ color: "#9b59b6", margin: "0 0 12px 0" }}>
            Set Global Parameters
          </h4>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr",
              gap: "8px",
              marginBottom: "8px",
            }}
          >
            <div>
              <label
                style={{
                  color: "#abb2bf",
                  fontSize: "12px",
                  marginRight: "8px",
                }}
              >
                Temperature (-30 to +8°C):
              </label>
              <input
                type="number"
                min="-30"
                max="8"
                value={globalParamsForm.temperature}
                onChange={(e) =>
                  setGlobalParamsForm({
                    ...globalParamsForm,
                    temperature: parseInt(e.target.value) || -30,
                  })
                }
                onKeyDown={handleGlobalParamsKeyDown}
                style={{
                  ...getInputStyle(),
                  width: "120px",
                  maxWidth: "100%",
                }}
              />
            </div>
            <div>
              <label
                style={{
                  color: "#abb2bf",
                  fontSize: "12px",
                  marginRight: "8px",
                }}
              >
                Oxygen (0-14%):
              </label>
              <input
                type="number"
                min="0"
                max="14"
                value={globalParamsForm.oxygen}
                onChange={(e) =>
                  setGlobalParamsForm({
                    ...globalParamsForm,
                    oxygen: parseInt(e.target.value) || 0,
                  })
                }
                onKeyDown={handleGlobalParamsKeyDown}
                style={{
                  ...getInputStyle(),
                  width: "120px",
                  maxWidth: "100%",
                }}
              />
            </div>
            <div>
              <label
                style={{
                  color: "#abb2bf",
                  fontSize: "12px",
                  marginRight: "8px",
                }}
              >
                Oceans (0-9):
              </label>
              <input
                type="number"
                min="0"
                max="9"
                value={globalParamsForm.oceans}
                onChange={(e) =>
                  setGlobalParamsForm({
                    ...globalParamsForm,
                    oceans: parseInt(e.target.value) || 0,
                  })
                }
                onKeyDown={handleGlobalParamsKeyDown}
                style={{
                  ...getInputStyle(),
                  width: "120px",
                  maxWidth: "100%",
                }}
              />
            </div>
          </div>
          <button onClick={handleSetGlobalParams} style={buttonStyle}>
            Set Global Parameters
          </button>
        </div>
      )}

      {selectedCommand === "start-tile-selection" && (
        <div style={{ marginBottom: "16px" }}>
          <h4 style={{ color: "#9b59b6", margin: "0 0 12px 0" }}>
            Start Tile Selection (Demo)
          </h4>
          <div style={{ marginBottom: "8px" }}>
            <select
              value={tileSelectionForm.playerId}
              onChange={(e) =>
                setTileSelectionForm({
                  ...tileSelectionForm,
                  playerId: e.target.value,
                })
              }
              style={getSelectStyle(validationErrors.tileSelectionPlayerId)}
            >
              <option value="">Select player...</option>
              {allPlayers.map((player) => (
                <option key={player.id} value={player.id}>
                  {player.name}
                </option>
              ))}
            </select>
          </div>
          <div style={{ marginBottom: "8px" }}>
            <select
              value={tileSelectionForm.tileType}
              onChange={(e) =>
                setTileSelectionForm({
                  ...tileSelectionForm,
                  tileType: e.target.value,
                })
              }
              style={getSelectStyle(validationErrors.tileSelectionTileType)}
            >
              <option value="">Select tile type...</option>
              <option value="city">City</option>
              <option value="greenery">Greenery</option>
              <option value="ocean">Ocean</option>
            </select>
          </div>
          <button onClick={handleStartTileSelection} style={buttonStyle}>
            Start Tile Selection
          </button>
          <div
            style={{
              marginTop: "8px",
              padding: "8px",
              background: "rgba(255, 193, 7, 0.1)",
              border: "1px solid rgba(255, 193, 7, 0.3)",
              borderRadius: "4px",
              fontSize: "11px",
              color: "#ffc107",
            }}
          >
            <strong>Demo:</strong> This will trigger tile selection for the
            chosen player. Available hexes will be highlighted on the Mars
            board. Click a highlighted hex to complete the tile placement.
          </div>
        </div>
      )}

      {selectedCommand === "set-corporation" && (
        <div style={{ marginBottom: "16px" }}>
          <h4 style={{ color: "#9b59b6", margin: "0 0 12px 0" }}>
            Set Player Corporation
          </h4>
          <div style={{ marginBottom: "8px" }}>
            <select
              value={setCorporationForm.playerId}
              onChange={(e) =>
                setSetCorporationForm({
                  ...setCorporationForm,
                  playerId: e.target.value,
                })
              }
              style={getSelectStyle(validationErrors.setCorporationPlayerId)}
            >
              <option value="">Select player...</option>
              {allPlayers.map((player) => (
                <option key={player.id} value={player.id}>
                  {player.name}
                </option>
              ))}
            </select>
          </div>
          <div style={{ marginBottom: "8px" }}>
            <input
              type="text"
              placeholder="Enter corporation ID (e.g., B03 for Helion)"
              value={setCorporationForm.corporationId}
              onChange={(e) =>
                setSetCorporationForm({
                  ...setCorporationForm,
                  corporationId: e.target.value,
                })
              }
              style={{
                ...getInputStyle(validationErrors.setCorporationCorporationId),
                width: "100%",
              }}
            />
          </div>
          <button onClick={handleSetCorporation} style={buttonStyle}>
            Set Corporation
          </button>
          <div
            style={{
              marginTop: "8px",
              padding: "8px",
              background: "rgba(255, 193, 7, 0.1)",
              border: "1px solid rgba(255, 193, 7, 0.3)",
              borderRadius: "4px",
              fontSize: "11px",
              color: "#ffc107",
            }}
          >
            ⚠️ <strong>Warning:</strong> This will clear any pending selections
            and apply corporation bonuses (resources, production, payment
            substitutes, effects, and actions).
          </div>
        </div>
      )}

      {!selectedCommand && (
        <div
          style={{
            color: "#666",
            textAlign: "center",
            padding: "20px",
            fontSize: "12px",
          }}
        >
          Select an admin command above to get started.
          <br />
          <br />
          Available commands:
          <ul style={{ textAlign: "left", marginTop: "12px" }}>
            <li>Give cards to players</li>
            <li>Change game phase</li>
            <li>Set player resources</li>
            <li>Set player production</li>
            <li>Modify global parameters</li>
            <li>Start tile selection (demo)</li>
            <li>Set player corporation</li>
          </ul>
        </div>
      )}
    </div>
  );
};

export default AdminCommandPanel;
