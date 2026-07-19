package com.squad.clan.client;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.client.inject.GrpcClient;
import org.springframework.stereotype.Service;

@Service
@Slf4j
@RequiredArgsConstructor
public class StatsGrpcClient {

    @GrpcClient("stats-service")
    private StatsServiceGrpc.
}
