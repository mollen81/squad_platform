package com.squad.gateway.service;

import com.squad.clan.grpc.ClanServiceGrpc;
import com.squad.clan.grpc.CreateClanRequest;
import com.squad.gateway.record.ClanResponse;
import net.devh.boot.grpc.client.inject.GrpcClient;
import org.springframework.stereotype.Service;

@Service
public class ClanGrpcClientService {

    @GrpcClient("clan-service")
    private ClanServiceGrpc.ClanServiceBlockingStub clanStub;

    public ClanResponse createClan(String name, String tag, String leaderSteamId) {
        CreateClanRequest request = CreateClanRequest.newBuilder()
                .setName(name)
                .setTag(tag)
                .setLeaderId(leaderSteamId)
                .build();

        var response = clanStub.createClan(request);

        return new ClanResponse(
                response.getClanId(),
                response.get
        );
    }
}
