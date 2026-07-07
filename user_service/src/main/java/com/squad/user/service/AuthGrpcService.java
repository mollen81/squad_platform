package com.squad.user.service;

import com.squad.grpc.user.AuthServiceGrpc;
import com.squad.grpc.user.ResolveSteamAuthRequest;
import com.squad.grpc.user.ResolveSteamAuthResponse;
import com.squad.user.domain.UserEntity;
import com.squad.user.event.UserRegisteredEvent;
import com.squad.user.kafka.UserEventProducer;
import com.squad.user.repository.UserRepository;
import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.server.service.GrpcService;

import java.time.LocalDateTime;
import java.time.OffsetDateTime;
import java.util.UUID;

@GrpcService
@Slf4j
@RequiredArgsConstructor
public class AuthGrpcService extends AuthServiceGrpc.AuthServiceImplBase {
    private final SteamOpenIdValidator steamOpenIdvalidator;
    private final UserRepository userRepository;
    private final UserEventProducer userEventProducer;

    @Override
    public void resolveSteamAuth(ResolveSteamAuthRequest request, StreamObserver<ResolveSteamAuthResponse> responseObserver) {
        try {
            String steamId = steamOpenIdvalidator.validateAndExtractSteamId(request.getOpenidParamsMap());

            UserEntity user = userRepository.findBySteamId(steamId).orElse(null);
            boolean isNewUser = false;

            if (user == null) {
                user = UserEntity.builder().steamId(steamId).lastLoginAt(OffsetDateTime.now()).build();
                userRepository.save(user);
                isNewUser = true;

                UserRegisteredEvent event = UserRegisteredEvent.builder()
                        .userId(user.getId()).steamId(steamId).timestamp(System.currentTimeMillis()).build();
                userEventProducer.sendUserRegisteredEvent(event);
            } else {
                user.setLastLoginAt(OffsetDateTime.now());
                userRepository.save(user);
            }

            ResolveSteamAuthResponse response = ResolveSteamAuthResponse.newBuilder()
                    .setUserId(user.getId().toString())
                    .setSteamId(user.getSteamId())
                    .setToken("token_" + UUID.randomUUID())
                    .setIsNewUser(isNewUser)
                    .build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(Status.UNAUTHENTICATED
                    .withDescription("Auth error: " + e.getMessage())
                    .asRuntimeException());
        }
    }
}
