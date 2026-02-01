import React, { useRef, useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { GameDto } from "../../../types/generated/api-types.ts";
import { globalWebSocketManager } from "../../../services/globalWebSocketManager.ts";
import CopyLinkButton from "../buttons/CopyLinkButton.tsx";
import MainMenuSettingsButton from "../buttons/MainMenuSettingsButton.tsx";

interface WaitingRoomOverlayProps {
  game: GameDto;
  playerId: string;
  isExiting?: boolean;
}

interface LeavingPlayer {
  id: string;
  name: string;
}

const WaitingRoomOverlay: React.FC<WaitingRoomOverlayProps> = ({
  game,
  playerId,
  isExiting = false,
}) => {
  const navigate = useNavigate();
  const isHost = game.hostPlayerId === playerId;
  const joinUrl = `${window.location.origin}/game/${game.id}?type=join`;

  const handleStartGame = () => {
    if (!isHost) return;
    void globalWebSocketManager.startGame();
  };

  const playerCount = (game.currentPlayer ? 1 : 0) + (game.otherPlayers?.length || 0);

  const allPlayers = React.useMemo(() => {
    const players: { id: string; name: string }[] = [];
    if (game.currentPlayer)
      players.push({ id: game.currentPlayer.id, name: game.currentPlayer.name });
    game.otherPlayers?.forEach((p) => players.push({ id: p.id, name: p.name }));
    return players;
  }, [game.currentPlayer, game.otherPlayers]);

  const currentPlayerIds = React.useMemo(() => new Set(allPlayers.map((p) => p.id)), [allPlayers]);

  const prevPlayerIdsRef = useRef<Set<string>>(new Set());
  const [newPlayerIds, setNewPlayerIds] = useState<Set<string>>(new Set());
  const [leavingPlayers, setLeavingPlayers] = useState<LeavingPlayer[]>([]);
  const prevPlayersRef = useRef<Map<string, string>>(new Map());

  const handleAnimationEnd = useCallback((id: string) => {
    setNewPlayerIds((prev) => {
      const next = new Set(prev);
      next.delete(id);
      return next;
    });
  }, []);

  useEffect(() => {
    const prevIds = prevPlayerIdsRef.current;

    const joined = new Set<string>();
    for (const id of currentPlayerIds) {
      if (!prevIds.has(id)) joined.add(id);
    }

    const left: LeavingPlayer[] = [];
    for (const id of prevIds) {
      if (!currentPlayerIds.has(id)) {
        left.push({ id, name: prevPlayersRef.current.get(id) ?? "Player" });
      }
    }

    if (joined.size > 0) setNewPlayerIds(joined);

    if (left.length > 0) {
      setLeavingPlayers((prev) => [...prev, ...left]);
      setTimeout(() => {
        setLeavingPlayers((prev) => prev.filter((p) => !left.some((l) => l.id === p.id)));
      }, 300);
    }

    prevPlayerIdsRef.current = new Set(currentPlayerIds);
    const nameMap = new Map<string, string>();
    allPlayers.forEach((p) => nameMap.set(p.id, p.name));
    prevPlayersRef.current = nameMap;
  }, [currentPlayerIds, allPlayers]);

  const animationClass = isExiting
    ? "animate-[lobbyExit_800ms_ease-out_forwards]"
    : "animate-[modalFadeIn_0.3s_ease-out]";

  return (
    <>
      <button
        onClick={() => void navigate("/")}
        className="fixed top-[30px] left-[30px] bg-space-black-darker/80 border border-white/20 rounded-lg py-2.5 px-4 text-white text-sm cursor-pointer hover:bg-white/20 transition-colors backdrop-blur-space z-[10000]"
      >
        ← Back
      </button>
      <div className="z-[10000]">
        <MainMenuSettingsButton />
      </div>
      <div
        className={`absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 z-[1000] w-[450px] max-w-[90vw] max-h-[90vh] overflow-y-auto ${animationClass}`}
      >
        <div className="bg-space-black-darker/95 border-2 border-space-blue-400 rounded-[20px] p-8 backdrop-blur-space shadow-[0_20px_60px_rgba(0,0,0,0.6),0_0_40px_rgba(30,60,150,0.3)]">
          <style>{`
          @keyframes modalFadeIn {
            0% { opacity: 0; }
            100% { opacity: 1; }
          }
          @keyframes lobbyExit {
            from {
              transform: scale(1);
              opacity: 1;
            }
            to {
              transform: scale(0.95);
              opacity: 0;
            }
          }
          @keyframes playerSlideIn {
            from { opacity: 0; transform: translateY(-8px) scale(0.95); max-height: 0; }
            to { opacity: 1; transform: translateY(0) scale(1); max-height: 60px; }
          }
          @keyframes playerSlideOut {
            from { opacity: 1; transform: translateY(0) scale(1); max-height: 60px; }
            to { opacity: 0; transform: translateY(-8px) scale(0.95); max-height: 0; padding: 0; margin: 0; }
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

                const playerItems = orderedPlayers.map((player) => ({
                  id: player.id,
                  name: player.name,
                  isLeaving: false,
                }));

                leavingPlayers.forEach((lp) => {
                  if (!playerMap.has(lp.id)) {
                    playerItems.push({ id: lp.id, name: lp.name, isLeaving: true });
                  }
                });

                return playerItems.map((player) => {
                  let animClass = "";
                  if (player.isLeaving) {
                    animClass = "animate-[playerSlideOut_0.3s_ease-out_forwards] overflow-hidden";
                  } else if (newPlayerIds.has(player.id)) {
                    animClass = "animate-[playerSlideIn_0.3s_ease-out]";
                  }

                  return (
                    <div
                      key={player.id}
                      className={`flex justify-between items-center py-2 px-3 bg-black/40 rounded-lg border border-space-blue-600/50 ${animClass}`}
                      onAnimationEnd={
                        !player.isLeaving && newPlayerIds.has(player.id)
                          ? () => handleAnimationEnd(player.id)
                          : undefined
                      }
                    >
                      <span className="text-white text-sm font-medium">{player.name}</span>
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
                  );
                });
              })()}
            </div>

            {/* Join Link */}
            <div className="mt-4 flex justify-center">
              <CopyLinkButton
                textToCopy={joinUrl}
                defaultText="Join Link"
                copiedText="Copied!"
                icon={
                  <svg
                    width="14"
                    height="14"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  >
                    <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                    <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                  </svg>
                }
              />
            </div>
          </div>

          {/* Start Game Button (Host only) */}
          {isHost && (
            <div className="text-center">
              <button
                className="w-full bg-space-black-darker/90 border-2 border-space-blue-500 rounded-xl py-4 px-8 text-lg font-semibold font-orbitron tracking-wide text-white cursor-pointer transition-all duration-300 backdrop-blur-space hover:border-space-blue-900 hover:shadow-glow hover:shadow-glow-lg hover:-translate-y-1 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:translate-y-0 disabled:hover:shadow-none"
                onClick={handleStartGame}
                disabled={playerCount < 1}
              >
                START GAME
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
    </>
  );
};

export default WaitingRoomOverlay;
