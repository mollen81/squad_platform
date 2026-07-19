package com.squad.clan.service;

import com.squad.clan.grpc.ClanServiceGrpc;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.server.service.GrpcService;

@GrpcService
@Slf4j
@RequiredArgsConstructor
public class ClanGrpcService extends ClanServiceGrpc.ClanServiceImplBase {

    private final
}
