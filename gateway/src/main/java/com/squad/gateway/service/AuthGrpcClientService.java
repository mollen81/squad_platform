package com.squad.gateway.service;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.squad.grpc.user.AuthServiceGrpc;
import com.squad.grpc.user.ResolveSteamAuthRequest;
import com.squad.grpc.user.ResolveSteamAuthResponse;
import lombok.RequiredArgsConstructor;
import net.devh.boot.grpc.client.inject.GrpcClient;
import org.springframework.stereotype.Service;

import java.util.Map;

@Service
@RequiredArgsConstructor
public class AuthGrpcClientService {

    @GrpcClient("user-service")
    private AuthServiceGrpc.AuthServiceBlockingStub authServiceStub;

    private final ObjectMapper mapper;

    public ResolveSteamAuthResponse loginWithSteam(String openidParamsJson) {
        try {
            Map<String, String> params = mapper.readValue(
                    openidParamsJson,
                    new TypeReference<>() {});
            ResolveSteamAuthRequest request = ResolveSteamAuthRequest.newBuilder()
                    .putAllOpenidParams(params)
                    .build();
            return authServiceStub.resolveSteamAuth(request);
        }
        catch (Exception e) {
            throw new RuntimeException(
                    "Steam authorization request could not be processed: " + e.getMessage());
        }
    }
}
