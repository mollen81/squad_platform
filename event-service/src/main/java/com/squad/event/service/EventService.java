package com.squad.event.service;

import com.squad.event.kafka.event.EventMatchTimelineEvent;
import com.squad.event.kafka.event.MatchFinishedEvent;
import com.squad.event.kafka.event.ServerReadyEvent;
import com.squad.event.model.domain.*;
import com.squad.event.model.dto.CreateEventHttpRequest;
import com.squad.event.model.dto.JoinEventHttpRequest;
import com.squad.event.model.dto.MapVoteHttpRequest;
import com.squad.event.model.dto.UpdateSideReadyHttpRequest;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.UUID;

@Service
@Slf4j
@RequiredArgsConstructor
public class EventService {

    private final EventLifecycleManagementService eventLifecycleManagementService;
    private final EventParticipantManagementService eventParticipantManagementService;
    private final EventSideManagementService eventSideManagementService;
    private final EventMatchLayerManagementService eventMatchLayerManagementService;
    private final EventMatchTimelineService eventMatchTimelineService;


    public UUID createEvent(CreateEventHttpRequest request) {
        return eventLifecycleManagementService.createEvent(request);
    }


    public UUID joinEvent(JoinEventHttpRequest request) {
        return eventParticipantManagementService.joinEvent(request);
    }


    public void setSideReady(UpdateSideReadyHttpRequest request) {
        eventSideManagementService.setSideReady(request);
    }


    public void mapVote(MapVoteHttpRequest request) {
        eventMatchLayerManagementService.mapVote(request);
    }


    public void startEventPreparation(UUID eventId) {
        eventLifecycleManagementService.startEventPreparation(eventId);
    }


    public void handleMatchFinished(MatchFinishedEvent matchFinishedEvent) {
        eventLifecycleManagementService.handleMatchFinished(matchFinishedEvent);
    }


    public void handleServerReady(ServerReadyEvent event) {
        eventLifecycleManagementService.handleServerReady(event);
    }

    public void processMatchTimelineEvent(EventMatchTimelineEvent eventMatchTimelineEvent) {
        eventMatchTimelineService.processTimelineEvent(eventMatchTimelineEvent);
    }

    public List<TimelineEvent> getMatchTimelineHistory(UUID eventId, int matchSequenceNumber) {
        return eventMatchTimelineService.getMatchTimelineHistory(eventId, matchSequenceNumber);
    }
}
