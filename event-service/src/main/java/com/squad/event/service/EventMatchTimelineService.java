package com.squad.event.service;

import com.squad.event.grpc.GrpcTimelineStreamManager;
import com.squad.event.kafka.event.EventMatchTimelineEvent;
import com.squad.event.model.domain.Event;
import com.squad.event.model.domain.EventMatch;
import com.squad.event.model.domain.TimelineEvent;
import com.squad.event.repo.EventMatchRepository;
import com.squad.event.repo.EventMatchTimelineRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.UUID;

@Service
@Slf4j
@RequiredArgsConstructor
public class EventMatchTimelineService {
    private final EventMatchRepository eventMatchRepository;
    private final EventMatchTimelineRepository eventMatchTimelineRepository;
    private final GrpcTimelineStreamManager streamManager;

    @Transactional
    public void processTimelineEvent(EventMatchTimelineEvent event) {
        EventMatch eventMatch = eventMatchRepository.findByEventIdAndMatchId(
                        event.eventId(), event.matchId())
                .orElseThrow(() -> new IllegalArgumentException("Match " + event.matchId() +
                        " for event " + event.eventId() + " is not found"));

        TimelineEvent timelineEvent = TimelineEvent.builder()
                .matchId(eventMatch.getId())
                .timelineEventType(event.timelineEventType())
                .gameTimeSeconds(event.gameTimeSeconds())
                .payload(event.payload())
                .build();

        TimelineEvent savedEntry = eventMatchTimelineRepository.save(timelineEvent);

        streamManager.broadcastTimelineEvent(event.eventId(), savedEntry);

        log.debug("Processed timeline event {} for event {}, match {}",
                event.timelineEventType(), event.eventId(), event.matchId());
    }


    @Transactional
    public List<TimelineEvent> getMatchTimelineHistory(UUID eventId, UUID matchId) {
        EventMatch eventMatch = eventMatchRepository.findByEventIdAndMatchId(eventId, matchId)
                .orElseThrow(() -> new IllegalArgumentException(
                        "Match " + matchId + " not found for eventId " + eventId));
        return eventMatchTimelineRepository.findAllByMatchIdOrderByTime(eventMatch.getId());
    }
}
