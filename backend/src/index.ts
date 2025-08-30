import express from 'express';
import { createServer } from 'http';
import { Server } from 'socket.io';
import cors from 'cors';
import { GameState, Player, GlobalParameters, ResourceType, GamePhase } from './types';

const app = express();
const httpServer = createServer(app);
const io = new Server(httpServer, {
  cors: {
    origin: "http://localhost:3000",
    methods: ["GET", "POST"]
  }
});

app.use(cors());
app.use(express.json());

// Simple in-memory game storage
const games = new Map<string, GameState>();
const players = new Map<string, string>(); // socketId -> playerId

// Initialize a demo game
const createDemoGame = (): GameState => ({
  id: 'demo',
  players: [],
  currentPlayer: '',
  generation: 1,
  phase: GamePhase.RESEARCH,
  globalParameters: {
    temperature: -30,
    oxygen: 0,
    oceans: 0
  },
  milestones: [],
  awards: [],
  firstPlayer: '',
  deck: [],
  discardPile: [],
  soloMode: false,
  turn: 1,
  currentActionCount: 0,
  maxActionsPerTurn: 2,
  gameSettings: {
    expansions: [],
    corporateEra: false,
    draftVariant: false,
    initialDraft: false,
    preludeExtension: false,
    venusNextExtension: false,
    coloniesExtension: false,
    turmoilExtension: false,
    removeNegativeAttackCards: false,
    includeVenusMA: false,
    moonExpansion: false,
    pathfindersExpansion: false,
    underworldExpansion: false,
    escapeVelocityExpansion: false,
    fast: false,
    showOtherPlayersVP: true,
    soloTR: false,
    randomFirstPlayer: false,
    requiresVenusTrackCompletion: false,
    requiresMoonTrackCompletion: false,
    moonStandardProjectVariant: false,
    altVenusBoard: false,
    escapeVelocityMode: false,
    escapeVelocityThreshold: 30,
    escapeVelocityPeriod: 2,
    escapeVelocityPenalty: 1,
    twoTempTerraformingThreshold: false,
    heatFor: false,
    breakthrough: false
  }
});

// Create demo game
games.set('demo', createDemoGame());

// Socket.io connection handling
io.on('connection', (socket) => {
  console.log('Player connected:', socket.id);

  socket.on('join-game', (data: { gameId: string, playerName: string }) => {
    const game = games.get(data.gameId);
    if (!game) {
      socket.emit('error', 'Game not found');
      return;
    }

    // Create new player
    const player: Player = {
      id: socket.id,
      name: data.playerName,
      resources: {
        credits: 20,
        steel: 0,
        titanium: 0,
        plants: 0,
        energy: 0,
        heat: 0
      },
      production: {
        credits: 1,
        steel: 0,
        titanium: 0,
        plants: 0,
        energy: 1,
        heat: 1
      },
      terraformRating: 20,
      victoryPoints: 0,
      playedCards: [],
      hand: [],
      availableActions: 2,
      tags: [],
      actionsTaken: 0,
      actionsRemaining: 2
    };

    game.players.push(player);
    players.set(socket.id, socket.id);

    if (game.currentPlayer === '') {
      game.currentPlayer = socket.id;
    }

    socket.join(data.gameId);
    io.to(data.gameId).emit('game-updated', game);
    
    console.log(`${data.playerName} joined game ${data.gameId}`);
  });


  socket.on('raise-temperature', (data: { gameId: string }) => {
    const game = games.get(data.gameId);
    if (!game) return;

    const player = game.players.find((p: Player) => p.id === socket.id);
    if (!player || game.currentPlayer !== socket.id) return;

    // Cost 8 heat to raise temperature
    if (player.resources.heat >= 8 && player.actionsRemaining && player.actionsRemaining > 0) {
      player.resources.heat -= 8;
      if (game.globalParameters.temperature < 8) {
        game.globalParameters.temperature += 2;
        player.terraformRating += 1;
        
        // Update action tracking
        player.actionsTaken = (player.actionsTaken || 0) + 1;
        player.actionsRemaining = (player.actionsRemaining || 2) - 1;
        game.currentActionCount = (game.currentActionCount || 0) + 1;
        
        io.to(data.gameId).emit('game-updated', game);
      }
    }
  });

  socket.on('skip-action', (data: { gameId: string }) => {
    const game = games.get(data.gameId);
    if (!game) return;

    const player = game.players.find((p: Player) => p.id === socket.id);
    if (!player || game.currentPlayer !== socket.id) return;

    // Skip remaining actions and pass the turn
    player.passed = true;
    player.actionsRemaining = 0;
    
    // Move to next player
    const currentIndex = game.players.findIndex(p => p.id === game.currentPlayer);
    const nextIndex = (currentIndex + 1) % game.players.length;
    game.currentPlayer = game.players[nextIndex].id;
    
    // Reset action tracking for next player
    const nextPlayer = game.players[nextIndex];
    nextPlayer.actionsTaken = 0;
    nextPlayer.actionsRemaining = 2;
    game.currentActionCount = 0;
    
    io.to(data.gameId).emit('game-updated', game);
  });

  socket.on('disconnect', () => {
    console.log('Player disconnected:', socket.id);
    players.delete(socket.id);
  });
});

// REST endpoints
app.get('/health', (req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.get('/games/:id', (req, res) => {
  const game = games.get(req.params.id);
  if (!game) {
    return res.status(404).json({ error: 'Game not found' });
  }
  res.json(game);
});

const PORT = process.env.PORT || 3001;

httpServer.listen(PORT, () => {
  console.log(`Terraforming Mars 3D backend server running on port ${PORT}`);
  console.log(`Demo game available at: /games/demo`);
});