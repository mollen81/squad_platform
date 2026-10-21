package com.squad.event.grpc;

import com.squad.event.model.enums.EventGameMode;
import com.squad.event.service.EventService;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.server.service.GrpcService;

import java.time.Instant;
import java.util.UUID;

@GrpcService
@RequiredArgsConstructor
@Slf4j
public class GrpcEventService extends EventServiceGrpc.EventServiceImplBase {
    private final EventService eventService;

    @Override
    public void createEvent(CreateEventRequest request, StreamObserver<CreateEventResponse> responseStreamObserver) {
        log.info("Starting to process gRPC createEvent request from user {}", request.getCreatorUserId());

        com.squad.event.model.dto.CreateEventRequest createEventRequest = new com.squad.event.model.dto.CreateEventRequest(
                request.getName(),
                UUID.fromString(request.getCreatorUserId()),
                UUID.fromString(request.getSecondLeaderId()),
                request.getTargetGameCount(),
                Instant.ofEpochSecond(request.getTimeStart().getSeconds(), request.getTimeStart().getNanos()),
                EventGameMode.valueOf(request.getGameMode().toString()),
                request.getCreatorSideName(),
                request.getEnemySideName()
        );

        UUID eventId = eventService.createEvent(createEventRequest);

        log.info("gRPC createEvent request from user {} is processed. Event {} created",
                request.getCreatorUserId(), eventId.toString());

        CreateEventResponse response = CreateEventResponse.newBuilder()
                .setEventId(eventId.toString())
                .build();
        responseStreamObserver.onNext(response);
        responseStreamObserver.onCompleted();
    }

    @Override
    public void joinEvent(JoinEventRequest request, StreamObserver<JoinEventResponse> responseStreamObserver) {
        log.info("Starting to process gRPC joinEvent request from user {} to event {}",
                request.getUserId(), request.getEventId());

        com.squad.event.model.dto.JoinEventRequest joinEventRequest = new com.squad.event.model.dto.JoinEventRequest(
                UUID.fromString(request.getUserId()),
                UUID.fromString(request.getEventId()),
                UUID.fromString(request.getSideId())
        );

        UUID eventParticipantId = eventService.joinEvent(joinEventRequest);

        log.info("gRPC joinEvent request from user {} to event {} is processed",
                request.getUserId(), request.getEventId());

        JoinEventResponse response = JoinEventResponse.newBuilder()
                .setEventParticipantId(eventParticipantId.toString())
                .build();
        responseStreamObserver.onNext(response);
        responseStreamObserver.onCompleted();
    }
}
