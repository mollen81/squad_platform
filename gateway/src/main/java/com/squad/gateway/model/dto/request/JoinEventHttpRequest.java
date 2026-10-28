package com.squad.gateway.model.dto.request;

import jakarta.validation.constraints.NotNull;

import java.util.UUID;

public record JoinEventHttpRequest (
        @NotNull(message = "eventId is required")
        UUID eventId,

        @NotNull(message = "sideId is required")
        UUID sideId
) {}