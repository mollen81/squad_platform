package com.squad.event.kafka.event;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.UUID;

public record MatchFinishedEvent(
        @JsonProperty("event_id") UUID eventId,
        @JsonProperty("match_sequence_number") int matchSequenceNumber,
        @JsonProperty("winner_side_id") UUID winnerSideId
) {}