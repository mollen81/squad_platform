package com.squad.gateway.service.event;

import com.google.protobuf.Timestamp;
import com.squad.event.grpc.*;
import com.squad.gateway.model.dto.request.CreateEventHttpRequest;
import com.squad.gateway.model.dto.request.JoinEventHttpRequest;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.client.inject.GrpcClient;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
public class EventGrpcClientService {
    @GrpcClient("event-service")
    private EventServiceGrpc.EventServiceBlockingStub eventServiceBlockingStub;

    public UUID createEvent(CreateEventHttpRequest request, UUID creatorUserId) {
        log.info("Received create event request from user {}", creatorUserId);

        CreateEventRequest grpcResuest = CreateEventRequest.newBuilder()
                .setName(request.name())
                .setCreatorUserId(creatorUserId.toString())
                .setSecondLeaderId(request.secondLeaderId().toString())
                .setTargetGameCount(request.targetGameCount())
                .setTimeStart(Timestamp.newBuilder()
                        .setSeconds(request.timeStart().getEpochSecond())
                        .setNanos(request.timeStart().getNano())
                        .build())
                .setGameMode(EventGameMode.valueOf(request.gameMode()))
                .setCreatorSideName(request.creatorSideName())
                .setEnemySideName(request.enemySideName())
                .build();

        CreateEventResponse response = eventServiceBlockingStub.createEvent(grpcResuest);
        return UUID.fromString(response.getEventId());
    }

    public UUID joinEvent(JoinEventHttpRequest request, UUID userId) {
        log.info("Received join event request from user {}", userId);

        JoinEventRequest grpcRequest = JoinEventRequest.newBuilder()
                .setUserId(userId.toString())
                .setEventId(request.eventId().toString())
                .setSideId(request.sideId().toString())
                .build();

        JoinEventResponse response = eventServiceBlockingStub.joinEvent(grpcRequest);
        return UUID.fromString(response.getEventParticipantId());
    }

    public void setSideReady(UUID eventId, UUID sideId, UUID userId, Boolean isReady) {
        log.info("Received set side ready request: eventId {}, sideId {}, userId {}, isReady {}",
                eventId, sideId, userId, isReady);

        UpdateSideReadyRequest grpcRequest = UpdateSideReadyRequest.newBuilder()
                .setEventId(eventId.toString())
                .setSideId(sideId.toString())
                .setUserId(userId.toString())
                .setIsReady(isReady)
                .build();

        eventServiceBlockingStub.updateSideReady(grpcRequest);
    }

    public void mapVote(UUID eventId, UUID matchId, UUID userId, String layer) {
        log.info("Received map vote request from user {}, eventId {}, match {}, layer {}",
                userId, eventId, matchId, layer);

        MapVoteRequest grpcRequest = MapVoteRequest.newBuilder()
                .setEventId(eventId.toString())
                .setMatchId(matchId.toString())
                .setUserId(userId.toString())
                .setLayer(layer)
                .build();

        eventServiceBlockingStub.mapVote(grpcRequest);
    }

    public List<>
}
