import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { apiService } from "../../services/apiService";
import { globalWebSocketManager } from "../../services/globalWebSocketManager";
import { GameSettingsDto } from "../../types/generated/api-types.ts";
import { skyboxCache } from "../../services/SkyboxCache.ts";
import LoadingOverlay from "../ui/overlay/LoadingOverlay";

const CreateGamePage: React.FC = () => {
  const navigate = useNavigate();
  const [playerName, setPlayerName] = useState("");
  const [developmentMode, setDevelopmentMode] = useState(true);
  const [isLoading, setIsLoading] = useState(false);
  const [loadingStep, setLoadingStep] = useState<"game" | "environment" | null>(
    null,
  );
  const [error, setError] = useState<string | null>(null);
  const [skyboxReady, setSkyboxReady] = useState(false);
  const [isFadedIn, setIsFadedIn] = useState(false);

  // Check if skybox is already loaded on component mount
  useEffect(() => {
    if (skyboxCache.isReady()) {
      setSkyboxReady(true);
    }
    // Trigger fade in animation
    setTimeout(() => {
      setIsFadedIn(true);
    }, 10);
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!playerName.trim()) {
      setError("Please enter your name");
      return;
    }

    if (playerName.trim().length < 2) {
      setError("Name must be at least 2 characters long");
      return;
    }

    setIsLoading(true);
    setError(null);
    setLoadingStep("game");

    try {
      // Step 1: Create game
      const gameSettings: GameSettingsDto = {
        maxPlayers: 4, // Default max players
        developmentMode: developmentMode,
      };

      const game = await apiService.createGame(gameSettings);

      // Connect player to the game via WebSocket
      const playerConnectedResult = await globalWebSocketManager.playerConnect(
        playerName.trim(),
        game.id,
      );

      if (playerConnectedResult.game) {
        // Step 2: Load 3D environment if not already loaded
        if (!skyboxReady) {
          setLoadingStep("environment");
          await skyboxCache.preload();
        }

        const gameData = {
          gameId: playerConnectedResult.game.id,
          playerId: playerConnectedResult.playerId,
          playerName: playerName.trim(),
          createdAt: new Date().toISOString(),
        };
        localStorage.setItem(
          "terraforming-mars-game",
          JSON.stringify(gameData),
        );

        // Navigate to the main game interface with the complete game state
        navigate("/game", {
          state: {
            game: playerConnectedResult.game,
            playerId: playerConnectedResult.playerId,
            playerName: playerName.trim(),
          },
        });
      } else {
        setError("Failed to connect to the game");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create game");
    } finally {
      setIsLoading(false);
      setLoadingStep(null);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPlayerName(e.target.value);
    if (error) setError(null); // Clear error when user starts typing
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      void handleSubmit(e as React.FormEvent);
    }
  };

  const handleBackToHome = () => {
    navigate("/");
  };

  const getLoadingMessage = () => {
    if (loadingStep === "game") return "Creating game...";
    if (loadingStep === "environment") return "Loading 3D environment...";
    return "Loading...";
  };

  return (
    <div
      className={`bg-transparent text-white min-h-screen flex items-center justify-center font-sans relative z-10 transition-opacity duration-300 ease-in ${isFadedIn ? "opacity-100" : "opacity-0"}`}
    >
      <div className="relative z-[1] flex items-center justify-center w-full min-h-screen">
        <button
          onClick={handleBackToHome}
          className="fixed top-[30px] left-[30px] bg-space-black-darker/80 border-2 border-space-blue-400 rounded-lg py-3 px-5 text-white cursor-pointer transition-all duration-300 text-sm backdrop-blur-space z-[100] hover:bg-space-black-darker/90 hover:border-space-blue-800 hover:shadow-glow hover:-translate-y-0.5"
        >
          ← Back to Home
        </button>
        <div className="max-w-[600px] w-full px-5 py-10">
          <div className="text-center">
            <h1 className="font-orbitron text-[42px] text-white mb-[60px] text-shadow-glow font-bold tracking-wider">
              Create a new game
            </h1>

            <form onSubmit={handleSubmit} className="max-w-[400px] mx-auto">
              <div className="relative flex items-center bg-space-black-darker/80 border-2 border-space-blue-400 rounded-xl p-0 transition-all duration-200 backdrop-blur-space focus-within:border-space-blue-800 focus-within:shadow-glow">
                <input
                  type="text"
                  value={playerName}
                  onChange={handleInputChange}
                  onKeyDown={handleKeyDown}
                  placeholder="Enter your name"
                  disabled={isLoading}
                  className="flex-1 bg-transparent border-none py-5 px-6 text-white text-lg outline-none rounded-l-xl placeholder:text-white/50 disabled:opacity-60"
                  autoFocus
                  maxLength={50}
                />

                <button
                  type="submit"
                  disabled={isLoading || !playerName.trim()}
                  className="bg-gradient-to-br from-space-blue-800 to-space-blue-600 border-none py-5 px-6 cursor-pointer rounded-r-xl flex items-center justify-center transition-all duration-200 min-w-[80px] hover:from-space-blue-900 hover:to-space-blue-800 hover:shadow-glow hover:translate-x-0.5 disabled:bg-gray-700/50 disabled:cursor-not-allowed disabled:transform-none"
                  title="Connect"
                >
                  <img
                    src="/assets/misc/arrow.png"
                    alt="Connect"
                    className="w-6 h-6 brightness-0 invert disabled:opacity-60"
                  />
                </button>
              </div>

              <div className="mt-5 text-center flex justify-center">
                <label className="flex items-center gap-3 cursor-pointer py-2 transition-all duration-200">
                  <input
                    type="checkbox"
                    checked={developmentMode}
                    onChange={(e) => setDevelopmentMode(e.target.checked)}
                    disabled={isLoading}
                    className="w-[18px] h-[18px] accent-space-blue-solid cursor-pointer m-0 disabled:opacity-60 disabled:cursor-not-allowed"
                  />
                  <span className="text-white text-base font-medium leading-none m-0 flex items-center gap-2">
                    Development Mode
                    <div className="relative inline-block group">
                      <span className="text-space-blue-solid text-base cursor-help w-[18px] h-[18px] flex items-center justify-center rounded-full bg-space-blue-100 border border-space-blue-400 transition-all duration-200 shadow-[0_0_10px_rgba(30,60,150,0.2)] group-hover:bg-space-blue-200 group-hover:shadow-[0_0_15px_rgba(30,60,150,0.4)]">
                        ⓘ
                      </span>
                      <div className="invisible opacity-0 w-[280px] bg-space-black/[0.98] text-white text-left rounded-lg p-3 absolute z-[1000] bottom-[125%] right-0 text-[13px] leading-normal border border-space-blue-400 shadow-glow transition-all duration-300 group-hover:visible group-hover:opacity-100 after:content-[''] after:absolute after:top-full after:right-3 after:border-8 after:border-solid after:border-t-space-black/[0.98] after:border-r-transparent after:border-b-transparent after:border-l-transparent">
                        Enable admin commands for debugging and testing. Allows
                        you to give cards to players, modify
                        resources/production, change game phases, and adjust
                        global parameters through the debug panel.
                      </div>
                    </div>
                  </span>
                </label>
              </div>

              {error && (
                <div className="text-error-red mt-4 p-3 bg-error-red/10 border border-error-red/30 rounded-lg text-sm">
                  {error}
                </div>
              )}
            </form>
          </div>
        </div>
      </div>

      <LoadingOverlay isLoading={isLoading} message={getLoadingMessage()} />
    </div>
  );
};

export default CreateGamePage;
