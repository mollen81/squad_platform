package com.squad.event.kafka.event;

import com.squad.event.model.enums.TimelineEvenType;

import java.util.UUID;

public record EventMatchTimelineEvent(
        UUID id,
        UUID matchId,
        TimelineEvenType evenType,
        int gameTimeSeconds
) {}