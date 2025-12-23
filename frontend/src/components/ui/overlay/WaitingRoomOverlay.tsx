import React from "react";
import { GameDto } from "../../../types/generated/api-types.ts";
import { globalWebSocketManager } from "../../../services/globalWebSocketManager.ts";
import CopyLinkButton from "../buttons/CopyLinkButton.tsx";

interface WaitingRoomOverlayProps {
  game: GameDto;
  playerId: string;
}

const WaitingRoomOverlay: React.FC<WaitingRoomOverlayProps> = ({
  game,
  playerId,
}) => {
  const isHost = game.hostPlayerId === playerId;
  const joinUrl = `${window.location.origin}/join?code=${game.id}`;

  const handleStartGame = () => {
    if (!isHost) return;
    void globalWebSocketManager.startGame();
  };

  const playerCount =
    (game.currentPlayer ? 1 : 0) + (game.otherPlayers?.length || 0);

  return (
    <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 z-[1000] w-[450px] max-w-[90vw] max-h-[90vh] overflow-y-auto animate-[modalFadeIn_0.3s_ease-out]">
      <div className="bg-space-black-darker/95 border-2 border-space-blue-400 rounded-[20px] p-8 backdrop-blur-space shadow-[0_20px_60px_rgba(0,0,0,0.6),0_0_40px_rgba(30,60,150,0.3)]">
        <style>{`
          @keyframes modalFadeIn {
            0% { opacity: 0; }
            100% { opacity: 1; }
          }
        `}</style>

        {/* Header */}
        <div className="text-center mb-6">
          <h2 className="font-orbitron text-white text-[24px] m-0 mb-2 text-shadow-glow font-bold tracking-wider">
            Game Lobby
          </h2>
          <p className="text-white/60 text-sm m-0">
            {playerCount} player{playerCount !== 1 ? "s" : ""} joined
          </p>
        </div>

        {/* Player List */}
        <div className="mb-6">
          <h3 className="text-white text-sm font-semibold mb-2 uppercase tracking-wide">
            Players
          </h3>
          <div className="flex flex-col gap-2">
            {(() => {
              const playerMap = new Map();
              if (game.currentPlayer) {
                playerMap.set(game.currentPlayer.id, game.currentPlayer);
              }
              game.otherPlayers?.forEach((otherPlayer) => {
                playerMap.set(otherPlayer.id, otherPlayer);
              });

              const orderedPlayers =
                game.turnOrder && game.turnOrder.length > 0
                  ? game.turnOrder
                      .map((pid) => playerMap.get(pid))
                      .filter((player) => player !== undefined)
                  : Array.from(playerMap.values());

              return orderedPlayers.map((player) => (
                <div
                  key={player.id}
                  className="flex justify-between items-center py-2 px-3 bg-black/40 rounded-lg border border-space-blue-600/50"
                >
                  <span className="text-white text-sm font-medium">
                    {player.name}
                  </span>
                  <div className="flex gap-1.5 items-center">
                    {player.id === playerId && (
                      <span className="bg-space-blue-800 text-white py-0.5 px-1.5 rounded text-[10px] font-bold uppercase">
                        You
                      </span>
                    )}
                    {game.hostPlayerId === player.id && (
                      <span className="bg-gradient-to-br from-[#ffa500] to-[#ff8c00] text-white py-0.5 px-1.5 rounded text-[10px] font-bold uppercase">
                        Host
                      </span>
                    )}
                  </div>
                </div>
              ));
            })()}
          </div>

          {/* Join Link */}
          <div className="mt-4">
            <CopyLinkButton textToCopy={joinUrl} defaultText="Copy Join Link" />
          </div>
        </div>

        {/* Start Game Button (Host only) */}
        {isHost && (
          <div className="text-center">
            <button
              className="w-full bg-space-blue-600 border-2 border-space-blue-400 rounded-xl py-3 px-8 text-lg font-bold text-white cursor-pointer transition-all duration-300 hover:bg-space-blue-500 hover:shadow-glow disabled:bg-gray-700/50 disabled:border-gray-500/30 disabled:cursor-not-allowed disabled:opacity-60"
              onClick={handleStartGame}
              disabled={playerCount < 1}
            >
              Start Game
            </button>
          </div>
        )}

        {!isHost && (
          <p className="text-white/50 text-sm text-center">
            Waiting for host to start the game...
          </p>
        )}
      </div>
    </div>
  );
};

export default WaitingRoomOverlay;
