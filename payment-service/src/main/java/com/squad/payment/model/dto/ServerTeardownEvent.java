package com.squad.payment.model.dto;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.time.Instant;

public record ServerTeardownEvent(
        @JsonProperty("type") String type, // server.teardown
        @JsonProperty("eventId") String eventId,
        @JsonProperty("server_ip") String serverIp,
        @JsonProperty("reason") String reason,
        @JsonProperty("timestamp") Instant timestamp
) {}