#!/bin/bash
# Comprehensive fix for all restructured package imports

# Fix all Go files
find . -name "*.go" -type f -exec sed -i '
    s|"terraforming-mars-backend/internal/game/parameters"|"terraforming-mars-backend/internal/session/gameplay/mechanics/parameters"|g
    s|"terraforming-mars-backend/internal/game/resources"|"terraforming-mars-backend/internal/session/gameplay/mechanics/resources"|g
    s|"terraforming-mars-backend/internal/game/tiles"|"terraforming-mars-backend/internal/session/gameplay/mechanics/tiles"|g
    s|"terraforming-mars-backend/internal/game/production"|"terraforming-mars-backend/internal/session/gameplay/mechanics/production"|g
    s|"terraforming-mars-backend/internal/game/turn"|"terraforming-mars-backend/internal/session/gameplay/mechanics/turn"|g
    s|"terraforming-mars-backend/internal/game/actions/standard_projects"|"terraforming-mars-backend/internal/session/gameplay/actions/standard_projects"|g
    s|"terraforming-mars-backend/internal/game/actions/card_selection"|"terraforming-mars-backend/internal/session/gameplay/actions/card_selection"|g
    s|"terraforming-mars-backend/internal/game/actions"|"terraforming-mars-backend/internal/session/gameplay/actions"|g
    s|"terraforming-mars-backend/internal/service"|"terraforming-mars-backend/internal/session/gameplay/services"|g
    s|"terraforming-mars-backend/internal/cards"|"terraforming-mars-backend/internal/session/cards"|g
    s|"terraforming-mars-backend/internal/lobby"|"terraforming-mars-backend/internal/session/lobby"|g
' {} \;

echo "Fixed all import paths"
