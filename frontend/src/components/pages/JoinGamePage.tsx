import React, { useEffect, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { apiService } from "../../services/apiService";
import { globalWebSocketManager } from "../../services/globalWebSocketManager";
import { skyboxCache } from "../../services/SkyboxCache.ts";
import LoadingOverlay from "../ui/overlay/LoadingOverlay";
import GameIcon from "../ui/display/GameIcon.tsx";
import backgroundMusicService from "../../services/backgroundMusicService.ts";

// UUIDv4 validation regex
const UUID_V4_REGEX =
  /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

const JoinGamePage: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [gameId, setGameId] = useState("");
  const [playerName, setPlayerName] = useState("");
  const [isLoadingGameValidation, setIsLoadingGameValidation] = useState(false);
  const [isLoadingJoin, setIsLoadingJoin] = useState(false);
  const [loadingStep, setLoadingStep] = useState<"game" | "environment" | null>(
    null,
  );
  const [error, setError] = useState<string | null>(null);
  const [gameValidated, setGameValidated] = useState(false);
  const [validatedGame, setValidatedGame] = useState<any>(null);
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

  // Start background music on mount
  useEffect(() => {
    void backgroundMusicService.play();

    return () => {
      // Pause instead of stop to maintain playback position
      backgroundMusicService.pause();
    };
  }, []);

  // Handle URL parameter on mount
  useEffect(() => {
    const urlParams = new URLSearchParams(location.search);
    const codeParam = urlParams.get("code");

    if (codeParam && UUID_V4_REGEX.test(codeParam)) {
      setGameId(codeParam);
      // Auto-validate the game ID from URL
      void validateGameFromUrl(codeParam);
    }
  }, [location.search]);

  const validateGameFromUrl = async (gameIdFromUrl: string) => {
    setIsLoadingGameValidation(true);
    setError(null);

    try {
      const game = await apiService.getGame(gameIdFromUrl);
      if (!game) {
        throw new Error("Game not found");
      }

      // Check if game is full
      if (
        (game.currentPlayer ? 1 : 0) + (game.otherPlayers?.length || 0) >=
        (game.settings?.maxPlayers || 4)
      ) {
        throw new Error("Game is full");
      }

      // Check if game is in a joinable state
      if (game.status !== "lobby" && game.status !== "waiting") {
        throw new Error("Game has already started");
      }

      // Game is valid and joinable
      setValidatedGame(game);
      setGameValidated(true);
    } catch (err) {
      if (err instanceof Error) {
        if (err.message.includes("404") || err.message.includes("not found")) {
          setError("Game not found. Please check the game ID.");
          return;
        } else if (err.message.includes("full")) {
          setError("This game is full. Please try another game.");
          return;
        } else if (err.message.includes("started")) {
          setError("This game has already started.");
          return;
        } else {
          setError(err.message || "Failed to find game");
          return;
        }
      }
      setError("Failed to find game");
    } finally {
      setIsLoadingGameValidation(false);
    }
  };

  const handleGameIdSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validate UUID format
    if (!gameId.trim()) {
      setError("Please enter a game ID");
      return;
    }

    if (!UUID_V4_REGEX.test(gameId.trim())) {
      setError("Please enter a valid game ID (UUID format)");
      return;
    }

    setIsLoadingGameValidation(true);
    setError(null);

    try {
      // Verify the game exists and is joinable
      const game = await apiService.getGame(gameId.trim());
      if (!game) {
        throw new Error("Game not found");
      }

      // Check if game is full
      if (
        (game.currentPlayer ? 1 : 0) + (game.otherPlayers?.length || 0) >=
        (game.settings?.maxPlayers || 4)
      ) {
        throw new Error("Game is full");
      }

      // Check if game is in a joinable state
      if (game.status !== "lobby" && game.status !== "waiting") {
        throw new Error("Game has already started");
      }

      // Game is valid and joinable
      setValidatedGame(game);
      setGameValidated(true);
    } catch (err) {
      if (err instanceof Error) {
        if (err.message.includes("404") || err.message.includes("not found")) {
          setError("Game not found. Please check the game ID.");
        } else if (err.message.includes("full")) {
          setError("This game is full. Please try another game.");
        } else if (err.message.includes("started")) {
          setError("This game has already started.");
        } else {
          setError(err.message || "Failed to find game");
        }
      } else {
        setError("Failed to find game");
      }
    } finally {
      setIsLoadingGameValidation(false);
    }
  };

  const handlePlayerNameSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validation
    if (!playerName.trim()) {
      setError("Please enter your name");
      return;
    }

    if (playerName.trim().length < 2) {
      setError("Name must be at least 2 characters long");
      return;
    }

    setIsLoadingJoin(true);
    setError(null);
    setLoadingStep("game");

    try {
      // Step 1: Connect player to the game
      const playerConnectedResult = await globalWebSocketManager.playerConnect(
        playerName.trim(),
        validatedGame.id,
      );

      if (playerConnectedResult.game) {
        // Step 2: Load 3D environment if not already loaded
        if (!skyboxReady) {
          setLoadingStep("environment");
          await skyboxCache.preload();
        }

        const gameData = {
          gameId: validatedGame.id,
          playerId: playerConnectedResult.playerId,
          playerName: playerName.trim(),
          joinedAt: new Date().toISOString(),
        };
        localStorage.setItem(
          "terraforming-mars-game",
          JSON.stringify(gameData),
        );

        navigate("/game", {
          state: {
            game: playerConnectedResult.game,
            playerId: playerConnectedResult.playerId,
            playerName: playerName.trim(),
          },
        });
      } else {
        setError("Failed to join the game. Please try again.");
      }
    } catch (err) {
      if (err instanceof Error) {
        if (err.message.includes("WebSocket")) {
          setError("Connection failed. Please try again.");
        } else {
          setError(err.message || "Failed to join game");
        }
      } else {
        setError("Failed to join game");
      }
    } finally {
      setIsLoadingJoin(false);
      setLoadingStep(null);
    }
  };

  const handleGameIdChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setGameId(e.target.value); // No transformation - keep original input
    if (error) setError(null);
  };

  const handlePlayerNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPlayerName(e.target.value);
    if (error) setError(null);
  };

  const handleGameIdKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      void handleGameIdSubmit(e as React.FormEvent);
    }
  };

  const handlePlayerNameKeyDown = (
    e: React.KeyboardEvent<HTMLInputElement>,
  ) => {
    if (e.key === "Enter") {
      void handlePlayerNameSubmit(e as React.FormEvent);
    }
  };

  const handleEditGameId = () => {
    setGameValidated(false);
    setValidatedGame(null);
    setPlayerName("");
    setError(null);
  };

  const handleBackToHome = () => {
    navigate("/");
  };

  const getLoadingMessage = () => {
    if (isLoadingGameValidation) return "Finding game...";
    if (loadingStep === "game") return "Joining game...";
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
              Join a game
            </h1>

            {!gameValidated ? (
              <form
                onSubmit={handleGameIdSubmit}
                className="max-w-[400px] mx-auto"
              >
                <div className="relative flex items-center bg-space-black-darker/95 border-2 border-space-blue-400 rounded-xl p-0 transition-all duration-200 backdrop-blur-space shadow-[0_0_20px_rgba(30,60,150,0.2)] focus-within:border-space-blue-600 focus-within:shadow-[0_0_30px_rgba(30,60,150,0.4)]">
                  <input
                    type="text"
                    value={gameId}
                    onChange={handleGameIdChange}
                    onKeyDown={handleGameIdKeyDown}
                    placeholder="Enter game ID"
                    disabled={isLoadingGameValidation}
                    className="flex-1 bg-transparent border-none py-5 px-6 text-white text-lg outline-none rounded-l-xl placeholder:text-white/50 disabled:opacity-60"
                    autoFocus
                  />
                  <button
                    type="submit"
                    disabled={isLoadingGameValidation || !gameId.trim()}
                    className="bg-transparent border-none py-4 px-5 cursor-pointer rounded-r-xl flex items-center justify-center transition-all duration-200 min-w-[80px] hover:bg-space-blue-200 hover:shadow-glow hover:translate-x-0.5 disabled:bg-transparent disabled:cursor-not-allowed disabled:transform-none"
                    title="Find Game"
                  >
                    <div className="w-5 h-7 pr-1 brightness-0 invert disabled:opacity-60">
                      <GameIcon iconType="arrow" size="small" />
                    </div>
                  </button>
                </div>

                {error && (
                  <div className="text-error-red mt-4 p-3 bg-error-red/10 border border-error-red/30 rounded-lg text-sm">
                    {error}
                  </div>
                )}
              </form>
            ) : (
              <form
                onSubmit={handlePlayerNameSubmit}
                className="max-w-[400px] mx-auto"
              >
                <div className="mb-[30px]">
                  <p className="text-white/90 text-base m-0">
                    Game found!
                    <button
                      type="button"
                      onClick={handleEditGameId}
                      className="bg-transparent border-none text-space-blue-solid cursor-pointer text-sm ml-1.5 underline transition-all duration-200 hover:text-space-blue-900 hover:text-shadow-glow-sm"
                    >
                      (edit)
                    </button>
                  </p>
                </div>

                <div className="relative flex items-center bg-space-black-darker/95 border-2 border-space-blue-400 rounded-xl p-0 transition-all duration-200 backdrop-blur-space shadow-[0_0_20px_rgba(30,60,150,0.2)] focus-within:border-space-blue-600 focus-within:shadow-[0_0_30px_rgba(30,60,150,0.4)]">
                  <input
                    type="text"
                    value={playerName}
                    onChange={handlePlayerNameChange}
                    onKeyDown={handlePlayerNameKeyDown}
                    placeholder="Enter your name"
                    disabled={isLoadingJoin}
                    className="flex-1 bg-transparent border-none py-5 px-6 text-white text-lg outline-none rounded-l-xl placeholder:text-white/50 disabled:opacity-60"
                    autoFocus
                    maxLength={50}
                  />
                  <button
                    type="submit"
                    disabled={isLoadingJoin || !playerName.trim()}
                    className="bg-transparent border-none py-4 px-5 cursor-pointer rounded-r-xl flex items-center justify-center transition-all duration-200 min-w-[80px] hover:bg-space-blue-200 hover:shadow-glow hover:translate-x-0.5 disabled:bg-transparent disabled:cursor-not-allowed disabled:transform-none"
                    title="Join Game"
                  >
                    <div className="w-5 h-7 pr-1 brightness-0 invert disabled:opacity-60">
                      <GameIcon iconType="arrow" size="small" />
                    </div>
                  </button>
                </div>

                {error && (
                  <div className="text-error-red mt-4 p-3 bg-error-red/10 border border-error-red/30 rounded-lg text-sm">
                    {error}
                  </div>
                )}
              </form>
            )}
          </div>
        </div>
      </div>

      <LoadingOverlay
        isLoading={isLoadingGameValidation || isLoadingJoin}
        message={getLoadingMessage()}
      />
    </div>
  );
};

export default JoinGamePage;
