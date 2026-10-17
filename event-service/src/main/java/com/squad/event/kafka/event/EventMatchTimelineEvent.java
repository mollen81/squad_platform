package com.squad.event.kafka.event;

import com.squad.event.model.enums.TimelineEventType;

import java.util.UUID;

public record EventMatchTimelineEvent(
        UUID id,
        UUID matchId,
        TimelineEventType evenType,
        int gameTimeSeconds
) {}