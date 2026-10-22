package com.squad.event.client;

import com.squad.event.grpc.ClanServiceGrpc;
import com.squad.event.grpc.GetClanIdFromUserIdRequest;
import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.client.inject.GrpcClient;
import org.springframework.stereotype.Service;

import java.util.UUID;

@Service
@RequiredArgsConstructor
@Slf4j
public class ClanServiceGrpcClient {
    @GrpcClient("clan-service")
    private ClanServiceGrpc.ClanServiceBlockingStub clanServiceBlockingStub;

    // userId -> clanId (if exists)
    public UUID resolveUserClanId(UUID userId) {
        try {
            GetClanIdFromUserIdRequest request = GetClanIdFromUserIdRequest.newBuilder()
                    .setUserId(userId.toString())
                    .build();

            String rawClanId = clanServiceBlockingStub.getClanIdFromUserId(request).getClanId();

            if(rawClanId.isBlank()) {
                return null;
            }

            return UUID.fromString(rawClanId);
        }
        catch (StatusRuntimeException e) {
            if(e.getStatus().getCode() == Status.Code.NOT_FOUND) {
                return null;
            }

            log.error("gRPC clan-service error for user {}: {}", userId, e.getStatus().getDescription());
            throw e;
        }
    }
}
