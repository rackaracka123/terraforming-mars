import React from "react";

const JoinGamePage: React.FC = () => {
  return (
    <div className="join-game-page">
      <div className="container">
        <div className="content">
          <h1>Join Game</h1>
          <p className="subtitle">This feature is coming soon!</p>

          <div className="placeholder-content">
            <p>
              Join an existing terraforming mission by entering a game code or
              selecting from available games.
            </p>
          </div>
        </div>
      </div>

      <style jsx>{`
        .join-game-page {
          background: #000011;
          color: white;
          min-height: 100vh;
          display: flex;
          align-items: center;
          justify-content: center;
          font-family:
            -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen",
            "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue",
            sans-serif;
        }

        .container {
          max-width: 600px;
          width: 100%;
          padding: 40px 20px;
        }

        .content {
          text-align: center;
        }

        h1 {
          font-size: 48px;
          color: #ffffff;
          margin-bottom: 16px;
          text-shadow: 0 2px 4px rgba(0, 0, 0, 0.8);
          font-weight: bold;
        }

        .subtitle {
          font-size: 18px;
          color: rgba(255, 255, 255, 0.7);
          margin-bottom: 40px;
        }

        .placeholder-content {
          background: rgba(255, 255, 255, 0.05);
          border: 1px solid rgba(255, 255, 255, 0.1);
          border-radius: 12px;
          padding: 40px 30px;
          backdrop-filter: blur(10px);
        }

        .placeholder-content p {
          font-size: 16px;
          color: rgba(255, 255, 255, 0.8);
          margin: 0;
          line-height: 1.6;
        }

        @media (max-width: 768px) {
          .container {
            padding: 20px 15px;
          }

          h1 {
            font-size: 36px;
          }

          .subtitle {
            font-size: 16px;
          }

          .placeholder-content {
            padding: 30px 24px;
          }

          .placeholder-content p {
            font-size: 14px;
          }
        }
      `}</style>
    </div>
  );
};

export default JoinGamePage;
