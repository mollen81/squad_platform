package com.squad.event.repo;

import com.squad.event.model.domain.TimelineEvent;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.UUID;

@Repository
public interface EventMatchTimelineRepository extends JpaRepository<TimelineEvent, UUID> {
    @Query(value = "SELECT * FROM timeline_event WHERE match_id = :match_id", nativeQuery = true)
    List<TimelineEvent> findAllByMatchId(@Param("match_id") UUID matchId);

    @Query(value = "SELECT * FROM timeline_event WHERE match_id = :match_id ORDER BY game_time_seconds ASC", nativeQuery = true)
    List<TimelineEvent> findAllByMatchIdOrderByTime(@Param("match_id") UUID matchID);
}
