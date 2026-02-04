import React, { useRef, useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { GameDto } from "../../../types/generated/api-types.ts";
import { globalWebSocketManager } from "../../../services/globalWebSocketManager.ts";
import CopyLinkButton from "../buttons/CopyLinkButton.tsx";
import GameMenuButton from "../buttons/GameMenuButton.tsx";
import GameMenuModal from "./GameMenuModal.tsx";

interface WaitingRoomOverlayProps {
  game: GameDto;
  playerId: string;
  visible?: boolean;
  onExited?: () => void;
}

interface LeavingPlayer {
  id: string;
  name: string;
}

const WaitingRoomOverlay: React.FC<WaitingRoomOverlayProps> = ({
  game,
  playerId,
  visible,
  onExited,
}) => {
  const navigate = useNavigate();
  const isHost = game.hostPlayerId === playerId;
  const joinUrl = `${window.location.origin}/game/${game.id}?type=join`;
  const [showLeaveConfirm, setShowLeaveConfirm] = useState(false);
  const [leaveConfirmVisible, setLeaveConfirmVisible] = useState(true);
  const [pendingLeave, setPendingLeave] = useState(false);

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

  const handleCancelLeave = () => {
    setLeaveConfirmVisible(false);
    setPendingLeave(false);
  };

  const handleConfirmLeave = () => {
    setLeaveConfirmVisible(false);
    setPendingLeave(true);
  };

  const handleLeaveConfirmExited = () => {
    setShowLeaveConfirm(false);
    setLeaveConfirmVisible(true);
    if (pendingLeave) {
      setPendingLeave(false);
      globalWebSocketManager.disconnect();
      void navigate("/");
    }
  };

  const openLeaveConfirm = () => {
    setShowLeaveConfirm(true);
    setLeaveConfirmVisible(true);
  };

  return (
    <>
      {/* Leave Confirmation Modal */}
      {showLeaveConfirm && (
        <GameMenuModal
          title="Leave game?"
          subtitle={isHost && playerCount > 1 ? "Another player will become the host" : undefined}
          visible={leaveConfirmVisible}
          onExited={handleLeaveConfirmExited}
          showBackdrop={true}
          showSettings={false}
          zIndex={2000}
          onClose={handleCancelLeave}
        >
          <div className="flex gap-3 justify-center">
            <GameMenuButton variant="secondary" size="sm" onClick={handleCancelLeave}>
              Cancel
            </GameMenuButton>
            <GameMenuButton variant="error" size="sm" onClick={handleConfirmLeave}>
              Leave
            </GameMenuButton>
          </div>
        </GameMenuModal>
      )}

      <GameMenuModal
        title="Game Lobby"
        subtitle={`${playerCount} player${playerCount !== 1 ? "s" : ""} joined`}
        onBack={() => void navigate("/")}
        visible={visible}
        onExited={onExited}
      >
        {/* Leave Button - positioned at top-left of modal content */}
        <GameMenuButton
          variant="text"
          size="sm"
          onClick={openLeaveConfirm}
          className="absolute top-8 left-8 !p-2.5 hover:text-red-400"
        >
          <svg
            width="22"
            height="22"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            className="-scale-x-100"
          >
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
            <polyline points="16 17 21 12 16 7" />
            <line x1="21" y1="12" x2="9" y2="12" />
          </svg>
        </GameMenuButton>

        <style>{`
          @keyframes playerSlideIn {
            from { opacity: 0; transform: translateY(-8px) scale(0.95); max-height: 0; }
            to { opacity: 1; transform: translateY(0) scale(1); max-height: 60px; }
          }
          @keyframes playerSlideOut {
            from { opacity: 1; transform: translateY(0) scale(1); max-height: 60px; }
            to { opacity: 0; transform: translateY(-8px) scale(0.95); max-height: 0; padding: 0; margin: 0; }
          }
        `}</style>

        {/* Player List */}
        <div className="mb-6">
          <h3 className="text-white text-sm font-semibold mb-2 uppercase tracking-wide">Players</h3>
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
                      {isHost && player.id !== playerId && !player.isLeaving && (
                        <button
                          onClick={() => void globalWebSocketManager.kickPlayer(player.id)}
                          className="ml-1 text-red-400 hover:text-red-300 transition-colors cursor-pointer"
                          title={`Kick ${player.name}`}
                        >
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
                            <line x1="18" y1="6" x2="6" y2="18" />
                            <line x1="6" y1="6" x2="18" y2="18" />
                          </svg>
                        </button>
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
            <GameMenuButton
              variant="primary"
              size="lg"
              onClick={handleStartGame}
              disabled={playerCount < 1}
              className="w-full"
            >
              START GAME
            </GameMenuButton>
          </div>
        )}

        {!isHost && (
          <p className="text-white/50 text-sm text-center">Waiting for host to start the game...</p>
        )}
      </GameMenuModal>
    </>
  );
};

export default WaitingRoomOverlay;
