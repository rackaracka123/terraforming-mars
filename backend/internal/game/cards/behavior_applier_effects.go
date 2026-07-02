package cards

import (
	"context"
	"fmt"
	"log/slog"

	"terraforming-mars-backend/internal/game/shared"
)

func (a *BehaviorApplier) applyEffectOutput(ctx context.Context, o *shared.EffectCondition, amount int, log *slog.Logger) error {
	switch o.ResourceType {
	case shared.ResourcePaymentSubstitute:
		if a.player == nil {
			return fmt.Errorf("cannot apply payment substitute: no player context")
		}
		resources := GetResourcesFromSelectors(o.Selectors)
		if len(resources) > 0 {
			resourceType := shared.ResourceType(resources[0])
			a.player.Resources().AddPaymentSubstitute(resourceType, amount)
			log.Debug("Added payment substitute",
				slog.String("resource_type", string(resourceType)), slog.Int("conversion_rate", amount))
		} else {
			log.Warn("payment-substitute output missing selectors with resources")
		}

	case shared.ResourceDiscount:
		log.Debug("Discount effect registered", slog.Int("amount", amount), slog.Any("selectors", o.Selectors))

	case shared.ResourceGlobalParameterLenience:
		log.Debug("Global parameter lenience effect registered", slog.Int("amount", amount), slog.String("temporary", o.Temporary))

	case shared.ResourceIgnoreGlobalRequirements:
		log.Debug("Ignore global requirements effect registered", slog.String("temporary", o.Temporary))

	case shared.ResourceValueModifier:
		if a.player == nil {
			return fmt.Errorf("cannot apply value modifier: no player context")
		}
		for _, resourceStr := range GetResourcesFromSelectors(o.Selectors) {
			resourceType := shared.ResourceType(resourceStr)
			a.player.Resources().AddValueModifier(resourceType, amount)
			log.Debug("Added resource value modifier",
				slog.String("resource_type", string(resourceType)), slog.Int("modifier_amount", amount))
		}

	case shared.ResourceStoragePaymentSubstitute:
		if a.player == nil {
			return fmt.Errorf("cannot apply storage payment substitute: no player context")
		}
		if a.sourceCardID == "" {
			log.Warn("storage-payment-substitute output missing source card ID")
			return nil
		}
		storageResourceType := shared.ResourceFloater
		if a.cardRegistry != nil {
			if sourceCard, err := a.cardRegistry.GetByID(a.sourceCardID); err == nil && sourceCard.ResourceStorage != nil {
				storageResourceType = sourceCard.ResourceStorage.Type
			}
		}
		targetResource := shared.ResourceCredit
		resources := GetResourcesFromSelectors(o.Selectors)
		if len(resources) > 0 {
			targetResource = shared.ResourceType(resources[0])
		}
		a.player.Resources().AddStoragePaymentSubstitute(shared.StoragePaymentSubstitute{
			CardID:         a.sourceCardID,
			ResourceType:   storageResourceType,
			ConversionRate: amount,
			TargetResource: targetResource,
			Selectors:      o.Selectors,
		})
		log.Debug("Added storage payment substitute",
			slog.String("card_id", a.sourceCardID), slog.String("resource_type", string(storageResourceType)),
			slog.String("target_resource", string(targetResource)), slog.Int("conversion_rate", amount))

	case shared.ResourceOceanAdjacencyBonus:
		log.Debug("Ocean adjacency bonus effect registered", slog.Int("amount", amount))

	case shared.ResourceDefense:
		log.Debug("Defense effect registered", slog.Int("amount", amount), slog.Any("selectors", o.Selectors))

	case shared.ResourceActionReuse:
		log.Debug("Skipping action-reuse output (handled at action layer)")

	case shared.ResourceEffect:
		log.Debug("Effect registered", slog.Int("amount", amount))

	case shared.ResourceTag:
		log.Debug("Tag effect registered", slog.Int("amount", amount))

	default:
		log.Warn("Unhandled effect type", slog.String("type", string(o.ResourceType)))
	}
	return nil
}
