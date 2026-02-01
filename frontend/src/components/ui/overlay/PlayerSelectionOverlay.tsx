import React from "react";
import { GameDto, OtherPlayerDto, PlayerDto } from "../../../types/generated/api-types.ts";
import { getCorporationLogo } from "../../../utils/corporationLogos.tsx";
import MainMenuSettingsButton from "../buttons/MainMenuSettingsButton.tsx";

interface PlayerSelectionOverlayProps {
  game: GameDto;
  onSelectPlayer: (playerId: string, playerName: string) => void;
  onCancel: () => void;
}

const PlayerSelectionOverlay: React.FC<PlayerSelectionOverlayProps> = ({
  game,
  onSelectPlayer,
  onCancel,
}) => {
  const allPlayers: (PlayerDto | OtherPlayerDto)[] = [
    ...(game.currentPlayer ? [game.currentPlayer] : []),
    ...(game.otherPlayers || []),
  ];

  const orderedPlayers =
    game.turnOrder && game.turnOrder.length > 0
      ? game.turnOrder
          .map((pid) => allPlayers.find((p) => p.id === pid))
          .filter((player): player is PlayerDto | OtherPlayerDto => player !== undefined)
      : allPlayers;

  return (
    <>
      <button
        onClick={onCancel}
        className="fixed top-[30px] left-[30px] bg-space-black-darker/80 border border-white/20 rounded-lg py-2.5 px-4 text-white text-sm cursor-pointer hover:bg-white/20 transition-colors backdrop-blur-space z-[10000]"
      >
        &larr; Back
      </button>
      <div className="z-[10000]">
        <MainMenuSettingsButton />
      </div>
      <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-[999]" />
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 z-[1000] w-[450px] max-w-[90vw] max-h-[90vh] overflow-y-auto animate-[modalFadeIn_0.3s_ease-out]">
        <div className="bg-space-black-darker/95 border-2 border-space-blue-400 rounded-[20px] p-8 backdrop-blur-space shadow-[0_20px_60px_rgba(0,0,0,0.6),0_0_40px_rgba(30,60,150,0.3)]">
          <style>{`
            @keyframes modalFadeIn {
              0% { opacity: 0; }
              100% { opacity: 1; }
            }
          `}</style>

          <div className="text-center mb-6">
            <h2 className="font-orbitron text-white text-[24px] m-0 mb-2 text-shadow-glow font-bold tracking-wider">
              Reconnect to Game
            </h2>
            <p className="text-white/60 text-sm m-0">Select a player to reconnect as</p>
          </div>

          <div className="mb-6">
            <h3 className="text-white text-sm font-semibold mb-2 uppercase tracking-wide">
              Players
            </h3>
            <div className="flex flex-col gap-2">
              {orderedPlayers.map((player) => {
                const isConnected = player.isConnected;
                const canSelect = !isConnected;

                return (
                  <button
                    key={player.id}
                    onClick={() => canSelect && onSelectPlayer(player.id, player.name)}
                    disabled={!canSelect}
                    className={`flex justify-between items-center py-3 px-4 bg-black/40 rounded-lg border transition-all text-left w-full ${
                      canSelect
                        ? "border-space-blue-600/50 hover:border-space-blue-400 hover:bg-black/60 cursor-pointer"
                        : "border-white/10 opacity-50 cursor-not-allowed"
                    }`}
                  >
                    <div className="flex items-center gap-3">
                      {player.corporation && (
                        <div className="w-[80px] h-6 flex-shrink-0 flex items-center justify-start overflow-hidden">
                          <div className="origin-left scale-[0.2]">
                            {getCorporationLogo(
                              player.corporation.name.toLowerCase() as Parameters<
                                typeof getCorporationLogo
                              >[0],
                            )}
                          </div>
                        </div>
                      )}
                      <span className="text-white text-sm font-medium">{player.name}</span>
                    </div>
                    {!isConnected && (
                      <span className="px-1.5 py-px rounded-lg text-[8px] font-semibold uppercase tracking-[0.3px] shadow-[0_1px_2px_rgba(0,0,0,0.2)] bg-[linear-gradient(135deg,#e74c3c,#c0392b)] text-white border border-[rgba(231,76,60,0.5)]">
                        DISCONNECTED
                      </span>
                    )}
                  </button>
                );
              })}
            </div>
          </div>

          <p className="text-white/40 text-xs text-center">
            Only disconnected players can be selected
          </p>
        </div>
      </div>
    </>
  );
};

export default PlayerSelectionOverlay;
