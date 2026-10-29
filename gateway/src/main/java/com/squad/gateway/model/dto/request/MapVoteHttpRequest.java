package com.squad.gateway.model.dto.request;

import jakarta.validation.constraints.NotBlank;


public record MapVoteHttpRequest(
        @NotBlank(message = "Layer name is required")
        String layer
) {}