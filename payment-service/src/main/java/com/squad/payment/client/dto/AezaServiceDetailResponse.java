package com.squad.payment.client.dto;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.math.BigDecimal;

public record AezaServiceDetailResponse(
        @JsonProperty("data") ServiceData data
) {
    public record ServiceData(
            @JsonProperty("status") String status, //active, installing
            @JsonProperty("ip") String ipAddress,
            @JsonProperty("password") String password,
            @JsonProperty("price") BigDecimal price
    ) {}
}
