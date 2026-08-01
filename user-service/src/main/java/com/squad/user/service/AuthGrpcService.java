package com.squad.user.service;

import com.squad.grpc.user.AuthServiceGrpc;
import com.squad.grpc.user.ResolveSteamAuthRequest;
import com.squad.grpc.user.ResolveSteamAuthResponse;
import com.squad.user.client.SteamWebApiClient;
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
import java.util.Optional;
import java.util.UUID;

@GrpcService
@Slf4j
@RequiredArgsConstructor
public class AuthGrpcService extends AuthServiceGrpc.AuthServiceImplBase {

    private final SteamOpenIdValidator steamOpenIdValidator;
    private final SteamWebApiClient steamWebApiClient;
    private final JwtService jwtService;
    private final UserRepository userRepository;
    private final UserEventProducer userEventProducer;

    @Override
    public void resolveSteamAuth(ResolveSteamAuthRequest request, StreamObserver<ResolveSteamAuthResponse> responseObserver) {
        try {
            String steamId = steamOpenIdValidator.validateAndExtractSteamId(request.getOpenidParamsMap());

            Optional<UserEntity> userOpt = userRepository.findBySteamId(steamId);
            UserEntity user;
            boolean isNewUser = false;

            // SteamApi
            SteamWebApiClient.SteamProfile profile = steamWebApiClient.fetchUserProfile(steamId);

            if (userOpt.isEmpty()) {
                // Registration
                user = UserEntity.builder()
                        .steamId(steamId)
                        .displayName(profile.displayName())
                        .avatarUrl(profile.avatarUrl())
                        .build();

                user = userRepository.save(user);
                isNewUser = true;
                log.info("New user is registered, UUID: {}, Ник: {}", user.getId(), user.getDisplayName());

                // Kafka -> BattleMetrics
                UserRegisteredEvent event = UserRegisteredEvent.builder()
                        .userId(user.getId())
                        .steamId(steamId)
                        .timestamp(System.currentTimeMillis())
                        .build();
                userEventProducer.sendUserRegisteredEvent(event);
            }
            else {
                // Auth
                user = userOpt.get();
                user.setDisplayName(profile.displayName());
                user.setAvatarUrl(profile.avatarUrl());
                user.setLastLoginAt(OffsetDateTime.from(LocalDateTime.now()));
                user = userRepository.save(user);
                log.info("User {} ({}) is successfully authenticated.", user.getDisplayName(), steamId);
            }

            // JWT generation
            String jwtToken = jwtService.generateToken(user.getId(), steamId);

            // Gateway
            ResolveSteamAuthResponse response = ResolveSteamAuthResponse.newBuilder()
                    .setUserId(user.getId().toString())
                    .setSteamId(steamId)
                    .setToken(jwtToken)
                    .setIsNewUser(isNewUser)
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception e) {
            log.error("Auth error: {}", e.getMessage(), e);
            responseObserver.onError(Status.UNAUTHENTICATED
                    .withDescription("Auth error: " + e.getMessage())
                    .asRuntimeException());
        }
    }
}
