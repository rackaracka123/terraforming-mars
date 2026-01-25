import React from "react";
import GameIcon from "../../../display/GameIcon.tsx";

interface DiscountLayoutProps {
  behavior: any;
}

const getStandardProjectIcon = (project: string): string | null => {
  const mapping: { [key: string]: string } = {
    "power-plant": "power-tag", // Power tag icon for power plant SP
    "convert-plants-to-greenery": "greenery-tile",
    "convert-heat-to-temperature": "heat",
    aquifer: "ocean-tile",
    asteroid: "temperature",
    "air-scrapping": "venus",
  };
  return mapping[project] || null;
};

const IconWithBadge: React.FC<{
  iconType: string;
  showSpBadge?: boolean;
}> = ({ iconType, showSpBadge = false }) => {
  return (
    <div className="relative inline-flex items-center justify-center">
      <GameIcon iconType={iconType} size="small" />
      {showSpBadge && (
        <span className="absolute -bottom-[2px] -right-[2px] text-[8px] font-black text-white bg-[rgba(80,80,80,0.9)] px-[3px] py-[1px] rounded-[2px] leading-none [text-shadow:0_0_2px_rgba(0,0,0,0.8)]">
          SP
        </span>
      )}
    </div>
  );
};

const DiscountAmount: React.FC<{
  amount: number;
  resourceType: string;
}> = ({ amount, resourceType }) => {
  if (resourceType === "credit") {
    return <GameIcon iconType="credit" amount={-amount} size="small" />;
  }

  return (
    <div className="flex items-center gap-[2px]">
      <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
        -{amount}
      </span>
      <GameIcon iconType={resourceType} size="small" />
    </div>
  );
};

const DiscountRow: React.FC<{
  icons: React.ReactNode;
  amount: number;
  resourceType: string;
}> = ({ icons, amount, resourceType }) => {
  return (
    <div className="flex gap-[3px] items-center justify-center">
      <div className="flex gap-[3px] items-center">{icons}</div>

      <span className="text-base font-bold text-white mx-[3px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
        :
      </span>

      <DiscountAmount amount={amount} resourceType={resourceType} />
    </div>
  );
};

const DiscountLayout: React.FC<DiscountLayoutProps> = ({ behavior }) => {
  if (!behavior.outputs || behavior.outputs.length === 0) return null;

  const discountOutput = behavior.outputs.find((output: any) => output.type === "discount");
  if (!discountOutput) return null;

  const amount = Math.abs(discountOutput.amount ?? 0);
  const affectedTags: string[] = discountOutput.affectedTags || [];
  const affectedStandardProjects: string[] = discountOutput.affectedStandardProjects || [];
  const affectedResources: string[] = discountOutput.affectedResources || [];

  const discountResourceType = affectedResources.length > 0 ? affectedResources[0] : "credit";

  const rows: React.ReactNode[] = [];

  if (affectedTags.length > 0) {
    const tagIcons = affectedTags.map((tag: string, tagIndex: number) => (
      <React.Fragment key={`tag-${tagIndex}`}>
        {tagIndex > 0 && (
          <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
            /
          </span>
        )}
        <IconWithBadge iconType={`${tag.toLowerCase()}-tag`} showSpBadge={false} />
      </React.Fragment>
    ));

    rows.push(
      <DiscountRow
        key="tags"
        icons={tagIcons}
        amount={amount}
        resourceType={discountResourceType}
      />,
    );
  }

  if (affectedStandardProjects.length > 0) {
    const spIcons = affectedStandardProjects.map((project: string, spIndex: number) => {
      const iconType = getStandardProjectIcon(project);
      if (!iconType) return null;

      return (
        <React.Fragment key={`sp-${spIndex}`}>
          {spIndex > 0 && (
            <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
              /
            </span>
          )}
          <IconWithBadge iconType={iconType} showSpBadge={true} />
        </React.Fragment>
      );
    });

    const validIcons = spIcons.filter((icon) => icon !== null);
    if (validIcons.length > 0) {
      rows.push(
        <DiscountRow
          key="standard-projects"
          icons={validIcons}
          amount={amount}
          resourceType={discountResourceType}
        />,
      );
    }
  }

  if (affectedTags.length === 0 && affectedStandardProjects.length === 0) {
    rows.push(
      <DiscountRow
        key="blanket"
        icons={
          <span className="text-[10px] font-semibold text-white bg-[rgba(60,60,60,0.8)] px-1.5 py-0.5 rounded [text-shadow:0_0_2px_rgba(0,0,0,0.6)]">
            All cards
          </span>
        }
        amount={amount}
        resourceType={discountResourceType}
      />,
    );
  }

  if (rows.length === 0) return null;

  if (rows.length === 1) {
    return <>{rows[0]}</>;
  }

  return <div className="flex flex-col gap-1 items-center">{rows}</div>;
};

export default DiscountLayout;
