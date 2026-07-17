package com.squad.stats.service;

import com.squad.stats.dto.UserStats;
import com.squad.grpc.stats.GetPlayerStatsByInternalIdRequest;
import com.squad.grpc.stats.GetPlayerStatsResponse;
import com.squad.grpc.stats.StatsServiceGrpc;
import com.squad.stats.repository.UserStatsRepository;
import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.server.service.GrpcService;

import java.util.Optional;
import java.util.UUID;

@GrpcService
@Slf4j
@RequiredArgsConstructor
public class StatsGrpcService extends StatsServiceGrpc.StatsServiceImplBase {
    private final UserStatsRepository userStatsRepository;

    @Override
    public void getPlayerStats(GetPlayerStatsByInternalIdRequest request, StreamObserver<GetPlayerStatsResponse> responseObserver) {
        log.info("New gRPC statistics query for userId: {}", request.getUserId());

        try {
            UUID userId = UUID.fromString(request.getUserId());

            Optional<UserStats> statsOpt = userStatsRepository.findById(userId);

            if(statsOpt.isEmpty()) {
                log.warn("Statistics for userId: {} is not found yet (in process...)", userId);
                responseObserver.onError(Status.NOT_FOUND
                                .withDescription("Statistics not yet processed or user not found")
                                .asRuntimeException());
                return;
            }

            UserStats userStats = statsOpt.get();

            GetPlayerStatsResponse response = GetPlayerStatsResponse.newBuilder()
                    .setEloRating(userStats.getEloRating())
                    .setFavouriteRole(userStats.getFavouriteRole())
                    .setKills(userStats.getKills())
                    .setVehiclesDestroyed(userStats.getVehiclesDestroyed())
                    .setDeaths(userStats.getKills())
                    .setRevives(userStats.getRevives())
                    .build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }
        catch (IllegalArgumentException e) {
            log.error("Incorrect UUID format: {}", request.getUserId());
            responseObserver.onError(Status.INVALID_ARGUMENT
                    .withDescription(e.getMessage())
                    .withCause(e.getCause())
                    .asRuntimeException());
        }
        catch (Exception e) {
            log.error("Internal server error while stats fetching", e);
            responseObserver.onError(Status.INTERNAL
                    .withDescription(e.getMessage())
                    .withCause(e.getCause())
                    .asRuntimeException());
        }
    }
}
