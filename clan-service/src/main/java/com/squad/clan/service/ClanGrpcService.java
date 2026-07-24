package com.squad.clan.service;

import com.squad.clan.dto.ClanRequests;
import com.squad.clan.entity.Clan;
import com.squad.clan.repository.ClanRepository;
import com.squad.grpc.clan.CreateClanRequest;
import com.squad.grpc.clan.CreateClanResponse;
import com.squad.grpc.clan.GetClanRequest;
import com.squad.grpc.clan.GetClanResponse;
import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.server.service.GrpcService;

import java.util.Optional;
import java.util.UUID;

@GrpcService
@Slf4j
@RequiredArgsConstructor
public class ClanGrpcService extends com.squad.grpc.clan.ClanServiceGrpc.ClanServiceImplBase {

    private final ClanManagementService clanManagementService;
    private final ClanRepository clanRepository;

    @Override
    public void createClan(CreateClanRequest request, StreamObserver<CreateClanResponse> responseObserver) {
        try {
            log.info("Create clan request is sent from user: {}", request.getLeaderId());
            ClanRequests.CreateClanDto createClanDto = ClanRequests.CreateClanDto.builder()
                    .tag(request.getTag())
                    .name(request.getName())
                    .build();
            Clan clan = clanManagementService.createClan(createClanDto);

            CreateClanResponse response = CreateClanResponse.newBuilder()
                    .setClanId(clan.getId().toString())
                    .setMessage("Clan successfully created")
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }
        catch (Exception e) {
            log.error("Internal server error while creating clan", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription(e.getMessage())
                    .withCause(e.getCause())
                    .asRuntimeException());
        }
    }
}
