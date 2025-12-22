import React, { useEffect, useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { apiService } from "../../services/apiService";
import { globalWebSocketManager } from "../../services/globalWebSocketManager.ts";
import { useSpaceBackground } from "../../contexts/SpaceBackgroundContext.tsx";
import { GameDto } from "../../types/generated/api-types.ts";
import { getCorporationLogo } from "../../utils/corporationLogos.tsx";
import { clearGameSession } from "../../utils/sessionStorage.ts";

const GameLandingPage: React.FC = () => {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [isFadingOut, setIsFadingOut] = useState(false);
  const [isCreatingDemo, setIsCreatingDemo] = useState(false);
  const { preloadSkybox } = useSpaceBackground();
  const [savedGameData, setSavedGameData] = useState<{
    game: GameDto;
    playerId: string;
    playerName: string;
  } | null>(null);

  useEffect(() => {
    const checkExistingGame = async () => {
      try {
        // Preload skybox in parallel with game check
        void preloadSkybox();

        // Check localStorage for existing game
        const savedGameDataString = localStorage.getItem(
          "terraforming-mars-game",
        );
        if (savedGameDataString) {
          const { gameId, playerId, playerName } =
            JSON.parse(savedGameDataString);

          if (gameId && playerId) {
            // Try to get the current game state from server with player ID for personalized view
            const game = await apiService.getGame(gameId, playerId);
            if (!game) {
              throw new Error("Saved game not found on server");
            }

            // Automatically reconnect to the game
            setIsFadingOut(true);
            setTimeout(() => {
              navigate("/game", {
                state: {
                  game: game,
                  playerId: playerId,
                  playerName: playerName,
                  isReconnection: true,
                },
              });
            }, 300);
          }
        }
      } catch (err: any) {
        void err;
        // Clear invalid saved game data
        clearGameSession();
        setError("Unable to load previous game");
      }
    };

    void checkExistingGame();
  }, [preloadSkybox, navigate]);

  const handleCreateGame = (e: React.MouseEvent<HTMLAnchorElement>) => {
    // Allow CTRL+Click, CMD+Click, and middle mouse button to open in new tab
    if (e.ctrlKey || e.metaKey || e.button === 1) {
      return;
    }

    // For normal clicks, prevent default and use fade-out animation
    e.preventDefault();
    setIsFadingOut(true);
    setTimeout(() => {
      navigate("/create");
    }, 300); // Match CSS transition duration
  };

  const handleJoinGame = (e: React.MouseEvent<HTMLAnchorElement>) => {
    // Allow CTRL+Click, CMD+Click, and middle mouse button to open in new tab
    if (e.ctrlKey || e.metaKey || e.button === 1) {
      return;
    }

    // For normal clicks, prevent default and use fade-out animation
    e.preventDefault();
    setIsFadingOut(true);
    setTimeout(() => {
      navigate("/join");
    }, 300); // Match CSS transition duration
  };

  const handleReconnect = async () => {
    if (!savedGameData) return;

    setIsFadingOut(true);
    setTimeout(async () => {
      try {
        // Verify game still exists before attempting reconnection
        const game = await apiService.getGame(savedGameData.game.id);
        if (!game) {
          // Game no longer exists, clear storage and show error
          console.log(
            "Game no longer exists, clearing session and showing error",
          );
          clearGameSession();
          setError("Game no longer exists");
          setIsFadingOut(false);
          setSavedGameData(null);
          return;
        }

        // Reconnect to the game using global WebSocket manager
        await globalWebSocketManager.playerConnect(
          savedGameData.playerName,
          savedGameData.game.id,
          savedGameData.playerId,
        );

        // Navigate to game interface with the retrieved game state
        navigate("/game", {
          state: {
            game: savedGameData.game,
            playerId: savedGameData.playerId,
            playerName: savedGameData.playerName,
          },
        });
      } catch (err) {
        console.error("Failed to reconnect:", err);
        setError("Failed to reconnect to game");
        setIsFadingOut(false);
      }
    }, 300); // Match CSS transition duration
  };

  const handleDemoGame = async (e: React.MouseEvent<HTMLButtonElement>) => {
    e.preventDefault();
    if (isCreatingDemo) return;

    setIsCreatingDemo(true);
    setError(null);

    try {
      const result = await apiService.createDemoLobby({
        playerCount: 5, // Max players - actual count determined by who joins
        playerName: "You",
      });

      // Store session in localStorage
      localStorage.setItem(
        "terraforming-mars-game",
        JSON.stringify({
          gameId: result.game.id,
          playerId: result.playerId,
          playerName: "You",
          createdAt: new Date().toISOString(),
        }),
      );

      // Initialize WebSocket
      await globalWebSocketManager.initialize();

      // Connect player via WebSocket
      await globalWebSocketManager.playerConnect(
        "You",
        result.game.id,
        result.playerId,
      );

      // Navigate to game with fade-out
      setIsFadingOut(true);
      setTimeout(() => {
        navigate("/game", {
          state: {
            game: result.game,
            playerId: result.playerId,
            playerName: "You",
          },
        });
      }, 300);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to create demo lobby",
      );
      setIsCreatingDemo(false);
    }
  };

  return (
    <div
      className={`flex items-center justify-center min-h-screen text-white font-sans transition-opacity duration-300 ease-out relative z-10 ${isFadingOut ? "opacity-0" : "opacity-100"}`}
    >
      <div className="relative z-[1] flex items-center justify-center w-full min-h-screen">
        <div className="text-center px-5 py-5">
          {/* Title with Orbitron font */}
          <h1 className="font-orbitron text-[56px] text-white mb-[60px] text-shadow-glow-strong font-bold tracking-wider-2xl text-center mx-auto leading-tight">
            TERRAFORMING
            <br />
            MARS
          </h1>

          {/* Main action buttons - smaller and darker */}
          <div className="flex gap-5 mb-10 justify-center">
            <Link
              to="/create"
              onClick={handleCreateGame}
              className="bg-space-black-darker/90 border-2 border-space-blue-500 rounded-xl px-10 py-5 cursor-pointer transition-all duration-300 backdrop-blur-space text-white text-lg font-semibold font-orbitron tracking-wide hover:border-space-blue-900 hover:shadow-glow hover:shadow-glow-lg hover:-translate-y-1 no-underline inline-block"
            >
              CREATE
            </Link>

            <Link
              to="/join"
              onClick={handleJoinGame}
              className="bg-space-black-darker/90 border-2 border-space-blue-500 rounded-xl px-10 py-5 cursor-pointer transition-all duration-300 backdrop-blur-space text-white text-lg font-semibold font-orbitron tracking-wide hover:border-space-blue-900 hover:shadow-glow hover:shadow-glow-lg hover:-translate-y-1 no-underline inline-block"
            >
              JOIN
            </Link>

            <button
              onClick={(e) => void handleDemoGame(e)}
              disabled={isCreatingDemo}
              className="bg-space-black-darker/90 border-2 border-space-blue-500 rounded-xl px-10 py-5 cursor-pointer transition-all duration-300 backdrop-blur-space text-white text-lg font-semibold font-orbitron tracking-wide hover:border-space-blue-900 hover:shadow-glow hover:shadow-glow-lg hover:-translate-y-1 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isCreatingDemo ? "CREATING..." : "DEMO"}
            </button>
          </div>

          {/* Error message - shown below buttons, near reconnect card */}
          {error && (
            <div className="text-error-red mb-5 p-3 bg-error-red/10 border border-error-red/30 rounded-lg text-sm">
              {error}
            </div>
          )}

          {/* Reconnect card - shown when saved game exists */}
          {savedGameData && (
            <div className="flex justify-center mb-10">
              <div className="w-[500px] bg-space-black-darker/90 border-2 border-space-blue-500 rounded-xl p-8 backdrop-blur-space transition-all duration-300 hover:border-space-blue-900 hover:shadow-glow hover:shadow-glow-lg hover:-translate-y-1">
                {/* Corporation Logo */}
                <div className="mb-6 flex justify-center">
                  {savedGameData.game.currentPlayer.corporation ? (
                    getCorporationLogo(
                      savedGameData.game.currentPlayer.corporation.name.toLowerCase(),
                    )
                  ) : (
                    <div className="text-white/60 text-sm italic">
                      No Corporation
                    </div>
                  )}
                </div>

                {/* Game Info */}
                <div className="flex justify-center gap-6 mb-4 text-white/90 text-base">
                  <div className="flex items-center gap-2">
                    <span className="font-semibold">
                      {1 + savedGameData.game.otherPlayers.length}
                    </span>
                    <span className="text-white/70">
                      {1 + savedGameData.game.otherPlayers.length === 1
                        ? "Player"
                        : "Players"}
                    </span>
                  </div>
                  <div className="text-white/40">•</div>
                  <div className="flex items-center gap-2">
                    <span className="text-white/70">Generation</span>
                    <span className="font-semibold">
                      {savedGameData.game.generation}
                    </span>
                  </div>
                </div>

                {/* Reconnect Button */}
                <button
                  onClick={handleReconnect}
                  className="w-full bg-space-blue-600 border-2 border-space-blue-500 rounded-lg py-4 px-6 cursor-pointer transition-all duration-300 text-white text-lg font-bold font-orbitron tracking-wide hover:bg-space-blue-500 hover:border-space-blue-900 hover:shadow-glow hover:shadow-glow-lg"
                >
                  RECONNECT
                </button>
              </div>
            </div>
          )}
        </div>

        {/* View Cards button - bottom right corner */}
        <Link
          to="/cards"
          className="fixed bottom-[30px] right-[30px] bg-space-black-darker/80 border border-white/20 rounded-lg py-3 px-5 text-white/70 cursor-pointer transition-all duration-300 text-sm backdrop-blur-space-light font-orbitron hover:text-white/95 hover:border-space-blue-600 hover:bg-space-black-darker/95 hover:shadow-glow-sm no-underline"
        >
          View Cards
        </Link>
      </div>
    </div>
  );
};

export default GameLandingPage;
