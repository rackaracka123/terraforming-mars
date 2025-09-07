import React from "react";
import { PlayerDto } from "../../../types/generated/api-types.ts";

interface GlobalParameters {
  temperature: number;
  oxygen: number;
  oceans: number;
}

interface Milestone {
  value: number;
  icon: string;
  tooltip: string;
  reward?: string;
}

interface RightSidebarProps {
  globalParameters?: GlobalParameters;
  generation?: number;
  currentPlayer?: PlayerDto;
  temperatureMilestones?: Milestone[];
  oxygenMilestones?: Milestone[];
}

const RightSidebar: React.FC<RightSidebarProps> = ({
  globalParameters,
  generation,
  currentPlayer,
  temperatureMilestones,
  oxygenMilestones,
}) => {
  // Set default values
  const defaultTemperatureMilestones = temperatureMilestones || [
    {
      value: -8,
      icon: "/assets/global-parameters/temperature.png",
      tooltip: "-8°C: +1 TR",
      reward: "+1 TR",
    },
  ];
  const defaultOxygenMilestones = oxygenMilestones || [
    {
      value: 8,
      icon: "/assets/global-parameters/oxygen.png",
      tooltip: "8%: +1 TR",
      reward: "+1 TR",
    },
  ];
  // Helper function to create milestone objects easily
  // const createMilestone = (
  //   value: number,
  //   icon: string,
  //   reward: string = "+1 TR",
  // ): Milestone => ({
  //   value,
  //   icon,
  //   tooltip: `${value}${value < 0 || value > 20 ? "°C" : "%"}: ${reward}`,
  //   reward,
  // });

  // Example usage for multiple milestones:
  // temperatureMilestones = [
  //   createMilestone(-24, '❄️', '+1 TR'),
  //   createMilestone(-8, '🌡️', '+1 TR'),
  //   createMilestone(0, '🔥', '+1 TR')
  // ]
  // oxygenMilestones = [
  //   createMilestone(8, '🫁', '+1 TR'),
  //   createMilestone(14, '🌬️', '+2 TR')
  // ]

  // Mock global parameter milestone rewards
  // const getGlobalParameterRewards = () => {
  //   const rewards: { [key: string]: string[] } = {
  //     temperature: [
  //       "-24°C: +1 TR",
  //       "-20°C: +1 TR",
  //       "-16°C: +1 TR",
  //       "-12°C: +1 TR",
  //       "-8°C: +1 TR",
  //       "-4°C: +1 TR",
  //       "0°C: +1 TR",
  //       "+4°C: +1 TR",
  //       "+8°C: +2 TR",
  //     ],
  //     oxygen: [
  //       "1%: +1 TR",
  //       "2%: +1 TR",
  //       "3%: +1 TR",
  //       "4%: +1 TR",
  //       "5%: +1 TR",
  //       "6%: +1 TR",
  //       "7%: +1 TR",
  //       "8%: +1 TR",
  //       "9%: +1 TR",
  //       "10%: +1 TR",
  //       "11%: +1 TR",
  //       "12%: +1 TR",
  //       "13%: +1 TR",
  //       "14%: +2 TR",
  //     ],
  //     oceans: [
  //       "1st Ocean: +1 TR",
  //       "2nd Ocean: +1 TR",
  //       "3rd Ocean: +1 TR",
  //       "4th Ocean: +1 TR",
  //       "5th Ocean: +1 TR",
  //       "6th Ocean: +1 TR",
  //       "7th Ocean: +1 TR",
  //       "8th Ocean: +1 TR",
  //       "9th Ocean: +2 TR",
  //     ],
  //   };
  //   return rewards;
  // };

  // Get temperature scale markings (every 2 degrees)
  const getTemperatureMarkings = () => {
    const markings = [];
    for (let temp = 8; temp >= -30; temp -= 2) {
      markings.push(temp);
    }
    return markings;
  };

  return (
    <div className="right-sidebar">
      {/* Generation Counter - matching reference design */}
      <div className="generation-counter">
        <div className="generation-hex">
          <div className="gen-text">GEN</div>
          <div className="gen-number">{generation || 1}</div>
        </div>
      </div>

      {/* Separate Meters */}
      <div className="global-parameters">
        <div className="meters-container">
          {/* Oxygen Meter (Left) */}
          <div className="oxygen-meter">
            <div className="oxygen-bulb">
              <div className="bulb-inner oxygen-bulb-inner"></div>
            </div>

            <div className="oxygen-tube">
              <div
                className="oxygen-fill"
                style={{
                  height: `${Math.max(0, ((globalParameters?.oxygen || 0) / 14) * 100)}%`,
                }}
              ></div>

              {/* Internal step markings for oxygen - every single step */}
              <div className="oxygen-steps">
                {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14].map(
                  (oxygen) => (
                    <div
                      key={oxygen}
                      className="oxygen-step-mark"
                      style={{
                        bottom: `${(oxygen / 14) * 90}%`,
                      }}
                    ></div>
                  ),
                )}
              </div>

              {/* Oxygen Milestone Indicators */}
              <div className="oxygen-milestones">
                {defaultOxygenMilestones.map((milestone, index) => (
                  <div
                    key={index}
                    className="milestone-indicator oxygen-milestone"
                    style={{ bottom: `${(milestone.value / 14) * 90}%` }}
                    title={`Oxygen Milestone: ${milestone.tooltip}`}
                  >
                    <div className="milestone-icon">
                      <img
                        src={milestone.icon}
                        alt="Oxygen"
                        className="milestone-icon-img"
                      />
                    </div>
                    <div className="milestone-tooltip">{milestone.tooltip}</div>
                  </div>
                ))}
              </div>

              {/* Internal oxygen numbers */}
              <div className="oxygen-internal-numbers">
                {[0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14].map(
                  (oxygen) => (
                    <div
                      key={oxygen}
                      className="oxygen-internal-number"
                      style={{
                        bottom: `${(oxygen / 14) * 90}%`,
                        opacity:
                          (globalParameters?.oxygen || 0) > oxygen ? 0 : 1,
                      }}
                    >
                      {oxygen}
                    </div>
                  ),
                )}
              </div>
            </div>

            <div className="current-oxygen">
              {globalParameters?.oxygen || 0}%
            </div>
          </div>

          {/* Temperature Meter (Right) */}
          <div className="temperature-meter">
            <div className="temperature-bulb">
              <div className="bulb-inner temperature-bulb-inner"></div>
            </div>

            <div className="thermometer-tube">
              <div
                className="temperature-fill"
                style={{
                  height: `${Math.max(0, (((globalParameters?.temperature || -30) + 30) / 38) * 100)}%`,
                }}
              ></div>

              {/* Internal step markings for temperature */}
              <div className="temperature-steps">
                {getTemperatureMarkings()
                  .filter((temp) => temp !== -30)
                  .map((temp) => (
                    <div
                      key={temp}
                      className="temperature-step-mark"
                      style={{
                        bottom: `${((temp + 30) / 38) * 90}%`,
                      }}
                    ></div>
                  ))}
              </div>

              {/* Temperature Milestone Indicators */}
              <div className="temperature-milestones">
                {defaultTemperatureMilestones.map((milestone, index) => (
                  <div
                    key={index}
                    className="milestone-indicator temperature-milestone"
                    style={{ bottom: `${((milestone.value + 30) / 38) * 90}%` }}
                    title={`Temperature Milestone: ${milestone.tooltip}`}
                  >
                    <div className="milestone-icon">
                      <img
                        src={milestone.icon}
                        alt="Temperature"
                        className="milestone-icon-img"
                      />
                    </div>
                    <div className="milestone-tooltip">{milestone.tooltip}</div>
                  </div>
                ))}
              </div>

              {/* Internal temperature numbers */}
              <div className="temperature-internal-numbers">
                {getTemperatureMarkings().map((temp) => (
                  <div
                    key={temp}
                    className="temperature-internal-number"
                    style={{
                      bottom: `${((temp + 30) / 38) * 90}%`,
                      opacity:
                        (globalParameters?.temperature || -30) > temp ? 0 : 1,
                    }}
                  >
                    {temp}
                  </div>
                ))}
              </div>
            </div>

            <div className="current-temp">
              {globalParameters?.temperature || -30}°C
            </div>
          </div>
        </div>

        {/* Ocean Counter */}
        <div className="ocean-counter">
          <div className="ocean-icon">
            <img
              src="/assets/hex_blue.png"
              alt="Ocean"
              className="ocean-icon-img"
            />
          </div>
          <div className="ocean-label">OCEANS</div>
          <div className="ocean-count">
            <span className="current-oceans">
              {globalParameters?.oceans || 0}
            </span>
            <span className="ocean-separator"> / </span>
            <span className="max-oceans">9</span>
          </div>
        </div>
      </div>

      {/* Player Score Section */}
      <div className="player-score-section">
        {currentPlayer && (
          <div className="score-hex-container">
            <div className="score-hex">
              <div className="score-value">
                {currentPlayer.terraformRating || 20}
              </div>
            </div>
            <div className="player-name">{currentPlayer.name}</div>
          </div>
        )}
      </div>

      <style>{`
        .right-sidebar {
          width: 100%;
          min-width: 150px;
          max-width: 250px;
          height: calc(100vh - 120px);
          background: linear-gradient(180deg, 
            rgba(5, 10, 20, 0.95) 0%, 
            rgba(10, 15, 30, 0.95) 50%, 
            rgba(5, 10, 25, 0.95) 100%);
          border-left: 1px solid rgba(40, 50, 70, 0.6);
          padding: clamp(4px, 1vw, 8px) clamp(10px, 2vw, 20px);
          overflow: visible;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: flex-start;
          position: relative;
          box-shadow: inset 1px 0 2px rgba(100, 150, 200, 0.1);
        }
        
        .generation-counter {
          margin-bottom: 15px;
          flex-shrink: 0;
        }
        
        .generation-hex {
          width: 36px;
          height: 36px;
          background: linear-gradient(135deg, #4a4a4a 0%, #2a2a2a 50%, #1a1a1a 100%);
          clip-path: polygon(30% 0%, 70% 0%, 100% 50%, 70% 100%, 30% 100%, 0% 50%);
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          color: #fff;
          font-weight: bold;
          border: 1px solid #666;
          box-shadow: 
            inset 0 1px 2px rgba(255, 255, 255, 0.1),
            0 2px 4px rgba(0, 0, 0, 0.5);
          position: relative;
        }
        
        .generation-hex::before {
          content: '';
          position: absolute;
          top: 2px;
          left: 2px;
          right: 2px;
          bottom: 2px;
          background: linear-gradient(135deg, rgba(255, 255, 255, 0.1) 0%, transparent 50%);
          clip-path: polygon(30% 0%, 70% 0%, 100% 50%, 70% 100%, 30% 100%, 0% 50%);
          pointer-events: none;
        }
        
        .gen-text {
          font-size: 8px;
          line-height: 1;
        }
        
        .gen-number {
          font-size: 16px;
          line-height: 1;
        }
        
        .global-parameters {
          flex: 1;
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 15px;
          width: 100%;
          height: calc(100vh - 220px);
        }
        
        .meters-container {
          display: flex;
          flex-direction: row;
          align-items: flex-start;
          gap: 30px;
          flex: 1;
          width: 100%;
          justify-content: center;
          padding-right: 80px;
        }
        
        .oxygen-meter, .temperature-meter {
          position: relative;
          height: clamp(300px, 40vh, 450px);
          display: flex;
          flex-direction: column;
          align-items: center;
          margin-top: clamp(8px, 2vh, 15px);
        }
        
        /* Dual Thermometer Styles */
        .temperature-bulb {
          width: 20px;
          height: 20px;
          border-radius: 50%;
          background: linear-gradient(135deg, #1a1a1a 0%, #2d2d2d 100%);
          border: 2px solid #444;
          display: flex;
          align-items: center;
          justify-content: center;
          position: relative;
          z-index: 110;
          margin-bottom: 5px;
        }
        
        .temperature-bulb-inner {
          width: 14px;
          height: 14px;
          border-radius: 50%;
          background: linear-gradient(135deg, #87ceeb 0%, #ff8c00 100%);
          box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.3);
        }
        
        .thermometer-tube, .oxygen-tube {
          width: clamp(14px, 2vw, 18px);
          height: calc(100% - 60px);
          background: linear-gradient(to right, #1a1a1a 0%, #0a0a0a 50%, #1a1a1a 100%);
          border: 1px solid #333;
          border-radius: 8px 8px 0 0;
          position: relative;
          overflow: visible;
        }
        
        .temperature-fill {
          position: absolute;
          bottom: 0;
          left: 2px;
          width: 14px;
          background: linear-gradient(to top, #87ceeb 0%, #ffb347 50%, #ff8c00 100%);
          border-radius: 0 0 7px 7px;
          transition: height 0.5s ease;
          box-shadow: 
            0 0 8px rgba(255, 140, 0, 1),
            0 0 15px rgba(255, 179, 71, 0.8),
            inset 0 1px 2px rgba(255, 255, 255, 0.3);
          opacity: 1;
          filter: brightness(1.2);
        }
        
        .oxygen-fill {
          position: absolute;
          bottom: 0;
          left: 2px;
          width: 14px;
          background: linear-gradient(to top, #006400 0%, #32cd32 50%, #00ff00 100%);
          border-radius: 0 0 7px 7px;
          transition: height 0.5s ease;
          box-shadow: 
            0 0 8px rgba(0, 255, 0, 1),
            0 0 15px rgba(50, 205, 50, 0.8),
            inset 0 1px 2px rgba(255, 255, 255, 0.3);
          opacity: 1;
          filter: brightness(1.2);
        }
        
        .oxygen-bulb {
          width: 20px;
          height: 20px;
          border-radius: 50%;
          background: linear-gradient(135deg, #1a1a1a 0%, #2d2d2d 100%);
          border: 2px solid #444;
          display: flex;
          align-items: center;
          justify-content: center;
          position: relative;
          z-index: 110;
          margin-bottom: 5px;
        }
        
        .oxygen-bulb-inner {
          width: 14px;
          height: 14px;
          border-radius: 50%;
          background: linear-gradient(135deg, #006400 0%, #00ff00 100%);
          box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.3);
        }
        
        /* Internal Step Markings */
        .oxygen-steps, .temperature-steps {
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          pointer-events: none;
        }
        
        .oxygen-step-mark {
          position: absolute;
          left: 0;
          right: 0;
          height: 1px;
          background: rgba(0, 255, 0, 0.3);
          border-top: 1px solid rgba(0, 255, 0, 0.5);
        }
        
        .temperature-step-mark {
          position: absolute;
          left: 0;
          right: 0;
          height: 1px;
          background: rgba(255, 140, 0, 0.3);
          border-top: 1px solid rgba(255, 140, 0, 0.5);
        }
        
        .temperature-scale {
          position: absolute;
          right: -30px;
          top: 0;
          height: 100%;
          width: 25px;
        }
        
        /* Internal Numbers */
        .oxygen-internal-numbers, .temperature-internal-numbers {
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          pointer-events: none;
        }
        
        .oxygen-internal-number, .temperature-internal-number {
          position: absolute;
          width: 100%;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 10px;
          font-weight: bold;
          transition: opacity 0.3s ease;
          transform: translateY(-50%);
        }
        
        .oxygen-internal-number {
          color: #00ff00;
          text-shadow: 0 0 3px rgba(0, 255, 0, 0.8);
        }
        
        .temperature-internal-number {
          color: #ff8c00;
          text-shadow: 0 0 3px rgba(255, 140, 0, 0.8);
        }
        
        /* Milestone Indicators */
        .oxygen-milestones, .temperature-milestones {
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          pointer-events: none;
        }
        
        .milestone-indicator {
          position: absolute;
          width: 20px;
          height: 16px;
          background: linear-gradient(135deg, 
            rgba(40, 40, 40, 0.95) 0%, 
            rgba(20, 20, 20, 0.9) 100%);
          border: 2px solid #666;
          border-radius: 4px;
          display: flex;
          align-items: center;
          justify-content: center;
          box-shadow: 
            0 0 8px rgba(0, 0, 0, 0.8),
            0 2px 4px rgba(0, 0, 0, 0.6);
          transform: translateY(-50%);
        }
        
        .oxygen-milestone {
          left: -25px;
        }
        
        .temperature-milestone {
          right: -25px;
        }
        
        .milestone-icon {
          font-size: 10px;
          filter: drop-shadow(0 1px 1px rgba(0, 0, 0, 0.5));
          display: flex;
          align-items: center;
          justify-content: center;
        }
        
        .milestone-icon-img {
          width: 12px;
          height: 12px;
          object-fit: contain;
        }
        
        .milestone-tooltip {
          position: absolute;
          background: rgba(0, 0, 0, 0.9);
          color: white;
          padding: 4px 6px;
          border-radius: 3px;
          font-size: 9px;
          font-weight: bold;
          white-space: nowrap;
          opacity: 0;
          pointer-events: none;
          transition: opacity 0.2s ease;
          z-index: 1000;
          border: 1px solid #666;
        }
        
        .oxygen-milestone .milestone-tooltip {
          right: 25px;
          top: 50%;
          transform: translateY(-50%);
        }
        
        .temperature-milestone .milestone-tooltip {
          left: 25px;
          top: 50%;
          transform: translateY(-50%);
        }
        
        .milestone-indicator:hover .milestone-tooltip {
          opacity: 1;
        }
        
        .current-values {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 4px;
        }
        
        .current-temp {
          font-size: 8px;
          font-weight: bold;
          color: #ff6b2d;
          text-align: center;
          background: rgba(0, 0, 0, 0.7);
          padding: 2px 4px;
          border-radius: 3px;
          border: 1px solid #444;
        }
        
        .current-oxygen {
          font-size: 8px;
          font-weight: bold;
          color: #87ceeb;
          text-align: center;
          background: rgba(0, 0, 0, 0.7);
          padding: 2px 4px;
          border-radius: 3px;
          border: 1px solid #444;
        }
        
        /* Ocean Counter Styles */
        .ocean-counter {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 4px;
          background: linear-gradient(135deg, rgba(0, 100, 200, 0.15) 0%, rgba(0, 50, 150, 0.2) 100%);
          border: 1px solid rgba(0, 150, 255, 0.3);
          border-radius: 6px;
          padding: 6px;
          width: 70%;
          margin-top: -5px;
        }
        
        .ocean-icon {
          font-size: 12px;
          color: #4da6ff;
          display: flex;
          align-items: center;
          justify-content: center;
        }
        
        .ocean-icon-img {
          width: 16px;
          height: 16px;
          object-fit: contain;
          filter: brightness(1.2);
        }
        
        .ocean-label {
          font-size: 6px;
          font-weight: bold;
          color: #4da6ff;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }
        
        .ocean-count {
          display: flex;
          align-items: center;
          font-size: 12px;
          font-weight: bold;
        }
        
        .current-oceans {
          color: #00bfff;
          text-shadow: 0 0 3px rgba(0, 191, 255, 0.6);
        }
        
        .ocean-separator {
          color: #666;
        }
        
        .max-oceans {
          color: #999;
        }
        
        

        .player-score-section {
          flex-shrink: 0;
          margin-top: 20px;
        }
        
        .score-hex-container {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 8px;
        }
        
        .score-hex {
          width: 42px;
          height: 42px;
          background: linear-gradient(135deg, #1e90ff 0%, #0066cc 50%, #004d99 100%);
          clip-path: polygon(30% 0%, 70% 0%, 100% 50%, 70% 100%, 30% 100%, 0% 50%);
          display: flex;
          align-items: center;
          justify-content: center;
          color: #fff;
          font-weight: bold;
          border: 2px solid #0a4d7a;
          box-shadow: 
            inset 0 1px 2px rgba(255, 255, 255, 0.2),
            0 0 10px rgba(30, 144, 255, 0.6),
            0 2px 6px rgba(0, 0, 0, 0.4);
          position: relative;
        }
        
        .score-hex::before {
          content: '';
          position: absolute;
          top: 3px;
          left: 3px;
          right: 3px;
          bottom: 3px;
          background: linear-gradient(135deg, rgba(255, 255, 255, 0.15) 0%, transparent 50%);
          clip-path: polygon(30% 0%, 70% 0%, 100% 50%, 70% 100%, 30% 100%, 0% 50%);
          pointer-events: none;
        }
        
        .score-value {
          font-size: 14px;
          line-height: 1;
        }
        
        .player-name {
          font-size: 8px;
          color: #4a90e2;
          text-align: center;
          max-width: 80px;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }

        @media (max-width: 1200px) {
          .right-sidebar {
            min-width: 120px;
            max-width: 180px;
          }
          
          .meters-container {
            gap: 20px;
            padding-right: 60px;
          }
          
          .generation-hex {
            width: 32px;
            height: 32px;
          }
          
          .gen-text {
            font-size: 7px;
          }
          
          .gen-number {
            font-size: 14px;
          }
        }

        @media (max-width: 900px) {
          .right-sidebar {
            min-width: 100px;
            max-width: 150px;
          }
          
          .meters-container {
            gap: 15px;
            padding-right: 40px;
          }
          
          .generation-hex {
            width: 28px;
            height: 28px;
          }
          
          .gen-text {
            font-size: 6px;
          }
          
          .gen-number {
            font-size: 12px;
          }
          
          .milestone-indicator {
            width: 16px;
            height: 14px;
          }
          
          .milestone-icon {
            font-size: 8px;
          }
        }

        @media (max-width: 768px) {
          .right-sidebar {
            width: 100%;
            max-width: 100%;
            height: auto;
            border-left: none;
            border-top: 1px solid rgba(40, 50, 70, 0.6);
            padding: 10px;
          }
          
          .global-parameters {
            height: auto;
          }
          
          .meters-container {
            flex-direction: row;
            justify-content: center;
            padding-right: 0;
            gap: 30px;
          }
          
          .oxygen-meter, .temperature-meter {
            height: clamp(200px, 25vh, 300px);
            margin-top: 0;
          }
          
          .milestones-awards-section {
            margin: 10px 0;
            padding: 10px;
          }
          
          .indicators-grid {
            grid-template-columns: repeat(3, 1fr);
            gap: 6px;
          }
          
          .score-hex {
            width: 36px;
            height: 36px;
          }
          
          .score-value {
            font-size: 12px;
          }
        }
      `}</style>
    </div>
  );
};

export default RightSidebar;
