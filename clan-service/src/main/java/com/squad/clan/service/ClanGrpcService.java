package com.squad.clan.service;

import com.squad.clan.dto.ClanRequests;
import com.squad.clan.entity.Clan;
import com.squad.clan.entity.ClanApplication;
import com.squad.clan.facade.ClanFacade;
import com.squad.clan.grpc.*;
import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.server.service.GrpcService;

import java.util.UUID;

@GrpcService
@Slf4j
@RequiredArgsConstructor
public class ClanGrpcService extends com.squad.clan.grpc.ClanServiceGrpc.ClanServiceImplBase {

    private final ClanFacade clanFacade;

    @Override
    public void createClan(CreateClanRequest request, StreamObserver<CreateClanResponse> responseObserver) {
        log.info("Create clan request is sent from user: {}", request.getLeaderId());
        try {
            // gRPC request -> DTO
            ClanRequests.CreateClanDto dto = ClanRequests.CreateClanDto.builder()
                    .tag(request.getTag())
                    .name(request.getName())
                    .build();

            // Facade call, (ELO fetching from stats-service + saving in DB)
            Clan clan = clanFacade.createClan(dto);

            // Clan -> ClanResponse (Entity -> gPRC response)
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


    @Override
    public void applyToClan(ApplyToClanRequest request, StreamObserver<ApplyToClanResponse> responseObserver) {
        log.info("Application to clan is received from user: {} to clan: {}",
                request.getUserId(), request.getClanId());
        try {
            // gRPC request -> DTO
            ClanRequests.ApplyToClanDto applyToClanDto = ClanRequests.ApplyToClanDto.builder()
                    .clanId(UUID.fromString(request.getClanId()))
                    .userId(UUID.fromString(request.getUserId()))
                    .socialLink(request.getSocialLink())
                    .experienceText(request.getExperienceText())
                    .build();

            // Facade call, ()
            ClanApplication clanApplication =

            ApplyToClanResponse response = ApplyToClanResponse.newBuilder()
                    .setMessage("Application to clan successfully sent")
                    .setApplicationId(clanApplication.getId().toString())
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }
        catch (Exception e) {
            log.error("Internal server error while applying user: {} to clan with clanId: {}",
                    request.getUserId(), request.getClanId());
            responseObserver.onError(Status.INTERNAL
                    .withDescription(e.getMessage())
                    .withCause(e.getCause())
                    .asRuntimeException());
        }
    }


    @Override
    public void getClanWithMembers(GetClanRequest request, StreamObserver<GetClanResponse> responseObserver) {

    }
}
