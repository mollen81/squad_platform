package com.squad.event.repo;

import com.squad.event.model.domain.EventMatch;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface EventMatchRepository extends JpaRepository<EventMatch, UUID> {

    // Pessimistic lock
    @Query(value = "SELECT * FROM event_match WHERE id = :id FOR UPDATE", nativeQuery = true)
    Optional<EventMatch> findByIdForUpdate(@Param("id") UUID id);

    @Query(value = "SELECT * FROM event_match WHERE event_id = :event_id ORDER BY sequence_number", nativeQuery = true)
    List<EventMatch> findAllByEventIdOrderBySequenceNumber(@Param("event_id") UUID eventId);

    @Query(value = "SELECT * FROM event_match WHERE event_id = :event_id AND sequence_number = :sequence_number", nativeQuery = true)
    Optional<EventMatch> findByEventIdAndSequenceNumber(@Param("event_id") UUID eventIUd, @Param("sequence_number") int sequenceNumber);

    @Query(value = "SELECT COUNT(*) FROM event_match WHERE event_id = :event_id AND winner_side_id = :winner_side_id", nativeQuery = true)
    int countByEventIdAndWinnerSideId(@Param("event_id") UUID eventId, @Param("winner_side_id") UUID winnerSideId);
}