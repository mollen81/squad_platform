package com.squad.gateway.controller;

import com.squad.gateway.record.ClanService.ApplyToClanResponse;
import com.squad.gateway.record.ClanService.Clan;
import com.squad.gateway.record.ClanService.CreateClanResponse;
import com.squad.gateway.record.ClanService.ProcessAcceptanceResponse;
import com.squad.gateway.service.ClanGrpcClientService;
import lombok.RequiredArgsConstructor;
import org.springframework.graphql.data.method.annotation.Argument;
import org.springframework.graphql.data.method.annotation.MutationMapping;
import org.springframework.graphql.data.method.annotation.QueryMapping;
import org.springframework.stereotype.Controller;

@Controller
@RequiredArgsConstructor
public class ClanController {
    private final ClanGrpcClientService clanGrpcClientService;

    @QueryMapping
    public Clan getClanWithMembers(@Argument String clanId) {
        return clanGrpcClientService.getClanWithMembers(clanId);
    }

    @MutationMapping
    public CreateClanResponse createClan(
            @Argument String leaderId,
            @Argument String name,
            @Argument String tag,
            @Argument String description,
            @Argument String requirements,
            @Argument String avatarUrl) {
        return clanGrpcClientService.createClan(name, tag, leaderId, description, requirements, avatarUrl);
    }

    @MutationMapping
    public ApplyToClanResponse applyToClanResponse(
            @Argument String userId,
            @Argument String clanId,
            @Argument String socialLink,
            @Argument String experienceText) {
        return clanGrpcClientService.applyToClan(userId, clanId, socialLink, experienceText);
    }

    @MutationMapping
    public ProcessAcceptanceResponse processAcceptance(
            @Argument String applicationId,
            @Argument String acceptorId) {
        return clanGrpcClientService.processAcceptance(applicationId, acceptorId);
    }
}
