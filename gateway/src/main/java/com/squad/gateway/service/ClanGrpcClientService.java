package com.squad.gateway.service;

import com.squad.clan.grpc.*;
import net.devh.boot.grpc.client.inject.GrpcClient;
import org.springframework.stereotype.Service;

@Service
public class ClanGrpcClientService {

    @GrpcClient("clan-service")
    private ClanServiceGrpc.ClanServiceBlockingStub clanStub;

    public CreateClanResponse createClan(String name, String tag, String leaderId, String description,
            String requirements, String avatarUrl) {
        try {
            CreateClanRequest request = CreateClanRequest.newBuilder()
                    .setName(name)
                    .setTag(tag)
                    .setLeaderId(leaderId)
                    .setDescription(description)
                    .setRequirements(requirements)
                    .setAvatarUrl(avatarUrl != null ? avatarUrl : "")
                    .build();
            return clanStub.createClan(request);
        }
        catch (Exception e) {
            throw new RuntimeException("Create clan request could not be processed: " + e.getMessage(), e);
        }
    }

    public GetClanResponse getClanWithMembers(String clanId) {
        try {
            GetClanRequest request = GetClanRequest.newBuilder()
                    .setClanId(clanId)
                    .build();
            return clanStub.getClanWithMembers(request);
        }
        catch (Exception e) {
            throw new RuntimeException("Get clan request could not be processed: " + e.getMessage(), e);
        }
    }

    public ApplyToClanResponse applyToClan(String userId, String clanId, String socialLink, String experienceText) {
        try {
            ApplyToClanRequest request = ApplyToClanRequest.newBuilder()
                    .setUserId(userId)
                    .setClanId(clanId)
                    .setSocialLink(socialLink)
                    .setExperienceText(experienceText)
                    .build();

            return clanStub.applyToClan(request);
        }
        catch (Exception e) {
            throw new RuntimeException("Apply to clan request could not be processed: " + e.getMessage(), e);
        }
    }

    public ProcessAcceptanceResponse processAcceptance(String applicationId, String acceptorId) {
        try {
            ProcessAcceptanceRequest request = ProcessAcceptanceRequest.newBuilder()
                    .setApplicationId(applicationId)
                    .setAcceptorId(acceptorId)
                    .build();

            return clanStub.processAcceptance(request);
        }
        catch (Exception e) {
            throw new RuntimeException("Process acceptance request could not be processed: " + e.getMessage(), e);
        }
    }
}
