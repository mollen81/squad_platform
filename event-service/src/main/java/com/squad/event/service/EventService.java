package com.squad.event.service;

import com.squad.event.model.domain.Event;
import com.squad.event.model.domain.EventSide;
import com.squad.event.model.dto.CreateEventRequest;
import com.squad.event.model.enums.EventStatus;
import com.squad.event.repo.EventParticipantRepository;
import com.squad.event.repo.EventRepository;
import com.squad.event.repo.EventSideRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.UUID;

@Service
@Slf4j
@RequiredArgsConstructor
public class EventService {

    private final EventRepository eventRepository;
    private final EventSideRepository eventSideRepository;
    private final EventParticipantRepository eventParticipantRepository;

    private static final int MAX_PLAYERS_PER_SIDE = 50;

    @Transactional
    public UUID createEvent(CreateEventRequest request) {
        log.info("Creating new event: {} by user {}", request.name(), request.creatorUserId());

        Event event = Event.builder()
                .id(UUID.randomUUID())
                .name(request.name())
                .creatorUserId(request.creatorUserId())
                .targetGameCount(request.targetGameCount())
                .timeStart(request.timeStart())
                .status(EventStatus.DRAFT)
                .build();

        eventRepository.save(event);

        EventSide side1 = EventSide.builder()
                .id(UUID.randomUUID())
                .eventId(event.getId())
                .name(request.creatorSideName())
                .leaderUserId(request.creatorUserId())
                .isReady(false)
                .build();

        EventSide side2 = EventSide.builder()
                .id(UUID.randomUUID())
                .eventId(event.getId())
                .name(request.enemySideName())
                .leaderUserId(request.secondLeaderId())
                .isReady(false)
                .build();

        eventSideRepository.save(side1);
        eventSideRepository.save(side2);

        return event.getId();
    }
}
