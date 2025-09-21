/* tslint:disable */
/* eslint-disable */
// Generated using typescript-generator version 3.2.1263 on 2025-09-21 21:08:47.

export interface BuildAquiferAction {
    type: ActionType;
    hexPosition: HexPositionDto;
}

export interface BuildAquiferCommand extends WebSocketCommand {
    gameId: string;
    playerId: string;
    hexPosition: HexPositionDto;
}

export interface BuildCityAction {
    type: ActionType;
    hexPosition: HexPositionDto;
}

export interface BuildCityCommand extends WebSocketCommand {
    gameId: string;
    playerId: string;
    hexPosition: HexPositionDto;
}

export interface BuildPowerPlantAction {
    type: ActionType;
}

export interface BuildPowerPlantCommand extends WebSocketCommand {
    gameId: string;
    playerId: string;
}

export interface CardBehaviorDto {
    triggers: CardTrigger[] | null;
    inputs: CardOutput[] | null;
    outputs: CardOutput[] | null;
    choices: CardChoice[] | null;
}

export interface CardChoice {
    inputs: CardOutput[] | null;
    outputs: CardOutput[] | null;
}

export interface CardDto {
    id: string;
    name: string;
    type: CardType;
    cost: number;
    description: string;
    tags: CardTag[] | null;
    requirements: CardRequirement[] | null;
    behaviors: CardBehaviorDto[] | null;
    resourceStorage: ResourceStorageDto | null;
    vpConditions: VictoryPointCondition[] | null;
}

export interface CardOutput {
    type: string;
    amount: number | null;
    player: string | null;
}

export interface CardRequirement {
    type: string;
    min: number | null;
    max: number | null;
    resource: string | null;
}

export interface CardTrigger {
    type: string;
}

export interface CorporationDto {
    id: string;
    name: string;
    description: string;
    startingCredits: number;
    startingResources: ResourceSet;
    startingProduction: ResourceSet;
    tags: CardTag[];
    specialEffects: string[];
    number: string;
}

export interface CreateGameRequest {
    maxPlayers: number;
    playerName: string;
}

export interface CreateGameResponse {
    game: GameDto;
}

export interface ErrorEvent extends WebSocketEvent {
    message: string;
    code: string | null;
}

export interface ErrorPayload {
    message: string;
    code: string | null;
}

export interface ErrorResponse {
    error: string;
    code: string | null;
    details: string | null;
}

export interface FullStateEvent extends WebSocketEvent {
    game: GameDto;
    playerId: string;
}

export interface FullStatePayload {
    game: GameDto;
    playerId: string;
}

export interface GameDto {
    id: string;
    status: GameStatus;
    settings: GameSettingsDto;
    hostPlayerId: string;
    currentPhase: GamePhase;
    globalParameters: GlobalParametersDto;
    currentPlayer: PlayerDto;
    otherPlayers: OtherPlayerDto[];
    viewingPlayerId: string;
    currentTurn: string | null;
    generation: number;
    remainingActions: number;
    turnOrder: string[];
}

export interface GameSettingsDto {
    maxPlayers: number;
}

export interface GameUpdatedEvent extends WebSocketEvent {
    game: GameDto;
}

export interface GameUpdatedPayload {
    game: GameDto;
}

export interface GetGameResponse {
    game: GameDto;
}

export interface GetPlayerResponse {
    player: PlayerDto;
}

export interface GlobalParametersDto {
    temperature: number;
    oxygen: number;
    oceans: number;
}

export interface HexPositionDto {
    q: number;
    r: number;
    s: number;
}

export interface JoinGameRequest {
    playerName: string;
}

export interface JoinGameResponse {
    game: GameDto;
    playerId: string;
}

export interface LaunchAsteroidAction {
    type: ActionType;
}

export interface LaunchAsteroidCommand extends WebSocketCommand {
    gameId: string;
    playerId: string;
}

export interface ListCardsResponse {
    cards: CardDto[];
    totalCount: number;
    offset: number;
    limit: number;
}

export interface ListGamesResponse {
    games: GameDto[];
}

export interface OtherPlayerDto {
    id: string;
    name: string;
    corporation: string;
    handCardCount: number;
    resources: ResourcesDto;
    resourceProduction: ProductionDto;
    terraformRating: number;
    playedCards: string[];
    passed: boolean;
    availableActions: number;
    victoryPoints: number;
    selectingCards: boolean;
    connected: boolean;
}

export interface PlantGreeneryAction {
    type: ActionType;
    hexPosition: HexPositionDto;
}

export interface PlantGreeneryCommand extends WebSocketCommand {
    gameId: string;
    playerId: string;
    hexPosition: HexPositionDto;
}

export interface PlayCardAction {
    cardId: string;
}

export interface PlayCardCommand extends WebSocketCommand {
    gameId: string;
    playerId: string;
    cardId: string;
}

export interface PlayerConnectCommand extends WebSocketCommand {
    playerName: string;
    gameId: string;
}

export interface PlayerConnectPayload {
    playerName: string;
    gameId: string;
    playerId: string | null;
}

export interface PlayerConnectedEvent extends WebSocketEvent {
    playerId: string;
    playerName: string;
    game: GameDto;
}

export interface PlayerConnectedPayload {
    playerId: string;
    playerName: string;
    game: GameDto;
}

export interface PlayerDisconnectedEvent extends WebSocketEvent {
    playerId: string;
    playerName: string;
    game: GameDto;
}

export interface PlayerDisconnectedPayload {
    playerId: string;
    playerName: string;
    game: GameDto;
}

export interface PlayerDto {
    id: string;
    name: string;
    corporation: string | null;
    cards: CardDto[];
    resources: ResourcesDto;
    resourceProduction: ProductionDto;
    terraformRating: number;
    playedCards: string[];
    passed: boolean;
    availableActions: number;
    victoryPoints: number;
    productionSelection: ProductionPhaseDto | null;
    startingSelection: CardDto[];
    connected: boolean;
}

export interface PlayerProductionData {
    playerId: string;
    playerName: string;
    beforeResources: ResourcesDto;
    afterResources: ResourcesDto;
    resourceDelta: ResourceDelta;
    production: ProductionDto;
    terraformRating: number;
    energyConverted: number;
    creditsIncome: number;
}

export interface PlayerReconnectedEvent extends WebSocketEvent {
    playerId: string;
    playerName: string;
    game: GameDto;
}

export interface PlayerReconnectedPayload {
    playerId: string;
    playerName: string;
    game: GameDto;
}

export interface ProductionDto {
    credits: number;
    steel: number;
    titanium: number;
    plants: number;
    energy: number;
    heat: number;
}

export interface ProductionPhaseDto {
    availableCards: CardDto[];
    selectionComplete: boolean;
}

export interface ProductionPhaseStartedEvent extends WebSocketEvent {
    generation: number;
    playersData: PlayerProductionData[];
    game: GameDto;
}

export interface ProductionPhaseStartedPayload {
    generation: number;
    playersData: PlayerProductionData[];
    game: GameDto;
}

export interface ResourceDelta {
    credits: number;
    steel: number;
    titanium: number;
    plants: number;
    energy: number;
    heat: number;
}

export interface ResourceSet {
    credits: number;
    steel: number;
    titanium: number;
    plants: number;
    energy: number;
    heat: number;
}

export interface ResourceStorageDto {
    type: string;
    capacity: number | null;
    starting: number;
}

export interface ResourcesDto {
    credits: number;
    steel: number;
    titanium: number;
    plants: number;
    energy: number;
    heat: number;
}

export interface SelectCardsCommand extends WebSocketCommand {
    gameId: string;
    playerId: string;
    cardIds: string[];
}

export interface SelectStartingCardAction {
    type: ActionType;
    cardIds: string[];
}

export interface SelectStartingCardCommand extends WebSocketCommand {
    gameId: string;
    playerId: string;
    cardIds: string[];
}

export interface SellPatentsAction {
    type: ActionType;
    cardCount: number;
}

export interface SellPatentsCommand extends WebSocketCommand {
    gameId: string;
    playerId: string;
    cardCount: number;
}

export interface SkipAction {
    type: ActionType;
}

export interface SkipActionCommand extends WebSocketCommand {
    gameId: string;
    playerId: string;
}

export interface StartGameAction {
    type: ActionType;
}

export interface StartGameCommand extends WebSocketCommand {
    gameId: string;
    playerId: string;
}

export interface UpdatePlayerResourcesResponse {
    player: PlayerDto;
}

export interface UpdateResourcesRequest {
    resources: ResourcesDto;
}

export interface VictoryPointCondition {
    amount: number;
    condition: string;
    maxTrigger: number | null;
    per: VictoryPointPer | null;
}

export interface VictoryPointPer {
    type: string;
    amount: number | null;
    location: string | null;
}

export interface WebSocketCommand {
    gameId: string | null;
    playerId: string | null;
}

export interface WebSocketEvent {
}

export interface WebSocketMessage {
    type: MessageType;
    payload: string;
    gameId: string | null;
}

export const enum ActionType {
    SELECT_STARTING_CARD = "SELECT_STARTING_CARD",
    SELECT_CARDS = "SELECT_CARDS",
    START_GAME = "START_GAME",
    SKIP_ACTION = "SKIP_ACTION",
    PLAY_CARD = "PLAY_CARD",
    SELL_PATENTS = "SELL_PATENTS",
    BUILD_POWER_PLANT = "BUILD_POWER_PLANT",
    LAUNCH_ASTEROID = "LAUNCH_ASTEROID",
    BUILD_AQUIFER = "BUILD_AQUIFER",
    PLANT_GREENERY = "PLANT_GREENERY",
    BUILD_CITY = "BUILD_CITY",
}

export const enum CardTag {
    SPACE = "SPACE",
    EARTH = "EARTH",
    SCIENCE = "SCIENCE",
    POWER = "POWER",
    BUILDING = "BUILDING",
    MICROBE = "MICROBE",
    ANIMAL = "ANIMAL",
    PLANT = "PLANT",
    EVENT = "EVENT",
    CITY = "CITY",
    VENUS = "VENUS",
    JOVIAN = "JOVIAN",
    WILDLIFE = "WILDLIFE",
    WILD = "WILD",
}

export const enum CardType {
    AUTOMATED = "AUTOMATED",
    ACTIVE = "ACTIVE",
    EVENT = "EVENT",
    CORPORATION = "CORPORATION",
    PRELUDE = "PRELUDE",
}

export const enum GamePhase {
    WAITING_FOR_GAME_START = "WAITING_FOR_GAME_START",
    STARTING_CARD_SELECTION = "STARTING_CARD_SELECTION",
    START_GAME_SELECTION = "START_GAME_SELECTION",
    ACTION = "ACTION",
    PRODUCTION_AND_CARD_DRAW = "PRODUCTION_AND_CARD_DRAW",
    COMPLETE = "COMPLETE",
}

export const enum GameStatus {
    LOBBY = "LOBBY",
    ACTIVE = "ACTIVE",
    COMPLETED = "COMPLETED",
}

export const enum MessageType {
    PLAYER_CONNECT = "PLAYER_CONNECT",
    GAME_UPDATED = "GAME_UPDATED",
    PLAYER_CONNECTED = "PLAYER_CONNECTED",
    PLAYER_RECONNECTED = "PLAYER_RECONNECTED",
    PLAYER_DISCONNECTED = "PLAYER_DISCONNECTED",
    ERROR = "ERROR",
    FULL_STATE = "FULL_STATE",
    PRODUCTION_PHASE_STARTED = "PRODUCTION_PHASE_STARTED",
    ACTION_SELL_PATENTS = "ACTION_SELL_PATENTS",
    ACTION_LAUNCH_ASTEROID = "ACTION_LAUNCH_ASTEROID",
    ACTION_BUILD_POWER_PLANT = "ACTION_BUILD_POWER_PLANT",
    ACTION_BUILD_AQUIFER = "ACTION_BUILD_AQUIFER",
    ACTION_PLANT_GREENERY = "ACTION_PLANT_GREENERY",
    ACTION_BUILD_CITY = "ACTION_BUILD_CITY",
    ACTION_START_GAME = "ACTION_START_GAME",
    ACTION_SKIP_ACTION = "ACTION_SKIP_ACTION",
    ACTION_PLAY_CARD = "ACTION_PLAY_CARD",
    ACTION_SELECT_STARTING_CARD = "ACTION_SELECT_STARTING_CARD",
    ACTION_SELECT_CARDS = "ACTION_SELECT_CARDS",
}
