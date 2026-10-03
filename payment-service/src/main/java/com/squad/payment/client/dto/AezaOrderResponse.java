package com.squad.payment.client.dto;

import com.fasterxml.jackson.annotation.JsonProperty;

public record AezaOrderResponse(
        @JsonProperty("error") String error,
        @JsonProperty("message") String message,
        @JsonProperty("data") OrderData data
) { 
    public record OrderData(
            @JsonProperty("id") Long serviceId
    ) {}
}
