package com.squad.event.repo;

import com.squad.event.model.domain.Event;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface EventRepository extends JpaRepository<Event, UUID> {
    // Pessimistic lock (SELECT ... FOR UPDATE)
    @Query(value = "SELECT * FROM event WHERE id = :id FOR UPDATE", nativeQuery = true)
    Optional<Event> findByIdForUpdate(@Param("id") UUID id);

    @Query(value = "SELECT * FROM event WHERE status NOT IN ('FINISHED', 'CANCELED')", nativeQuery = true)
    List<Event> findAllUnfinished();

    @Query(value = "SELECT * FROM event WHERE status = 'REGISTRATION'", nativeQuery = true)
    List<Event> findAllInRegistration();

    @Query(value = "SELECT * FROM event WHERE status = 'FINISHED' ORDER BY time_start", nativeQuery = true)
    List<Event> findAllFinished();

    @Query(value = "SELECT * FROM event WHERE status = 'LIVE'", nativeQuery = true)
    List<Event> findAllLive();

    @Query(value = "SELECT * FROM event WHERE id = (SELECT event_id FROM event_match WHERE id = :match_id)", nativeQuery = true)
    Optional<Event> findByMatchId(@Param("match_id") UUID matchId);
}