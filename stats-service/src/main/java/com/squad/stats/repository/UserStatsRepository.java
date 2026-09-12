package com.squad.stats.repository;

import com.squad.stats.dto.UserStats;
import org.checkerframework.checker.nullness.qual.NonNull;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;
import java.util.UUID;

@Repository
public interface UserStatsRepository extends JpaRepository<UserStats, UUID> {
    @Override
    Optional<UserStats> findById(@NonNull UUID uuid);
    Optional<UserStats> findBySteamId(String steamId);
}
