package com.squad.clan.service;

import com.squad.clan.dto.ClanRequests;
import com.squad.clan.entity.Clan;
import com.squad.clan.entity.ClanApplication;
import com.squad.clan.entity.ClanMember;
import com.squad.clan.facade.ClanFacade;
import com.squad.clan.grpc.*;
import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.server.service.GrpcService;

import java.util.List;
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

            // Facade call, (ELO fetching from stats-service + saving in DB)
            ClanApplication clanApplication = clanFacade.applyToClan(applyToClanDto);

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
        log.info("Requesting information about clan with clanId: {}", request.getClanId());
        try {
            ClanRequests.GetClanDto dto = ClanRequests.GetClanDto.builder()
                    .clanId(UUID.fromString(request.getClanId()))
                    .build();

            Clan clan = clanFacade.getClanWithMembers(dto);

            GetClanResponse response = mapToClanWithMembersResponse(clan);

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }
        catch (IllegalArgumentException e) {
            // Невалидный UUID или клан не найден
            log.warn("Invalid argument in getClanWithMembers: {}", e.getMessage());
            responseObserver.onError(
                    Status.INVALID_ARGUMENT
                            .withDescription(e.getMessage())
                            .asRuntimeException()
            );
        } catch (IllegalStateException e) {
            log.warn("Business rule violation in getClanWithMembers: {}", e.getMessage());
            responseObserver.onError(
                    Status.FAILED_PRECONDITION
                            .withDescription(e.getMessage())
                            .asRuntimeException()
            );
        } catch (Exception e) {
            log.error("Unexpected error in getClanWithMembers", e);
            responseObserver.onError(
                    Status.INTERNAL
                            .withDescription(e.getMessage())
                            .asRuntimeException()
            );
        }
    }


    public void processAcceptance(ProcessAcceptanceRequest request, StreamObserver<ProcessAcceptanceResponse> responseObserver) {
        log.info("Processing the acceptance of application: {}, accepted by user: {}",
                request.getApplicationId(), request.getAcceptorId());
        try {
            ClanRequests.AcceptApplicationDto acceptApplicationDto = ClanRequests.AcceptApplicationDto.builder()
                    .applicationId(UUID.fromString(request.getApplicationId()))
                    .accepterId(UUID.fromString(request.getAcceptorId()))
                    .build();
            ClanMember newMember = clanFacade.acceptApplication(acceptApplicationDto);

            ProcessAcceptanceResponse response = ProcessAcceptanceResponse.newBuilder()
                    .setNewMember(mapToMemberGrpcResponse(newMember))
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }
        catch (Exception e) {
            log.error("Unexpected error in processAcceptance", e);
            responseObserver.onError(
                    Status.INTERNAL
                            .withDescription(e.getMessage())
                            .asRuntimeException()
            );
        }
    }


    // MAPPERS
    private GetClanResponse mapToClanWithMembersResponse(Clan clan) {
        List<ClanMemberDto> memberResponses = clan.getMembers().stream()
                .map(this::mapToMemberGrpcResponse)
                .toList();

        return GetClanResponse.newBuilder()
                .setId(clan.getId().toString())
                .setName(clan.getName())
                .setTag(clan.getTag())
                .setDescription(clan.getDescription() != null ? clan.getDescription() : "")
                .setRequirements(clan.getRequirements() != null ? clan.getRequirements() : "")
                .setAvatarUrl(clan.getAvatar_url() != null ? clan.getAvatar_url() : "")
                .setIsRecruiting(clan.getIsRecruiting() != null ? clan.getIsRecruiting() : true)
                .setStatus(clan.getClanStatus() != null ? clan.getClanStatus().name() : "")
                .setTotalElo(clan.getTotalElo() != null ? clan.getTotalElo() : 0)
                .setMinElo(clan.getMinElo())
                .setCreatedAt(clan.getCreatedAt() != null ? clan.getCreatedAt().toString() : "")
                .addAllMembers(memberResponses)
                .build();
    }


    public ClanMemberDto mapToMemberGrpcResponse(ClanMember clanMember) {
        return ClanMemberDto.newBuilder()
                .setId(clanMember.getId() != null ? clanMember.getId().toString() : "")
                .setUserId(clanMember.getUserId() != null ? clanMember.getUserId().toString() : "")
                .setRole(clanMember.getRole() != null ? clanMember.getRole().toString() : "")
                .build();
    }
}
