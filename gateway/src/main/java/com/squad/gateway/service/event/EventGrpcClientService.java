package com.squad.gateway.service.event;

import com.google.protobuf.Timestamp;
import com.squad.event.grpc.CreateEventRequest;
import com.squad.event.grpc.CreateEventResponse;
import com.squad.event.grpc.EventGameMode;
import com.squad.event.grpc.EventServiceGrpc;
import com.squad.gateway.model.dto.request.CreateEventHttpRequest;
import io.grpc.stub.StreamObserver;
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

    public UUID createEvent(CreateEventHttpRequest request) {
        log.info("Received create event request from user {}", request.creatorUserId());
        try {
            CreateEventRequest grpcResuest = CreateEventRequest.newBuilder()
                    .setName(request.name())
                    .setCreatorUserId(request.creatorUserId().toString())
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
            throw new RuntimeException(e);
        }
    }
}
