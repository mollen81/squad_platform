package com.squad.event.repo;

import com.squad.event.model.domain.EventServer;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.Optional;
import java.util.UUID;

@Repository
public interface EventServerRepository extends JpaRepository<EventServer, UUID> {

    @Query(value = "SELECT * FROM event_server WHERE event_id = :event_id", nativeQuery = true)
    Optional<EventServer> findByEventId(@Param("event_id") UUID eventId);
}