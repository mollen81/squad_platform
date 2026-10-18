package com.squad.event.kafka.consumer;

import com.squad.event.kafka.event.EventMatchTimelineEvent;
import com.squad.event.kafka.event.MatchFinishedEvent;
import com.squad.event.kafka.event.ServerReadyEvent;
import com.squad.event.service.EventService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
@Slf4j
public class EventKafkaConsumer {
    private final EventService eventService;

    @KafkaListener(topics = "server.ready", groupId = "event-service-group")
    public void consumeServerReadyEvent(ServerReadyEvent readyEvent) {
        try {
            log.info("Consuming message from 'server.ready' for event: {}", readyEvent.eventId());
            eventService.handleServerReady(readyEvent);
        }
        catch (Exception e) {
            log.error("Failed to process server.ready event for event {}: {}", readyEvent.eventId(), e.getMessage(), e);
        }
    }


    @KafkaListener(topics = "match.finished", groupId = "event-service-group")
    public void consumeMatchFinishedEvent(MatchFinishedEvent matchFinishedEvent) {
        try {
            log.info("Consuming message from 'match.finished' for event: {}", matchFinishedEvent.eventId());
            eventService.handleMatchFinished(matchFinishedEvent);
        }
        catch (Exception e) {
            log.error("Failed to process MatchFinishedEvent for event {}: {}",
                    matchFinishedEvent.eventId(), e.getMessage(), e);
        }
    }


    @KafkaListener(topics = "event.match.timeline", groupId = "event-service-group")
    public void consumeEventMatchTimelineEvent(EventMatchTimelineEvent matchTimelineEvent) {
        try {
            log.info("Consuming message from 'event.match.timeline' for eventId {}, match #{}",
                    matchTimelineEvent.eventId(), matchTimelineEvent.matchSequenceNumber());
            eventService.processMatchTimelineEvent(matchTimelineEvent);
        }
        catch (Exception e) {
            log.error("Failed to process EventMatchTimelineEvent for eventId {}, match #{}: {}",
                    matchTimelineEvent.eventId(), matchTimelineEvent.matchSequenceNumber(), e.getMessage(), e);
        }
    }
}
