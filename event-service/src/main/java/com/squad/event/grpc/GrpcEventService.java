package com.squad.event.grpc;

import com.squad.event.model.enums.EventGameMode;
import com.squad.event.service.EventService;
import io.grpc.Status;
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

        try {
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
        catch (Exception e) {
            log.error("Error while processing createEvent gRPC request from user {}", request.getCreatorUserId(), e);
            responseStreamObserver.onError(Status.INTERNAL
                   .withDescription(e.getMessage())
                   .withCause(e.getCause())
                   .asRuntimeException());
        }
    }

}
