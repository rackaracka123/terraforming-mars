import React from "react";
import { PlayerDto } from "@/types/generated/api-types.ts";
import styles from "./RightSidebar.module.css";

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
  currentPlayer?: PlayerDto | null;
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
    <div className={styles.rightSidebar}>
      {/* Generation Counter - matching reference design */}
      <div className={styles.generationCounter}>
        <div className={styles.generationHex}>
          <div className={styles.genText}>GEN</div>
          <div className={styles.genNumber}>{generation || 1}</div>
        </div>
      </div>

      {/* Separate Meters */}
      <div className={styles.globalParameters}>
        <div className={styles.metersContainer}>
          {/* Oxygen Meter (Left) */}
          <div className={styles.oxygenMeter}>
            <div className={styles.oxygenBulb}>
              <div className={styles.oxygenBulbInner}></div>
            </div>

            <div className={styles.oxygenTube}>
              <div
                className={styles.oxygenFill}
                style={{
                  height: `${Math.max(0, ((globalParameters?.oxygen || 0) / 14) * 100)}%`,
                }}
              ></div>

              {/* Internal step markings for oxygen - every single step */}
              <div className={styles.oxygenSteps}>
                {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14].map(
                  (oxygen) => (
                    <div
                      key={oxygen}
                      className={styles.oxygenStepMark}
                      style={{
                        bottom: `${(oxygen / 14) * 90}%`,
                      }}
                    ></div>
                  ),
                )}
              </div>

              {/* Oxygen Milestone Indicators */}
              <div className={styles.oxygenMilestones}>
                {defaultOxygenMilestones.map((milestone, index) => (
                  <div
                    key={index}
                    className={`${styles.milestoneIndicator} ${styles.oxygenMilestone}`}
                    style={{ bottom: `${(milestone.value / 14) * 90}%` }}
                    title={`Oxygen Milestone: ${milestone.tooltip}`}
                  >
                    <div className={styles.milestoneIcon}>
                      <img
                        src={milestone.icon}
                        alt="Oxygen"
                        className={styles.milestoneIconImg}
                      />
                    </div>
                    <div className={styles.milestoneTooltip}>
                      {milestone.tooltip}
                    </div>
                  </div>
                ))}
              </div>

              {/* Internal oxygen numbers */}
              <div className={styles.oxygenInternalNumbers}>
                {[0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14].map(
                  (oxygen) => (
                    <div
                      key={oxygen}
                      className={styles.oxygenInternalNumber}
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

            <div className={styles.currentOxygen}>
              {globalParameters?.oxygen || 0}%
            </div>
          </div>

          {/* Temperature Meter (Right) */}
          <div className={styles.temperatureMeter}>
            <div className={styles.temperatureBulb}>
              <div className={styles.temperatureBulbInner}></div>
            </div>

            <div className={styles.thermometerTube}>
              <div
                className={styles.temperatureFill}
                style={{
                  height: `${Math.max(0, (((globalParameters?.temperature || -30) + 30) / 38) * 100)}%`,
                }}
              ></div>

              {/* Internal step markings for temperature */}
              <div className={styles.temperatureSteps}>
                {getTemperatureMarkings()
                  .filter((temp) => temp !== -30)
                  .map((temp) => (
                    <div
                      key={temp}
                      className={styles.temperatureStepMark}
                      style={{
                        bottom: `${((temp + 30) / 38) * 90}%`,
                      }}
                    ></div>
                  ))}
              </div>

              {/* Temperature Milestone Indicators */}
              <div className={styles.temperatureMilestones}>
                {defaultTemperatureMilestones.map((milestone, index) => (
                  <div
                    key={index}
                    className={`${styles.milestoneIndicator} ${styles.temperatureMilestone}`}
                    style={{ bottom: `${((milestone.value + 30) / 38) * 90}%` }}
                    title={`Temperature Milestone: ${milestone.tooltip}`}
                  >
                    <div className={styles.milestoneIcon}>
                      <img
                        src={milestone.icon}
                        alt="Temperature"
                        className={styles.milestoneIconImg}
                      />
                    </div>
                    <div className={styles.milestoneTooltip}>
                      {milestone.tooltip}
                    </div>
                  </div>
                ))}
              </div>

              {/* Internal temperature numbers */}
              <div className={styles.temperatureInternalNumbers}>
                {getTemperatureMarkings().map((temp) => (
                  <div
                    key={temp}
                    className={styles.temperatureInternalNumber}
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

            <div className={styles.currentTemp}>
              {globalParameters?.temperature || -30}°C
            </div>
          </div>
        </div>

        {/* Ocean Counter */}
        <div className={styles.oceanCounter}>
          <div className={styles.oceanIcon}>
            <img
              src="/assets/hex_blue.png"
              alt="Ocean"
              className={styles.oceanIconImg}
            />
          </div>
          <div className={styles.oceanLabel}>OCEANS</div>
          <div className={styles.oceanCount}>
            <span className={styles.currentOceans}>
              {globalParameters?.oceans || 0}
            </span>
            <span className={styles.oceanSeparator}> / </span>
            <span className={styles.maxOceans}>9</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default RightSidebar;
