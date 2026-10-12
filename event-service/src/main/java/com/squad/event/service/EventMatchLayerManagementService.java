package com.squad.event.service;

import com.squad.event.model.domain.Event;
import com.squad.event.model.enums.EventStatus;
import com.squad.event.repo.EventMatchLayerVoteRepository;
import com.squad.event.repo.EventParticipantRepository;
import com.squad.event.repo.EventRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
public class EventMatchLayerManagementService {

    private final EventRepository eventRepository;
    private final EventParticipantRepository eventParticipantRepository;
    private final EventMatchLayerVoteRepository eventMatchLayerVoteRepository;

    @Transactional
    public void mapVote(UUID matchId, UUID userId, String layer) {
        Event event = eventRepository.findByMatchId(matchId)
                .orElseThrow(() -> new IllegalArgumentException("Event for match " + matchId + " is not found"));

        if(event.getStatus() != EventStatus.REGISTRATION) {
            throw new IllegalStateException("Event not in REGISTRATION status");
        }

        if(!eventParticipantRepository.existsByEventIdAndUserId(event.getId(), userId)) {
            throw new IllegalStateException("User " + userId + " is not joined event " + event.getId());
        }

        eventMatchLayerVoteRepository.upsertVote(
                UUID.randomUUID(),
                matchId,
                userId,
                layer);

        log.info("User {} voted for layer {} in match {}", userId, layer, matchId);
    }
}
