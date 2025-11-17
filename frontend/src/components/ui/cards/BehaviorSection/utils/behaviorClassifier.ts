import { CardBehaviorDto } from "@/types/generated/api-types.ts";
import { ClassifiedBehavior } from "../types.ts";

export const classifyBehaviors = (
  behaviors: CardBehaviorDto[],
): ClassifiedBehavior[] => {
  return behaviors.map((behavior) => {
    const hasTrigger = behavior.triggers && behavior.triggers.length > 0;
    const triggerType = hasTrigger ? behavior.triggers?.[0]?.type : null;
    const hasCondition = behavior.triggers?.[0]?.condition !== undefined;
    const hasInputs = behavior.inputs && behavior.inputs.length > 0;
    const hasProduction =
      behavior.outputs &&
      behavior.outputs.some((output: any) =>
        output.type?.includes("production"),
      );

    const hasDiscount =
      behavior.outputs &&
      behavior.outputs.some((output: any) => output.type === "discount");

    const hasPaymentSubstitute =
      behavior.outputs &&
      behavior.outputs.some(
        (output: any) => output.type === "payment-substitute",
      );

    if (hasDiscount) {
      return { behavior, type: "discount" };
    }

    if (hasPaymentSubstitute) {
      return { behavior, type: "payment-substitute" };
    }

    if (triggerType === "manual") {
      return { behavior, type: "manual-action" };
    }

    // Auto triggers with conditions (e.g., placement-bonus-gained) should be triggered-effect
    if (triggerType === "auto" && hasCondition) {
      return { behavior, type: "triggered-effect" };
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
