import React from "react";
import { GameDto } from "@/types/generated/api-types";
import { useJoinGame } from "@/hooks/useJoinGame";
import LoadingOverlay from "../../game/view/LoadingOverlay";
import MainMenuSettingsButton from "../buttons/MainMenuSettingsButton";

interface JoinGameOverlayProps {
  game: GameDto;
  onCancel: () => void;
}

const JoinGameOverlay: React.FC<JoinGameOverlayProps> = ({ game, onCancel }) => {
  const { playerName, setPlayerName, isLoading, handleJoin, handleKeyDown, loadingMessage } =
    useJoinGame({ game });

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
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 z-[1000] w-[450px] max-w-[90vw] animate-[modalFadeIn_0.3s_ease-out]">
        <div className="bg-space-black-darker/95 border-2 border-space-blue-400 rounded-[20px] p-8 backdrop-blur-space shadow-[0_20px_60px_rgba(0,0,0,0.6),0_0_40px_rgba(30,60,150,0.3)]">
          <div className="text-center mb-6">
            <h2 className="font-orbitron text-white text-[24px] m-0 mb-2 text-shadow-glow font-bold tracking-wider">
              Enter your name
            </h2>
          </div>

          <div className="flex flex-row gap-3 items-center px-2">
            <input
              type="text"
              value={playerName}
              onChange={(e) => setPlayerName(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Enter your name"
              disabled={isLoading}
              spellCheck={false}
              autoComplete="off"
              autoCorrect="off"
              maxLength={50}
              autoFocus
              className="flex-1 bg-black/50 border border-white/20 rounded-lg py-3 px-4 text-white text-base outline-none placeholder:text-white/50 focus:border-white/60 focus:shadow-[0_0_20px_rgba(255,255,255,0.1)] transition-all duration-200 disabled:opacity-60"
            />
            <button
              onClick={() => void handleJoin()}
              disabled={isLoading || !playerName.trim()}
              className="font-orbitron bg-space-blue-600 border border-space-blue-500 rounded-lg py-3 px-6 text-white text-sm font-medium hover:bg-space-blue-500 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? "Joining..." : "Join"}
            </button>
          </div>
        </div>
      </div>

      {isLoading && <LoadingOverlay isLoaded={false} message={loadingMessage} />}
    </>
  );
};

export default JoinGameOverlay;
