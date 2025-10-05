import React from "react";
import { PlayerDto, OtherPlayerDto } from "@/types/generated/api-types.ts";

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
  currentPlayer?: PlayerDto | OtherPlayerDto | null;
  temperatureMilestones?: Milestone[];
  oxygenMilestones?: Milestone[];
}

const RightSidebar: React.FC<RightSidebarProps> = ({
  globalParameters,
  generation,
  currentPlayer: _currentPlayer,
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

  // Get temperature scale markings (every 2 degrees)
  const getTemperatureMarkings = () => {
    const markings = [];
    for (let temp = 8; temp >= -30; temp -= 2) {
      markings.push(temp);
    }
    return markings;
  };

  return (
    <div className="absolute bottom-[20%] right-0 z-10 w-[250px] min-w-[150px] max-w-[250px] h-auto bg-transparent p-[clamp(4px,1vw,8px)_clamp(10px,2vw,20px)] overflow-visible flex flex-col items-center justify-end pointer-events-auto max-xl:min-w-[120px] max-xl:max-w-[180px] max-lg:min-w-[100px] max-lg:max-w-[150px] max-md:w-full max-md:max-w-full max-md:h-auto max-md:border-l-none max-md:border-t max-md:border-t-[rgba(40,50,70,0.6)] max-md:p-[10px]">
      {/* Generation Counter - matching reference design */}
      <div className="mb-[15px] shrink-0">
        <div className="w-9 h-9 bg-gradient-to-br from-[#4a4a4a] via-[#2a2a2a] to-[#1a1a1a] [clip-path:polygon(30%_0%,70%_0%,100%_50%,70%_100%,30%_100%,0%_50%)] flex flex-col items-center justify-center text-white font-bold border border-[#666] shadow-[inset_0_1px_2px_rgba(255,255,255,0.1),0_2px_4px_rgba(0,0,0,0.5)] relative before:content-[''] before:absolute before:top-[2px] before:left-[2px] before:right-[2px] before:bottom-[2px] before:bg-gradient-to-br before:from-[rgba(255,255,255,0.1)] before:to-transparent before:[clip-path:polygon(30%_0%,70%_0%,100%_50%,70%_100%,30%_100%,0%_50%)] before:pointer-events-none max-xl:w-8 max-xl:h-8 max-lg:w-7 max-lg:h-7">
          <div className="text-[8px] leading-none max-xl:text-[7px] max-lg:text-[6px]">
            GEN
          </div>
          <div className="text-base leading-none max-xl:text-sm max-lg:text-xs">
            {generation || 1}
          </div>
        </div>
      </div>

      {/* Separate Meters */}
      <div className="flex flex-col items-center gap-[15px] w-full h-auto max-md:h-auto">
        <div className="flex flex-row items-end gap-10 w-full justify-center pr-0 mb-2.5 max-xl:gap-[30px] max-xl:pr-0 max-lg:gap-[25px] max-lg:pr-0 max-md:flex-row max-md:justify-center max-md:pr-0 max-md:gap-5">
          {/* Oxygen Meter (Left) */}
          <div className="relative h-[50vh] flex flex-col items-center mt-0 max-md:h-[clamp(200px,25vh,300px)] max-md:mt-0">
            <div className="w-5 h-5 rounded-full bg-gradient-to-br from-[#1a1a1a] to-[#2d2d2d] border-2 border-[#444] flex items-center justify-center relative z-[110] mb-[5px]">
              <div className="w-[14px] h-[14px] rounded-full bg-gradient-to-br from-[#006400] to-[#00ff00] shadow-[inset_0_2px_4px_rgba(0,0,0,0.3)]"></div>
            </div>

            <div className="w-[clamp(14px,2vw,18px)] h-[calc(100%-60px)] bg-[linear-gradient(to_right,#1a1a1a_0%,#0a0a0a_50%,#1a1a1a_100%)] border border-[#333] rounded-t-lg relative overflow-visible">
              <div
                className="absolute bottom-0 left-[2px] w-[14px] bg-[linear-gradient(to_top,#006400_0%,#32cd32_50%,#00ff00_100%)] rounded-b-[7px] transition-[height] duration-500 ease-[ease] shadow-[0_0_8px_rgba(0,255,0,1),0_0_15px_rgba(50,205,50,0.8),inset_0_1px_2px_rgba(255,255,255,0.3)] opacity-100 brightness-[1.2]"
                style={{
                  height: `${Math.max(0, ((globalParameters?.oxygen || 0) / 14) * 100)}%`,
                }}
              ></div>

              {/* Internal step markings for oxygen - every single step */}
              <div className="absolute top-0 left-0 right-0 bottom-0 pointer-events-none">
                {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14].map(
                  (oxygen) => (
                    <div
                      key={oxygen}
                      className="absolute left-0 right-0 h-px bg-[rgba(0,255,0,0.3)] border-t border-t-[rgba(0,255,0,0.5)]"
                      style={{
                        bottom: `${(oxygen / 14) * 90}%`,
                      }}
                    ></div>
                  ),
                )}
              </div>

              {/* Oxygen Milestone Indicators */}
              <div className="absolute top-0 left-0 right-0 bottom-0 pointer-events-none">
                {defaultOxygenMilestones.map((milestone, index) => (
                  <div
                    key={index}
                    className="absolute w-5 h-4 bg-[linear-gradient(135deg,rgba(40,40,40,0.95)_0%,rgba(20,20,20,0.9)_100%)] border-2 border-[#666] rounded flex items-center justify-center shadow-[0_0_8px_rgba(0,0,0,0.8),0_2px_4px_rgba(0,0,0,0.6)] -translate-y-1/2 left-[-25px] group max-lg:w-4 max-lg:h-[14px]"
                    style={{ bottom: `${(milestone.value / 14) * 90}%` }}
                    title={`Oxygen Milestone: ${milestone.tooltip}`}
                  >
                    <div className="text-[10px] [filter:drop-shadow(0_1px_1px_rgba(0,0,0,0.5))] flex items-center justify-center max-lg:text-[8px]">
                      <img
                        src={milestone.icon}
                        alt="Oxygen"
                        className="w-3 h-3 object-contain"
                      />
                    </div>
                    <div className="absolute bg-black/90 text-white px-1.5 py-1 rounded-[3px] text-[9px] font-bold whitespace-nowrap opacity-0 pointer-events-none transition-opacity duration-200 z-[1000] border border-[#666] right-[25px] top-1/2 -translate-y-1/2 group-hover:opacity-100">
                      {milestone.tooltip}
                    </div>
                  </div>
                ))}
              </div>

              {/* Internal oxygen numbers */}
              <div className="absolute top-0 left-0 right-0 bottom-0 pointer-events-none">
                {[0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14].map(
                  (oxygen) => (
                    <div
                      key={oxygen}
                      className="absolute w-full flex items-center justify-center text-[10px] font-bold transition-opacity duration-300 -translate-y-1/2 text-[#00ff00] [text-shadow:0_0_3px_rgba(0,255,0,0.8)]"
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

            <div className="text-[8px] font-bold text-[#87ceeb] text-center bg-black/70 px-1 py-0.5 rounded-[3px] border border-[#444]">
              {globalParameters?.oxygen || 0}%
            </div>
          </div>

          {/* Temperature Meter (Right) */}
          <div className="relative h-[50vh] flex flex-col items-center mt-0 max-md:h-[clamp(200px,25vh,300px)] max-md:mt-0">
            <div className="w-5 h-5 rounded-full bg-gradient-to-br from-[#1a1a1a] to-[#2d2d2d] border-2 border-[#444] flex items-center justify-center relative z-[110] mb-[5px]">
              <div className="w-[14px] h-[14px] rounded-full bg-gradient-to-br from-[#87ceeb] to-[#ff8c00] shadow-[inset_0_2px_4px_rgba(0,0,0,0.3)]"></div>
            </div>

            <div className="w-[clamp(14px,2vw,18px)] h-[calc(100%-60px)] bg-[linear-gradient(to_right,#1a1a1a_0%,#0a0a0a_50%,#1a1a1a_100%)] border border-[#333] rounded-t-lg relative overflow-visible">
              <div
                className="absolute bottom-0 left-[2px] w-[14px] bg-[linear-gradient(to_top,#87ceeb_0%,#ffb347_50%,#ff8c00_100%)] rounded-b-[7px] transition-[height] duration-500 ease-[ease] shadow-[0_0_8px_rgba(255,140,0,1),0_0_15px_rgba(255,179,71,0.8),inset_0_1px_2px_rgba(255,255,255,0.3)] opacity-100 brightness-[1.2]"
                style={{
                  height: `${Math.max(0, (((globalParameters?.temperature || -30) + 30) / 38) * 100)}%`,
                }}
              ></div>

              {/* Internal step markings for temperature */}
              <div className="absolute top-0 left-0 right-0 bottom-0 pointer-events-none">
                {getTemperatureMarkings()
                  .filter((temp) => temp !== -30)
                  .map((temp) => (
                    <div
                      key={temp}
                      className="absolute left-0 right-0 h-px bg-[rgba(255,140,0,0.3)] border-t border-t-[rgba(255,140,0,0.5)]"
                      style={{
                        bottom: `${((temp + 30) / 38) * 90}%`,
                      }}
                    ></div>
                  ))}
              </div>

              {/* Temperature Milestone Indicators */}
              <div className="absolute top-0 left-0 right-0 bottom-0 pointer-events-none">
                {defaultTemperatureMilestones.map((milestone, index) => (
                  <div
                    key={index}
                    className="absolute w-5 h-4 bg-[linear-gradient(135deg,rgba(40,40,40,0.95)_0%,rgba(20,20,20,0.9)_100%)] border-2 border-[#666] rounded flex items-center justify-center shadow-[0_0_8px_rgba(0,0,0,0.8),0_2px_4px_rgba(0,0,0,0.6)] -translate-y-1/2 right-[-25px] group max-lg:w-4 max-lg:h-[14px]"
                    style={{ bottom: `${((milestone.value + 30) / 38) * 90}%` }}
                    title={`Temperature Milestone: ${milestone.tooltip}`}
                  >
                    <div className="text-[10px] [filter:drop-shadow(0_1px_1px_rgba(0,0,0,0.5))] flex items-center justify-center max-lg:text-[8px]">
                      <img
                        src={milestone.icon}
                        alt="Temperature"
                        className="w-3 h-3 object-contain"
                      />
                    </div>
                    <div className="absolute bg-black/90 text-white px-1.5 py-1 rounded-[3px] text-[9px] font-bold whitespace-nowrap opacity-0 pointer-events-none transition-opacity duration-200 z-[1000] border border-[#666] left-[25px] top-1/2 -translate-y-1/2 group-hover:opacity-100">
                      {milestone.tooltip}
                    </div>
                  </div>
                ))}
              </div>

              {/* Internal temperature numbers */}
              <div className="absolute top-0 left-0 right-0 bottom-0 pointer-events-none">
                {getTemperatureMarkings().map((temp) => (
                  <div
                    key={temp}
                    className="absolute w-full flex items-center justify-center text-[10px] font-bold transition-opacity duration-300 -translate-y-1/2 text-[#ff8c00] [text-shadow:0_0_3px_rgba(255,140,0,0.8)]"
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

            <div className="text-[8px] font-bold text-[#ff6b2d] text-center bg-black/70 px-1 py-0.5 rounded-[3px] border border-[#444]">
              {globalParameters?.temperature || -30}°C
            </div>
          </div>
        </div>

        {/* Ocean Counter */}
        <div className="flex flex-col items-center gap-1 bg-[linear-gradient(135deg,rgba(0,100,200,0.15)_0%,rgba(0,50,150,0.2)_100%)] border border-[rgba(0,150,255,0.3)] rounded-md p-2 w-4/5 mt-0">
          <div className="text-xs text-[#4da6ff] flex items-center justify-center">
            <img
              src="/assets/hex_blue.png"
              alt="Ocean"
              className="w-4 h-4 object-contain brightness-[1.2]"
            />
          </div>
          <div className="text-[6px] font-bold text-[#4da6ff] uppercase tracking-[0.5px]">
            OCEANS
          </div>
          <div className="flex items-center text-xs font-bold">
            <span className="text-[#00bfff] [text-shadow:0_0_3px_rgba(0,191,255,0.6)]">
              {globalParameters?.oceans || 0}
            </span>
            <span className="text-[#666]"> / </span>
            <span className="text-[#999]">9</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default RightSidebar;
