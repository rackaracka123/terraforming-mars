import React from "react";
import { GameDto } from "../../../types/generated/api-types.ts";
import { globalWebSocketManager } from "../../../services/globalWebSocketManager.ts";
import CopyLinkButton from "../buttons/CopyLinkButton.tsx";

interface WaitingRoomOverlayProps {
  game: GameDto;
  playerId: string;
  onStartGame?: () => void;
}

const WaitingRoomOverlay: React.FC<WaitingRoomOverlayProps> = ({
  game,
  playerId,
  onStartGame,
}) => {
  const isHost = game.hostPlayerId === playerId;
  const joinUrl = `${window.location.origin}/join?code=${game.id}`;

  const handleStartGame = () => {
    // Start Game button clicked, ishost: isHost

    if (!isHost) return;

    // Send start game action via WebSocket
    void globalWebSocketManager.startGame();
    onStartGame?.();
  };

  return (
    <>
      {/* Translucent overlay over Mars */}
      <div className="absolute top-0 left-0 right-0 bottom-0 bg-black/60 backdrop-blur-sm z-10" />

      {/* Waiting room controls */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 z-20 min-w-[400px] max-w-[600px] max-sm:min-w-[320px] max-sm:mx-5">
        <div className="bg-space-black-darker/95 border-2 border-space-blue-400 rounded-[20px] p-10 backdrop-blur-space shadow-[0_20px_60px_rgba(0,0,0,0.6),0_0_40px_rgba(30,60,150,0.3)] text-center max-sm:p-6">
          <div>
            <h2 className="font-orbitron text-white text-[28px] m-0 mb-4 text-shadow-glow font-bold tracking-wider max-sm:text-[22px]">
              Waiting for players to join...
            </h2>
            <p className="text-white/80 text-lg m-0 mb-8 max-sm:text-base">
              {(game.currentPlayer ? 1 : 0) + (game.otherPlayers?.length || 0)}{" "}
              / {game.settings?.maxPlayers || 4} players
            </p>

            {/* Player List */}
            <div className="mt-5 text-left flex flex-col gap-2">
              {(() => {
                // Create a map of all players (current + others) for easy lookup
                const playerMap = new Map();
                if (game.currentPlayer) {
                  playerMap.set(game.currentPlayer.id, game.currentPlayer);
                }
                game.otherPlayers?.forEach((otherPlayer) => {
                  playerMap.set(otherPlayer.id, otherPlayer);
                });

                // Use turnOrder to display players in correct order
                const orderedPlayers =
                  game.turnOrder
                    ?.map((playerId) => playerMap.get(playerId))
                    .filter((player) => player !== undefined) || [];

                return orderedPlayers.map((player) => (
                  <div
                    key={player.id}
                    className="flex justify-between items-center py-3 px-4 bg-black/40 rounded-lg border border-space-blue-600 shadow-[0_0_10px_rgba(30,60,150,0.3)]"
                  >
                    <span className="text-white text-base font-medium">
                      {player.name}
                    </span>
                    <div className="flex gap-2 items-center">
                      {player.id === playerId && (
                        <span className="bg-space-blue-800 text-white py-1 px-2 rounded-md text-xs font-bold uppercase shadow-[0_2px_8px_rgba(30,60,150,0.4)]">
                          You
                        </span>
                      )}
                      {game.hostPlayerId === player.id && (
                        <span className="bg-gradient-to-br from-[#ffa500] to-[#ff8c00] text-white py-1 px-2 rounded-md text-xs font-bold uppercase shadow-[0_2px_8px_rgba(255,140,0,0.3)]">
                          Host
                        </span>
                      )}
                    </div>
                  </div>
                ));
              })()}
            </div>
          </div>

          {isHost && (
            <div className="mt-12 mb-8">
              <button
                className="bg-space-black-darker/90 border-2 border-space-blue-800 rounded-xl py-4 px-8 text-xl font-bold text-white cursor-pointer transition-all duration-300 text-shadow-dark shadow-[0_4px_20px_rgba(30,60,150,0.3)] hover:bg-space-black-darker/95 hover:border-space-blue-600 hover:-translate-y-0.5 hover:shadow-glow disabled:bg-gray-700/50 disabled:border-gray-500/30 disabled:cursor-not-allowed disabled:transform-none disabled:shadow-none disabled:opacity-60 max-sm:py-3 max-sm:px-6 max-sm:text-lg"
                onClick={handleStartGame}
                disabled={!game.currentPlayer}
              >
                <span>Start Game</span>
              </button>
            </div>
          )}

          <div className="text-center mt-6">
            <CopyLinkButton textToCopy={joinUrl} defaultText="Copy Join Link" />
          </div>
        </div>
      </div>
    </>
  );
};

export default WaitingRoomOverlay;
