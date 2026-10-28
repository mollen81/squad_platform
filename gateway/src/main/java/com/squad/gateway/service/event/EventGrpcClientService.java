package com.squad.gateway.service.event;

import com.google.protobuf.Timestamp;
import com.squad.event.grpc.*;
import com.squad.gateway.model.dto.request.CreateEventHttpRequest;
import com.squad.gateway.model.dto.request.JoinEventHttpRequest;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.client.inject.GrpcClient;
import org.springframework.stereotype.Service;

import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
public class EventGrpcClientService {
    @GrpcClient("event-service")
    private EventServiceGrpc.EventServiceBlockingStub eventServiceBlockingStub;

    public UUID createEvent(CreateEventHttpRequest request, UUID creatorUserId) {
        log.info("Received create event request from user {}", creatorUserId);
        try {
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
        catch (Exception e) {
            throw new RuntimeException("Error while create event request: " + e.getMessage(), e);
        }
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


}
