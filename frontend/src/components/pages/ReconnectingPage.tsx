import React, { useState, useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { globalWebSocketManager } from "../../services/globalWebSocketManager.ts";
import { apiService } from "../../services/apiService.ts";
import {
  clearGameSession,
  getGameSession,
} from "../../utils/sessionStorage.ts";

interface ReconnectingPageProps {
  // Add any props if needed
}

const ReconnectingPage: React.FC<ReconnectingPageProps> = () => {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [isReconnecting, setIsReconnecting] = useState(true);
  const reconnectionAttempted = useRef(false);

  useEffect(() => {
    const attemptReconnection = async () => {
      // Prevent multiple reconnection attempts (React Strict Mode protection)
      if (reconnectionAttempted.current) {
        return;
      }
      reconnectionAttempted.current = true;

      try {
        setError(null);
        setIsReconnecting(true);

        // Check localStorage for saved game data
        const savedGameData = getGameSession();
        if (!savedGameData) {
          // No saved data, clear storage and go to landing page
          clearGameSession();
          navigate("/", { replace: true });
          return;
        }

        const { gameId, playerId, playerName } = savedGameData;
        if (!gameId || !playerName || !playerId) {
          // Invalid data, clear storage and go to landing page
          clearGameSession();
          navigate("/", { replace: true });
          return;
        }

        // Verify game still exists
        const game = await apiService.getGame(gameId);
        if (!game) {
          // Game no longer exists, automatically clear storage and redirect
          console.log(
            "Game no longer exists, clearing session and returning to landing page",
          );
          clearGameSession();
          navigate("/", { replace: true });
          return;
        }

        // Ensure WebSocket is ready and attempt reconnection
        // Attempting to reconnect: playerName, gameId, playerId
        const reconnectionResult = await globalWebSocketManager.playerConnect(
          playerName,
          gameId,
          playerId,
        );
        // Reconnection successful

        if (reconnectionResult.game) {
          // CRITICAL FIX: Set the current player ID in globalWebSocketManager
          // This ensures the GameInterface component knows which player this client represents
          globalWebSocketManager.setCurrentPlayerId(
            reconnectionResult.playerId,
          );

          // Update localStorage with fresh data
          localStorage.setItem(
            "terraforming-mars-game",
            JSON.stringify({
              gameId: reconnectionResult.game.id,
              playerId: reconnectionResult.playerId,
              playerName: reconnectionResult.playerName,
              timestamp: Date.now(),
            }),
          );

          // Navigate to game with reconnected state
          navigate("/game", {
            state: {
              game: reconnectionResult.game,
              playerId: reconnectionResult.playerId,
              playerName: reconnectionResult.playerName,
              isReconnection: true,
            },
            replace: true,
          });
        }
      } catch (error: any) {
        console.error("Reconnection failed:", error);
        setError(error.message || "Reconnection failed");
      } finally {
        setIsReconnecting(false);
      }
    };

    attemptReconnection();
  }, [navigate]);

  const handleReturnToMenu = () => {
    clearGameSession();
    navigate("/", { replace: true });
  };

  return (
    <div className="bg-[#000011] text-white min-h-screen flex items-center justify-center font-sans">
      <div className="max-w-[600px] w-full py-10 px-5 max-[768px]:py-5 max-[768px]:px-[15px]">
        <div className="text-center">
          <h1 className="text-5xl text-white mb-[60px] [text-shadow:0_2px_4px_rgba(0,0,0,0.8)] font-bold max-[768px]:text-4xl max-[768px]:mb-10 max-[480px]:text-[28px] max-[480px]:mb-[30px]">
            Terraforming Mars
          </h1>

          {isReconnecting ? (
            <div className="flex flex-col items-center gap-6">
              <div className="w-[60px] h-[60px] border-4 border-[rgba(74,144,226,0.3)] border-t-[#4a90e2] rounded-full animate-[spin_1s_linear_infinite] max-[768px]:w-[50px] max-[768px]:h-[50px]"></div>
              <h2 className="text-[32px] text-[#4a90e2] m-0 [text-shadow:0_2px_4px_rgba(0,0,0,0.8)] font-bold max-[768px]:text-2xl max-[480px]:text-xl">
                Reconnecting to game...
              </h2>
              <p className="text-lg text-white/80 m-0 max-w-[400px] max-[768px]:text-base">
                Please wait while we restore your connection
              </p>
            </div>
          ) : error ? (
            <div className="flex flex-col items-center gap-5">
              <div className="text-[64px] mb-2">⚠️</div>
              <h2 className="text-[28px] text-[#ff6b6b] m-0 [text-shadow:0_2px_4px_rgba(0,0,0,0.8)] font-bold max-[768px]:text-2xl max-[480px]:text-xl">
                Reconnection Failed
              </h2>
              <p className="text-base text-white/90 m-0 max-w-[400px] leading-[1.5] max-[768px]:text-sm">
                {error}
              </p>
              <button
                className="bg-[linear-gradient(135deg,#4a90e2_0%,#5ba0f2_100%)] border-none rounded-xl py-4 px-8 text-lg font-bold text-white cursor-pointer transition-all duration-300 [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] shadow-[0_4px_20px_rgba(74,144,226,0.3)] mt-4 hover:bg-[linear-gradient(135deg,#357abd_0%,#4a90e2_100%)] hover:-translate-y-0.5 hover:shadow-[0_6px_25px_rgba(74,144,226,0.4)] max-[768px]:py-3.5 max-[768px]:px-7 max-[768px]:text-base max-[480px]:py-3 max-[480px]:px-6 max-[480px]:text-sm"
                onClick={handleReturnToMenu}
              >
                Return to Main Menu
              </button>
            </div>
          ) : null}
        </div>
      </div>
    </div>
  );
};

export default ReconnectingPage;
