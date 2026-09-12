package com.squad.stats.kafka;

import com.squad.stats.dto.EventEndedEvent;
import com.squad.stats.service.EventResultsProcessingService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Component;

@Component
@Slf4j
@RequiredArgsConstructor
public class EventEndedListener {

    EventResultsProcessingService processingService;

    @KafkaListener(topics = "event.ended", groupId = "stats-service-group")
    public void onEventEnded(EventEndedEvent event) {
        log.info("Event {} ended. Starting to process results and update users stats.", event.getEventId());
        processingService.processEventResults(event.getEventId());
    }
}
