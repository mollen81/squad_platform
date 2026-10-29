package com.squad.gateway.model.dto.request;

import jakarta.validation.constraints.NotNull;

public record UpdateSideReadyHttpRequest (
        @NotNull(message = "Field 'isReady' must not be null")
        Boolean isReady
) {}