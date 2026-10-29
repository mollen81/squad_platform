package com.squad.event.service;

import com.squad.event.model.domain.Event;
import com.squad.event.model.domain.EventMatch;
import com.squad.event.model.domain.EventServer;
import com.squad.event.model.domain.EventSide;
import com.squad.event.model.dto.UpdateSideReadyHttpRequest;
import com.squad.event.model.enums.EventMatchStatus;
import com.squad.event.model.enums.EventServerStatus;
import com.squad.event.model.enums.EventStatus;
import com.squad.event.repo.EventMatchRepository;
import com.squad.event.repo.EventRepository;
import com.squad.event.repo.EventServerRepository;
import com.squad.event.repo.EventSideRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.util.List;

@Service
@RequiredArgsConstructor
@Slf4j
public class EventSideManagementService {

    private final EventRepository eventRepository;
    private final EventServerRepository eventServerRepository;
    private final EventSideRepository eventSideRepository;
    private final EventMatchRepository eventMatchRepository;

    @Transactional
    public void setSideReady(UpdateSideReadyHttpRequest request) {
        log.info("updating side ready status: eventId {}, sideId {}, userId {}, isReady {}",
                request.eventId(), request.sideId(), request.userId(), request.isReady());

        Event event = eventRepository.findByIdForUpdate(request.eventId())
                .orElseThrow(() -> new IllegalArgumentException("Event " + request.eventId() + " is not found"));

        if(event.getStatus() == EventStatus.CANCELED || event.getStatus() == EventStatus.FINISHED) {
            throw new IllegalStateException("Event is canceled or finished");
        }

        EventServer server = eventServerRepository.findByEventId(request.eventId())
                .orElseThrow(() -> new IllegalArgumentException("Server for event " + request.eventId() + " is not found"));

        if(server.getStatus() != EventServerStatus.READY) {
            throw new IllegalStateException("Server is not ready yet or match already finished");
        }

        EventSide currentSide = eventSideRepository.findById(request.sideId())
                .orElseThrow(() -> new IllegalStateException("User " + request.userId() + "is not the side leader"));

        if(currentSide.isReady()) {
            return;
        }

        currentSide.setReady(true);
        currentSide.setReadyAt(Instant.now());
        eventSideRepository.save(currentSide);

        List<EventSide> allSides = eventSideRepository.findAllByEventId(request.eventId());
        boolean allReady = allSides.stream().allMatch(EventSide::isReady);

        if(allReady) {
            log.info("All sides are ready. Starting the first match for event {}", request.eventId());

            event.setStatus(EventStatus.LIVE);
            eventRepository.save(event);

            EventMatch firstMatch = eventMatchRepository.findByEventIdAndSequenceNumber(request.eventId(), 1)
                    .orElseThrow(() -> new IllegalStateException("Match 1 does not generated"));

            firstMatch.setStatus(EventMatchStatus.IN_PROGRESS);
            firstMatch.setStartedAt(Instant.now());
            eventMatchRepository.save(firstMatch);
        }
    }
}
