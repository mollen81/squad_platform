package com.squad.event.kafka.consumer;

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


    @KafkaListener(topics = "")
}
