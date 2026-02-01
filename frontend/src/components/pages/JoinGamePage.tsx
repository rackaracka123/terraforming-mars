import React, { useEffect, useRef, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { apiService } from "../../services/apiService";
import { GameDto } from "../../types/generated/api-types.ts";
import EnterCodePopover from "../ui/popover/EnterCodePopover.tsx";
import LoadingOverlay from "../game/view/LoadingOverlay.tsx";
import { useJoinGame } from "@/hooks/useJoinGame";

const UUID_V4_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

type JoinPageView = "browse" | "enterName";

const JoinGamePage: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [availableGames, setAvailableGames] = useState<GameDto[]>([]);
  const [isLoadingGames, setIsLoadingGames] = useState(false);
  const [selectedGame, setSelectedGame] = useState<GameDto | null>(null);
  const [initialCode, setInitialCode] = useState<string | undefined>(undefined);
  const [isFadedIn, setIsFadedIn] = useState(false);

  const [currentView, setCurrentView] = useState<JoinPageView>("browse");
  const [isTransitioning, setIsTransitioning] = useState(false);
  const [showEnterCodePopover, setShowEnterCodePopover] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const enterCodeButtonRef = useRef<HTMLButtonElement>(null);

  const pendingViewRef = useRef<JoinPageView | null>(null);

  const { playerName, setPlayerName, isLoading, handleJoin, handleKeyDown, loadingMessage } =
    useJoinGame({ game: selectedGame });

  useEffect(() => {
    setTimeout(() => {
      setIsFadedIn(true);
    }, 10);
  }, []);

  const fetchGames = async () => {
    setIsLoadingGames(true);
    try {
      const games = await apiService.listGames();
      const lobbyGames = games.filter((g) => g.status === "lobby");
      const reconnectGames = games.filter(
        (g) =>
          g.status === "active" &&
          [...(g.otherPlayers || []), ...(g.currentPlayer ? [g.currentPlayer] : [])].some(
            (p) => !p.isConnected,
          ),
      );
      setAvailableGames([...lobbyGames, ...reconnectGames]);
    } catch {
      setAvailableGames([]);
    } finally {
      setIsLoadingGames(false);
    }
  };

  useEffect(() => {
    void fetchGames();
  }, []);

  useEffect(() => {
    const urlParams = new URLSearchParams(location.search);
    const codeParam = urlParams.get("code");

    if (codeParam && UUID_V4_REGEX.test(codeParam)) {
      setInitialCode(codeParam);
      setShowEnterCodePopover(true);
    }
  }, [location.search]);

  const handleBackToHome = () => {
    navigate("/");
  };

  const transitionToView = (view: JoinPageView) => {
    pendingViewRef.current = view;
    setIsTransitioning(true);
  };

  const handleGameValidated = (game: GameDto) => {
    setShowEnterCodePopover(false);
    setSelectedGame(game);
    transitionToView("enterName");
  };

  const handleJoinGame = (game: GameDto) => {
    setSelectedGame(game);
    transitionToView("enterName");
  };

  const handleBack = () => {
    transitionToView("browse");
  };

  const handleTransitionEnd = () => {
    if (isTransitioning && pendingViewRef.current !== null) {
      setCurrentView(pendingViewRef.current);
      if (pendingViewRef.current === "browse") {
        setSelectedGame(null);
      }
      pendingViewRef.current = null;
      setIsTransitioning(false);
    }
  };

  const selectedPlayerCount = selectedGame
    ? (selectedGame.currentPlayer ? 1 : 0) + (selectedGame.otherPlayers?.length || 0)
    : 0;
  const selectedMaxPlayers = selectedGame?.settings?.maxPlayers || 4;

  return (
    <div
      className={`bg-transparent text-white min-h-screen flex items-center justify-center font-sans relative z-10 transition-opacity duration-300 ease-in ${isFadedIn ? "opacity-100" : "opacity-0"}`}
    >
      <div className="relative z-[1] flex items-center justify-center w-full min-h-screen">
        <button
          onClick={currentView === "enterName" ? handleBack : handleBackToHome}
          className="fixed top-[30px] left-[30px] bg-space-black-darker/80 border border-white/20 rounded-lg py-2.5 px-4 text-white text-sm cursor-pointer hover:bg-white/20 transition-colors backdrop-blur-space z-[100]"
        >
          &larr; Back
        </button>
        <div className="max-w-[600px] w-full px-5 py-10">
          <div
            className={`transition-opacity duration-300 ${isTransitioning ? "opacity-0" : "opacity-100"}`}
            onTransitionEnd={handleTransitionEnd}
          >
            {currentView === "browse" ? (
              <div className="text-center">
                <h1 className="font-orbitron text-[42px] text-white mb-8 text-shadow-glow font-bold tracking-wider">
                  Browse games
                </h1>

                <div className="max-w-[500px] mx-auto">
                  <div className="flex items-center gap-3 mb-6">
                    <div className="relative flex-1">
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth={2}
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-white/40 pointer-events-none z-10"
                      >
                        <circle cx="11" cy="11" r="8" />
                        <path d="M21 21l-4.35-4.35" />
                      </svg>
                      <input
                        type="text"
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        placeholder="Search games..."
                        spellCheck={false}
                        autoComplete="off"
                        className="w-full bg-space-black-darker/80 border border-white/20 rounded-lg py-2 pl-10 pr-3 text-white text-sm outline-none placeholder:text-white/40 focus:border-white/40 transition-colors backdrop-blur-space"
                      />
                    </div>
                    <button
                      ref={enterCodeButtonRef}
                      type="button"
                      onClick={() => setShowEnterCodePopover(true)}
                      className="bg-space-black-darker/80 border border-white/20 rounded-lg py-2 px-4 text-white text-sm cursor-pointer hover:bg-white/20 transition-colors backdrop-blur-space shrink-0"
                    >
                      Enter code
                    </button>
                    <button
                      type="button"
                      onClick={() => void fetchGames()}
                      disabled={isLoadingGames}
                      className="bg-space-black-darker/80 border border-white/20 rounded-lg p-2 text-white cursor-pointer hover:bg-white/20 transition-colors backdrop-blur-space disabled:opacity-50 disabled:cursor-not-allowed shrink-0"
                    >
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth={2}
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        className={`w-5 h-5${isLoadingGames ? " animate-spin" : ""}`}
                      >
                        <path d="M21 12a9 9 0 1 1-6.22-8.56" />
                        <polyline points="21 3 21 9 15 9" />
                      </svg>
                    </button>
                  </div>

                  <div className="min-h-[200px] max-h-[400px] overflow-y-auto">
                    {isLoadingGames ? (
                      <div className="text-white/50 text-sm text-center py-8">Loading games...</div>
                    ) : availableGames.length === 0 ? (
                      <div className="text-white/50 text-sm text-center py-8">
                        No games available. Create a new game or enter a game code.
                      </div>
                    ) : (
                      <div className="flex flex-col gap-3">
                        {availableGames
                          .filter((game) => {
                            if (!searchQuery.trim()) return true;
                            const query = searchQuery.toLowerCase();
                            const hostName = (
                              game.currentPlayer?.name ||
                              game.otherPlayers?.[0]?.name ||
                              ""
                            ).toLowerCase();
                            const playerNames = [
                              ...(game.currentPlayer ? [game.currentPlayer.name] : []),
                              ...(game.otherPlayers?.map((p) => p.name) || []),
                            ];
                            return (
                              hostName.includes(query) ||
                              playerNames.some((n) => n.toLowerCase().includes(query))
                            );
                          })
                          .map((game) => {
                            const playerCount =
                              (game.currentPlayer ? 1 : 0) + (game.otherPlayers?.length || 0);
                            const maxPlayers = game.settings?.maxPlayers || 4;
                            const hostName =
                              game.currentPlayer?.name || game.otherPlayers?.[0]?.name || "Unknown";
                            const isActive = game.status === "active";
                            return (
                              <div
                                key={game.id}
                                className="flex items-center justify-between bg-space-black-darker/95 border border-white/20 rounded-xl p-4 backdrop-blur-space"
                              >
                                <div className="flex flex-col gap-1 min-w-0 text-left">
                                  <span className="text-white text-sm font-medium truncate">
                                    {hostName}
                                  </span>
                                  <span className="text-white/50 text-xs">
                                    {playerCount}/{maxPlayers} Players
                                  </span>
                                </div>
                                {isActive ? (
                                  <button
                                    onClick={() => navigate(`/game/${game.id}`)}
                                    className="font-orbitron bg-space-blue-600 border border-space-blue-500 rounded-lg px-4 py-2 text-white text-sm font-medium hover:bg-space-blue-500 transition-colors shrink-0 ml-4"
                                  >
                                    Reconnect
                                  </button>
                                ) : (
                                  <button
                                    onClick={() => handleJoinGame(game)}
                                    className="font-orbitron bg-space-blue-600 border border-space-blue-500 rounded-lg px-4 py-2 text-white text-sm font-medium hover:bg-space-blue-500 transition-colors shrink-0 ml-4"
                                  >
                                    Join
                                  </button>
                                )}
                              </div>
                            );
                          })}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            ) : (
              <div className="max-w-[450px] mx-auto">
                <div className="bg-space-black-darker/95 rounded-[20px] p-8 backdrop-blur-space shadow-[0_20px_60px_rgba(0,0,0,0.6),0_0_40px_rgba(30,60,150,0.3)]">
                  <div className="text-center mb-6">
                    <h2 className="font-orbitron text-white text-[24px] m-0 mb-2 text-shadow-glow font-bold tracking-wider">
                      Join Game
                    </h2>
                    <p className="text-white/60 text-sm m-0">
                      {selectedPlayerCount}/{selectedMaxPlayers} players
                    </p>
                  </div>

                  <div className="mb-6">
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
                      className="w-full bg-black/40 border border-space-blue-600/50 rounded-lg py-2 px-3 text-white text-sm font-medium outline-none placeholder:text-white/50 focus:border-space-blue-400 transition-colors disabled:opacity-60"
                    />
                  </div>

                  <div className="text-center">
                    <button
                      onClick={() => void handleJoin()}
                      disabled={isLoading || !playerName.trim()}
                      className="w-full bg-space-black-darker/90 border-2 border-space-blue-500 rounded-xl py-4 px-8 text-lg font-semibold font-orbitron tracking-wide text-white cursor-pointer transition-all duration-300 backdrop-blur-space hover:border-space-blue-900 hover:shadow-glow hover:shadow-glow-lg hover:-translate-y-1 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:translate-y-0 disabled:hover:shadow-none"
                    >
                      {isLoading ? "JOINING..." : "JOIN GAME"}
                    </button>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      <EnterCodePopover
        isVisible={showEnterCodePopover}
        onClose={() => setShowEnterCodePopover(false)}
        onGameValidated={handleGameValidated}
        initialCode={initialCode}
        anchorRef={enterCodeButtonRef}
      />

      {isLoading && <LoadingOverlay isLoaded={false} message={loadingMessage} />}
    </div>
  );
};

export default JoinGamePage;
