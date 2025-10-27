/**
 * Global Music Manager
 * Manages background music across the entire application
 *
 * Simple two-state system:
 * - MUSIC ON: Landing pages (/, /create, /join) and game lobby
 * - MUSIC OFF: Active gameplay
 *
 * Music continues seamlessly when navigating between landing pages.
 * Music restarts from beginning when returning from active game.
 */

import backgroundMusicService from "./backgroundMusicService";

class GlobalMusicManager {
  private comingFromActiveGame: boolean = false;

  /**
   * Initialize the manager
   * Called once when the app starts
   */
  public init(): void {
    // Manager is ready - no action needed
  }

  /**
   * Start music playback
   * @param restart - If true, restarts from beginning. If false, resumes from current position.
   */
  public async startMusic(restart: boolean = false): Promise<void> {
    const settings = backgroundMusicService.getSettings();

    // If music is already playing and we don't need to restart, do nothing
    if (settings.isPlaying && !restart) {
      return;
    }

    if (restart) {
      await backgroundMusicService.restart();
      this.comingFromActiveGame = false;
    } else {
      await backgroundMusicService.play();
    }
  }

  /**
   * Stop music completely
   * Sets flag to restart music when returning to landing pages
   */
  public stopMusic(): void {
    backgroundMusicService.stop();
    this.comingFromActiveGame = true;
  }

  /**
   * Check if we should restart music (coming from active game)
   */
  public shouldRestart(): boolean {
    return this.comingFromActiveGame;
  }

  /**
   * Get current music playing state
   */
  public isPlaying(): boolean {
    return backgroundMusicService.getSettings().isPlaying;
  }
}

// Singleton instance
export const globalMusicManager = new GlobalMusicManager();
export default globalMusicManager;
