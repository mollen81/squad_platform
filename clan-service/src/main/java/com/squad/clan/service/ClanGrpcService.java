package com.squad.clan.service;

import com.squad.clan.dto.ClanRequests;
import com.squad.clan.entity.Clan;
import com.squad.clan.entity.ClanApplication;
import com.squad.clan.grpc.*;
import com.squad.clan.repository.ClanRepository;
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
public class ClanGrpcService extends com.squad.clan.grpc.ClanServiceGrpc.ClanServiceImplBase {

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


    @Override
    public void applyToClan(ApplyToClanRequest request, StreamObserver<ApplyToClanResponse> responseObserver) {
        try {
            ClanRequests.ApplyToClanDto applyToClanDto = ClanRequests.ApplyToClanDto.builder()
                    .clanId(UUID.fromString(request.getClanId()))
                    .userId(UUID.fromString(request.getUserId()))
                    .socialLink(request.getSocialLink())
                    .experienceText(request.getExperienceText())
                    .build();
            ClanApplication clanApplication = clanManagementService.applyToClan(applyToClanDto);

            ApplyToClanResponse response = ApplyToClanResponse.newBuilder()
                    .setMessage("Application to clan successfully sent")
                    .setApplicationId(clanApplication.getId().toString())
                    .build();

            responseObserver.onNext(response);
        }
        catch (Exception e) {
            log.error("Internal server error while applying user: {} to clan with clanId: {}",
                    request.getUserId(), request.getClanId());
        }
    }


}
