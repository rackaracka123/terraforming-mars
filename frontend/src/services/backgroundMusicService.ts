/**
 * Background music service for managing looping menu/lobby music
 * Separate from audioService which handles one-shot sound effects
 */
import { MUSIC_CONFIG } from "./musicConstants";

class BackgroundMusicService {
  private audio: HTMLAudioElement | null = null;
  private isEnabled: boolean = true;
  private volume: number = MUSIC_CONFIG.DEFAULT_VOLUME;
  private isFading: boolean = false;
  private fadeInterval: NodeJS.Timeout | null = null;

  constructor() {
    this.preloadMusic();
  }

  /**
   * Preload the background music file and setup event listeners
   */
  private preloadMusic() {
    try {
      const audio = new Audio(MUSIC_CONFIG.PATH);
      audio.preload = "auto";
      audio.loop = true;
      audio.volume = 0; // Start at 0 for fade in

      this.setupEventListeners(audio);
      this.audio = audio;
    } catch (error) {
      console.warn("⚠️ Error creating background music element:", error);
    }
  }

  /**
   * Setup audio element event listeners
   */
  private setupEventListeners(audio: HTMLAudioElement) {
    audio.addEventListener("error", (e) => {
      console.warn("⚠️ Failed to load background music", e);
    });
  }

  /**
   * Check if music is currently playing
   */
  private isCurrentlyPlaying(): boolean {
    return this.audio !== null && !this.audio.paused;
  }

  /**
   * Start playing background music with fade in
   */
  public async play(): Promise<void> {
    if (!this.isEnabled || !this.audio) {
      return;
    }

    // Already playing, no action needed
    if (this.isCurrentlyPlaying()) {
      return;
    }

    try {
      this.audio.volume = 0;
      await this.audio.play();
      this.fadeIn();
    } catch (error) {
      console.warn("⚠️ Failed to play background music:", error);
    }
  }

  /**
   * Pause background music without resetting position
   */
  public pause(): void {
    if (!this.audio || !this.isCurrentlyPlaying()) {
      return;
    }

    this.audio.pause();
    this.cancelFade();
  }

  /**
   * Stop playing background music and reset to beginning
   */
  public stop(): void {
    if (!this.audio) {
      return;
    }

    this.audio.pause();
    this.audio.currentTime = 0;
    this.cancelFade();
  }

  /**
   * Restart music from the beginning with fade in
   */
  public async restart(): Promise<void> {
    this.stop();
    await this.play();
  }

  /**
   * Fade in the music volume
   */
  private fadeIn(): void {
    if (!this.audio || this.isFading) {
      return;
    }

    this.cancelFade();
    this.isFading = true;

    const stepDuration =
      MUSIC_CONFIG.FADE_DURATION_MS / MUSIC_CONFIG.FADE_STEPS;
    const volumeIncrement = this.volume / MUSIC_CONFIG.FADE_STEPS;
    let currentStep = 0;

    this.fadeInterval = setInterval(() => {
      if (!this.audio) {
        this.cancelFade();
        return;
      }

      currentStep++;
      this.audio.volume = Math.min(this.volume, currentStep * volumeIncrement);

      if (currentStep >= MUSIC_CONFIG.FADE_STEPS) {
        this.cancelFade();
      }
    }, stepDuration);
  }

  /**
   * Cancel any ongoing fade operation
   */
  private cancelFade(): void {
    if (this.fadeInterval) {
      clearInterval(this.fadeInterval);
      this.fadeInterval = null;
    }
    this.isFading = false;
  }

  /**
   * Set whether background music is enabled
   */
  public setEnabled(enabled: boolean): void {
    this.isEnabled = enabled;
    if (!enabled && this.isCurrentlyPlaying()) {
      this.stop();
    }
  }

  /**
   * Set background music volume (0.0 to 1.0)
   */
  public setVolume(volume: number): void {
    this.volume = Math.max(0, Math.min(1, volume));
    if (this.audio && this.isCurrentlyPlaying() && !this.isFading) {
      this.audio.volume = this.volume;
    }
  }

  /**
   * Get current settings
   */
  public getSettings() {
    return {
      enabled: this.isEnabled,
      volume: this.volume,
      isPlaying: this.isCurrentlyPlaying(),
    };
  }
}

// Singleton instance for application-wide use
export const backgroundMusicService = new BackgroundMusicService();
export default backgroundMusicService;
