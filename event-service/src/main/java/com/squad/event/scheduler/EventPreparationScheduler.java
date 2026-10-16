package com.squad.event.scheduler;

import com.squad.event.model.enums.EventStatus;
import com.squad.event.repo.EventRepository;
import com.squad.event.service.EventService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.List;
import java.util.UUID;

@Component
@RequiredArgsConstructor
@Slf4j
public class EventPreparationScheduler {

    private final EventRepository eventRepository;
    private final EventService eventService;

    @Scheduled(fixedDelay = 60000)
    public void checkAndStartEventPreparations() {
        log.debug("Run scheduler: checking events for preparation...");

        Instant threshold = Instant.now().plus(15, ChronoUnit.MINUTES);

        List<UUID> eventsToPrepare = eventRepository.findEventIdsWithStatusAndTimeStartThreshold(
                EventStatus.REGISTRATION,
                threshold
        );

        if(eventsToPrepare.isEmpty()) {
            return;
        }

        for(UUID eventId : eventsToPrepare) {
            try {
                eventService.startEventPreparation(eventId);
            } catch (Exception e) {
                log.error("Failed to prepare event: {}. Reason: {}", eventId, e.getMessage(), e);
            }
        }
    }
}
