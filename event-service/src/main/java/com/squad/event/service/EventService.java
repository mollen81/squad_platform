package com.squad.event.service;

import com.squad.event.model.domain.Event;
import com.squad.event.model.domain.EventParticipant;
import com.squad.event.model.domain.EventSide;
import com.squad.event.model.dto.CreateEventRequest;
import com.squad.event.model.dto.JoinEventRequest;
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
                .name(request.name())
                .creatorUserId(request.creatorUserId())
                .targetGameCount(request.targetGameCount())
                .timeStart(request.timeStart())
                .status(EventStatus.DRAFT)
                .build();

        eventRepository.save(event);

        EventSide side1 = EventSide.builder()
                .eventId(event.getId())
                .name(request.creatorSideName())
                .leaderUserId(request.creatorUserId())
                .isReady(false)
                .build();

        EventSide side2 = EventSide.builder()
                .eventId(event.getId())
                .name(request.enemySideName())
                .leaderUserId(request.secondLeaderId())
                .isReady(false)
                .build();

        eventSideRepository.save(side1);
        eventSideRepository.save(side2);

        return event.getId();
    }

    @Transactional
    public UUID joinEvent(JoinEventRequest request) {
        Event event = eventRepository.findByIdForUpdate(request.eventId())
                .orElseThrow(() -> new IllegalArgumentException("Event " + request.eventId() + " is not found"));

        if(event.getStatus() != EventStatus.REGISTRATION) {
            throw new IllegalStateException("Event registration is closed. Current status: " + event.getStatus());
        }

        EventSide side = eventSideRepository.findById(request.sideId())
                .orElseThrow(() -> new IllegalArgumentException("Side " + request.sideId() + " is not found"));

        if(!side.getEventId().equals(event.getId())) {
            throw new IllegalArgumentException("Side " + side.getId() + " does not belong to event " + event.getId());
        }

        if(eventParticipantRepository.existsByEventIdAndUserId(event.getId(), request.userId())) {
            throw new IllegalStateException("User " + request.userId() + " is already registered to event " + event.getId());
        }

        int currentSidePlayers = eventParticipantRepository.countBySideId(side.getId());
        if(currentSidePlayers >= MAX_PLAYERS_PER_SIDE) {
            throw new IllegalStateException("Side " + side.getId() + " is full: current players count = " + currentSidePlayers);
        }

        EventParticipant participant = EventParticipant.builder()
                .eventId(request.eventId())
                .sideId(request.sideId())
                .userId(request.userId())
                .clanId(request.clanId())
                .build();

        eventParticipantRepository.save(participant);

        log.info("User {} is joined to event {} for side {}",
                request.userId(), request.eventId(), request.sideId());

        return participant.getId();
    }
}
