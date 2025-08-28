import React, { useState, useEffect } from 'react';
import { io, Socket } from 'socket.io-client';
import GameLayout from './GameLayout.tsx';

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
}

export default function GameInterface() {
  const [socket, setSocket] = useState<Socket | null>(null);
  const [gameState, setGameState] = useState<GameState | null>(null);
  const [playerName, setPlayerName] = useState('');
  const [isConnected, setIsConnected] = useState(false);
  const [currentPlayer, setCurrentPlayer] = useState<Player | null>(null);

  useEffect(() => {
    const newSocket = io('http://localhost:3001');
    setSocket(newSocket);

    newSocket.on('connect', () => {
      setIsConnected(true);
      console.log('Connected to server');
    });

    newSocket.on('game-updated', (updatedGameState: GameState) => {
      setGameState(updatedGameState);
      const player = updatedGameState.players.find(p => p.id === newSocket.id);
      setCurrentPlayer(player || null);
    });

    newSocket.on('disconnect', () => {
      setIsConnected(false);
      console.log('Disconnected from server');
    });

    return () => {
      newSocket.disconnect();
    };
  }, []);

  const joinGame = () => {
    if (socket && playerName.trim()) {
      socket.emit('join-game', { gameId: 'demo', playerName: playerName.trim() });
    }
  };



  if (!isConnected) {
    return (
      <div style={{ padding: '20px', color: 'white', background: '#000011', minHeight: '100vh' }}>
        <h2>Connecting to Terraforming Mars server...</h2>
      </div>
    );
  }

  if (!gameState) {
    return (
      <div style={{ padding: '20px', color: 'white', background: '#000011', minHeight: '100vh' }}>
        <h2>Terraforming Mars 3D</h2>
        <div>
          <input
            type="text"
            placeholder="Enter your name"
            value={playerName}
            onChange={(e) => setPlayerName(e.target.value)}
            style={{ padding: '10px', marginRight: '10px', fontSize: '16px' }}
          />
          <button
            onClick={joinGame}
            disabled={!playerName.trim()}
            style={{
              padding: '10px 20px',
              fontSize: '16px',
              backgroundColor: '#4a90e2',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: playerName.trim() ? 'pointer' : 'not-allowed'
            }}
          >
            Join Demo Game
          </button>
        </div>
      </div>
    );
  }

  return (
    <GameLayout 
      gameState={gameState} 
      currentPlayer={currentPlayer} 
      socket={socket}
    />
  );
}