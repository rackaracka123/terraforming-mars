import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiService } from "../../services/apiService";
import { globalWebSocketManager } from "../../services/globalWebSocketManager.ts";
import { useSpaceBackground } from "../../contexts/SpaceBackgroundContext.tsx";
import LoadingOverlay from "../ui/overlay/LoadingOverlay";
import backgroundMusicService from "../../services/backgroundMusicService.ts";

const GameLandingPage: React.FC = () => {
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isFadingOut, setIsFadingOut] = useState(false);
  const { preloadSkybox } = useSpaceBackground();

  useEffect(() => {
    const checkExistingGame = async () => {
      try {
        // Preload skybox in parallel with game check
        void preloadSkybox();

        // Check localStorage for existing game
        const savedGameData = localStorage.getItem("terraforming-mars-game");
        if (savedGameData) {
          const { gameId, playerId, playerName } = JSON.parse(savedGameData);

          if (gameId && playerId) {
            // Try to get the current game state from server
            const game = await apiService.getGame(gameId);
            if (!game) {
              throw new Error("Saved game not found on server");
            }

            // Reconnect to the game using global WebSocket manager with playerId for proper reconnection
            await globalWebSocketManager.playerConnect(
              playerName,
              gameId,
              playerId,
            );

            // Navigate to game interface with the retrieved game state
            navigate("/game", {
              state: {
                game,
                playerId,
                playerName,
              },
            });
            return;
          }
        }

        // No saved game found - we're on the actual landing page, start music
        void backgroundMusicService.play();
        setIsLoading(false);
      } catch (err: any) {
        void err;
        // Clear invalid saved game data
        localStorage.removeItem("terraforming-mars-game");
        setError("Unable to reconnect to previous game");
        // Start music since we're staying on landing page
        void backgroundMusicService.play();
        setIsLoading(false);
      }
    };

    void checkExistingGame();

    // Cleanup: stop music when unmounting
    return () => {
      backgroundMusicService.stop();
    };
  }, [navigate, preloadSkybox]);

  const handleCreateGame = () => {
    setIsFadingOut(true);
    setTimeout(() => {
      navigate("/create");
    }, 300); // Match CSS transition duration
  };

  const handleJoinGame = () => {
    setIsFadingOut(true);
    setTimeout(() => {
      navigate("/join");
    }, 300); // Match CSS transition duration
  };

  const handleViewCards = () => {
    navigate("/cards");
  };

  return (
    <div
      className={`flex items-center justify-center min-h-screen text-white font-sans transition-opacity duration-300 ease-out relative z-10 ${isFadingOut ? "opacity-0" : "opacity-100"}`}
    >
      <div className="relative z-[1] flex items-center justify-center w-full min-h-screen">
        <div className="text-center max-w-[500px] px-5 py-5">
          {/* Title with Orbitron font */}
          <h1 className="font-orbitron text-[56px] text-white mb-[60px] text-shadow-glow-strong font-bold tracking-wider-2xl">
            TERRAFORMING MARS
          </h1>

          {error && (
            <div className="text-error-red mb-[30px] p-3 bg-error-red/10 border border-error-red/30 rounded-lg text-sm">
              {error}
            </div>
          )}

          {/* Main action buttons - smaller and darker */}
          <div className="flex gap-5 mb-10 justify-center">
            <button
              onClick={handleCreateGame}
              className="bg-space-black-darker/90 border-2 border-space-blue-500 rounded-xl px-10 py-5 cursor-pointer transition-all duration-300 backdrop-blur-space text-white text-lg font-semibold font-orbitron tracking-wide hover:border-space-blue-900 hover:shadow-glow hover:shadow-glow-lg hover:-translate-y-1"
            >
              CREATE
            </button>

            <button
              onClick={handleJoinGame}
              className="bg-space-black-darker/90 border-2 border-space-blue-500 rounded-xl px-10 py-5 cursor-pointer transition-all duration-300 backdrop-blur-space text-white text-lg font-semibold font-orbitron tracking-wide hover:border-space-blue-900 hover:shadow-glow hover:shadow-glow-lg hover:-translate-y-1"
            >
              JOIN
            </button>
          </div>
        </div>

        {/* View Cards button - bottom right corner */}
        <button
          onClick={handleViewCards}
          className="fixed bottom-[30px] right-[30px] bg-space-black-darker/80 border border-white/20 rounded-lg py-3 px-5 text-white/70 cursor-pointer transition-all duration-300 text-sm backdrop-blur-space-light font-orbitron hover:text-white/95 hover:border-space-blue-600 hover:bg-space-black-darker/95 hover:shadow-glow-sm"
        >
          View Cards
        </button>
      </div>

      <LoadingOverlay
        isLoading={isLoading}
        message="Checking for existing game..."
      />
    </div>
  );
};

export default GameLandingPage;
