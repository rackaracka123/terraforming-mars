import React, { useState, useEffect } from 'react';
import { io, Socket } from 'socket.io-client';
import GameLayout from './GameLayout.tsx';
import CorporationSelectionModal from '../../ui/overlay/CorporationSelectionModal.tsx';

interface GameState {
  id: string;
  players: Player[];
  currentPlayer: string;
  generation: number;
  phase: string;
  globalParameters: {
    temperature: number;
    oxygen: number;
    oceans: number;
  };
}

interface Player {
  id: string;
  name: string;
  resources: {
    credits: number;
    steel: number;
    titanium: number;
    plants: number;
    energy: number;
    heat: number;
  };
  production: any;
  terraformRating: number;
  victoryPoints: number;
  corporation?: string;
  passed?: boolean;
  availableActions?: number;
}

export default function GameInterface() {
  const [socket, setSocket] = useState<Socket | null>(null);
  const [gameState, setGameState] = useState<GameState | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [currentPlayer, setCurrentPlayer] = useState<Player | null>(null);
  const [showCorporationModal, setShowCorporationModal] = useState(false);
  const [availableCorporations, setAvailableCorporations] = useState<any[]>([]);

  useEffect(() => {
    const newSocket = io('http://localhost:3001');
    setSocket(newSocket);

    newSocket.on('connect', () => {
      setIsConnected(true);
      console.log('Connected to server');
      // Auto-join with a default name
      newSocket.emit('join-game', { gameId: 'demo', playerName: 'Player' });
    });

    newSocket.on('game-updated', (updatedGameState: GameState) => {
      setGameState(updatedGameState);
      const player = updatedGameState.players.find(p => p.id === newSocket.id);
      setCurrentPlayer(player || null);
      
      // Show corporation modal if player hasn't selected a corporation yet
      if (player && !player.corporation) {
        setShowCorporationModal(true);
      } else {
        setShowCorporationModal(false);
      }
    });

    newSocket.on('corporations-available', (corporations: any[]) => {
      setAvailableCorporations(corporations);
    });

    newSocket.on('disconnect', () => {
      setIsConnected(false);
      console.log('Disconnected from server');
    });

    return () => {
      newSocket.disconnect();
    };
  }, []);

  const handleCorporationSelection = (corporationId: string) => {
    if (socket) {
      socket.emit('select-corporation', { corporationId });
      setShowCorporationModal(false);
    }
  };

  if (!isConnected || !gameState) {
    return (
      <div style={{ padding: '20px', color: 'white', background: '#000011', minHeight: '100vh' }}>
        <h2>Connecting to Terraforming Mars server...</h2>
      </div>
    );
  }

  return (
    <>
      <GameLayout 
        gameState={gameState} 
        currentPlayer={currentPlayer} 
        socket={socket}
      />
      
      <CorporationSelectionModal
        corporations={availableCorporations}
        onSelectCorporation={handleCorporationSelection}
        isVisible={showCorporationModal}
      />
    </>
  );
}