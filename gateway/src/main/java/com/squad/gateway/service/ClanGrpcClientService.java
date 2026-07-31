package com.squad.gateway.service;

import com.squad.clan.grpc.*;
import com.squad.gateway.record.ClanService.*;
import com.squad.gateway.record.ClanService.ApplyToClanResponse;
import com.squad.gateway.record.ClanService.CreateClanResponse;
import com.squad.gateway.record.ClanService.ProcessAcceptanceResponse;
import net.devh.boot.grpc.client.inject.GrpcClient;
import org.springframework.stereotype.Service;

@Service
public class ClanGrpcClientService {

    @GrpcClient("clan-service")
    private ClanServiceGrpc.ClanServiceBlockingStub clanStub;

    public CreateClanResponse createClan(String name, String tag, String leaderSteamId) {
        CreateClanRequest request = CreateClanRequest.newBuilder()
                .setName(name)
                .setTag(tag)
                .setLeaderId(leaderSteamId)
                .build();

        var response = clanStub.createClan(request);

        return new CreateClanResponse(
                response.getClanId(),
                response.getMessage()
        );
    }

    public Clan getClanWithMembers(String clanId) {
        GetClanRequest request = GetClanRequest.newBuilder()
                .setClanId(clanId)
                .build();
        GetClanResponse response = clanStub.getClanWithMembers(request);

        return new Clan(
                response.getId(),
                response.getName(),
                response.getTag(),
                response.getAvatarUrl(),
                response.getDescription(),
                response.getRequirements(),
                response.getIsRecruiting(),
                response.getStatus(),
                response.getTotalElo(),
                response.getMinElo(),
                response.getCreatedAt(),
                response.getMembersList().stream()
                        .map(memeber -> new ClanMember(memeber.getId(), memeber.getUserId(), memeber.getRole()))
                        .toList()
        );
    }

    public ApplyToClanResponse applyToClan(String userId, String clanId, String socialLink, String experienceText) {
        ApplyToClanRequest request = ApplyToClanRequest.newBuilder()
                .setUserId(userId)
                .setClanId(clanId)
                .setSocialLink(socialLink)
                .setExperienceText(experienceText)
                .build();

        com.squad.clan.grpc.ApplyToClanResponse response = clanStub.applyToClan(request);

        return new ApplyToClanResponse(response.getApplicationId(), response.getMessage());
    }

    public ProcessAcceptanceResponse processAcceptance(String applicationId, String acceptorId) {
        ProcessAcceptanceRequest request = ProcessAcceptanceRequest.newBuilder()
                .setApplicationId(applicationId)
                .setAcceptorId(acceptorId)
                .build();

        com.squad.clan.grpc.ProcessAcceptanceResponse response = clanStub.processAcceptance(request);

        ClanMember clanMember = response.hasNewMember() ?
                new ClanMember(
                        response.getNewMember().getId(),
                        response.getNewMember().getUserId(),
                        response.getNewMember().getRole()
                ) : null;

        return new ProcessAcceptanceResponse(clanMember, response.getMessage());
    }
}
