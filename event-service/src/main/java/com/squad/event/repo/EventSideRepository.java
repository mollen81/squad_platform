package com.squad.event.repo;

import com.squad.event.model.domain.EventSide;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface EventSideRepository extends JpaRepository<EventSide, UUID> {
    @Query(value = "SELECT * FROM event_side WHERE event_id = :event_id", nativeQuery = true)
    List<EventSide> findAllByEventId(@Param("event_id") UUID eventId);

    @Query(value = "SELECT * FROM event_side WHERE event_id = :event_id AND leader_user_id = :leader_user_id", nativeQuery = true)
    Optional<EventSide> findAllByEventIdAndLeaderUserId(@Param("event_id") UUID eventId, @Param("leader_id") UUID leaderId);
}