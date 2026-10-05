package com.squad.event.repo;

import com.squad.event.model.domain.EventMatchPlayerStats;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.Optional;
import java.util.UUID;

@Repository
public interface EventMatchPlayerStatsRepository extends JpaRepository<EventMatchPlayerStats, UUID> {

    @Query(value = "SELECT * FROM event_match_player_stats WHERE event_id = :event_id", nativeQuery = true)
    Optional<EventMatchPlayerStats> findByEventId(@Param("event_id") UUID eventId);


    // TODO: int countKillsBySideId(@Param("") UUID side_id);
}