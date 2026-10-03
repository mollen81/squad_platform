package com.squad.payment.client.dto;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.Map;

public record AezaOrderRequest(
        @JsonProperty("tariffId") int tariffId,
        @JsonProperty("term") String term,
        @JsonProperty("parameters")Map<String, Object> parameters
) {}

