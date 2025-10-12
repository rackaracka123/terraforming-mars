import { CardBehaviorDto } from "@/types/generated/api-types.ts";
import { ClassifiedBehavior } from "../types.ts";

export const classifyBehaviors = (
  behaviors: CardBehaviorDto[],
): ClassifiedBehavior[] => {
  return behaviors.map((behavior) => {
    const hasTrigger = behavior.triggers && behavior.triggers.length > 0;
    const triggerType = hasTrigger ? behavior.triggers?.[0]?.type : null;
    const hasInputs = behavior.inputs && behavior.inputs.length > 0;
    const hasProduction =
      behavior.outputs &&
      behavior.outputs.some((output: any) =>
        output.type?.includes("production"),
      );

    const hasDiscount =
      behavior.outputs &&
      behavior.outputs.some((output: any) => output.type === "discount");

    if (hasDiscount) {
      return { behavior, type: "discount" };
    }

    if (triggerType === "manual") {
      return { behavior, type: "manual-action" };
    }

    if (triggerType === "auto" && !hasInputs) {
      return { behavior, type: "auto-no-background" };
    }

    if (hasTrigger && hasInputs) {
      return { behavior, type: "triggered-effect" };
    }

    if (hasProduction && (!hasTrigger || triggerType === "auto")) {
      return { behavior, type: "immediate-production" };
    }

    return { behavior, type: "immediate-effect" };
  });
};
