package com.squad.clan.facade;

import com.squad.clan.client.StatsGrpcClient;
import com.squad.clan.dto.ClanRequests;
import com.squad.clan.entity.Clan;
import com.squad.clan.entity.ClanApplication;
import com.squad.clan.service.ClanService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

@Component
@RequiredArgsConstructor
@Slf4j
public class ClanFacade {
    private final ClanService clanService;
    private final StatsGrpcClient statsGrpcClient;

    public Clan createClan(ClanRequests.CreateClanDto dto) {
        int leaderElo = statsGrpcClient.getPlayerElo(dto.getLeaderId());

        return clanService.createClan(dto, leaderElo);
    }

    public ClanApplication applyClan(ClanRequests.ApplyToClanDto dto) {
        int userElo = statsGrpcClient.getPlayerElo(dto.getUserId());

        return clanService.applyToClan(dto, userElo);
    }

    public Clan getClanWithMembers(ClanRequests.GetClanWithAllMembersDto dto) {
        return clanService.getClanWithMembers(dto);
    }

}
