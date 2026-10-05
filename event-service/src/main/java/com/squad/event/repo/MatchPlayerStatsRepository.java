package com.squad.event.repo;

import com.squad.event.model.domain.MatchPlayerStats;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.UUID;

@Repository
public interface MatchPlayerStatsRepository extends JpaRepository<MatchPlayerStats, UUID> {
}