package com.squad.event.repo;

import com.squad.event.model.domain.EventParticipant;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface EventParticipantRepository extends JpaRepository<EventParticipant, UUID> {

    @Query(value = "SELECT EXISTS(SELECT 1 FROM event_participant WHERE event_id = :event_id AND user_id = :user_id)", nativeQuery = true)
    boolean existsByEventIdAndUserId(@Param("event_id") UUID eventId, @Param("user_id") UUID userId);

    @Query(value = "SELECT * FROM event_participant WHERE event_id = :event_id AND user_id = :user_id", nativeQuery = true)
    Optional<EventParticipant> findByEventIdAndUserId(@Param("event_id") UUID eventId, @Param("user_id") UUID userId);

    @Query(value = "SELECT COUNT(*) FROM event_participant WHERE side_id = :side_id")
    int countBySideId(@Param("side_id") UUID sideId);

    @Query(value = "SELECT COUNT(*) FROM event_participant WHERE side_id = :side_id AND clan_id = :clan_id")
    int countBySideIdAndClanId(@Param("side_id") UUID sideId, @Param("clan_id") UUID clanId);

    @Query(value = "SELECT * FROM event_participant WHERE side_id = :side_id AND clan_id = :clan_id", nativeQuery = true)
    List<EventParticipant> findAllBySideIdAndClanId(@Param("side_id") UUID sideId, @Param("clan_id") UUID clanId);

    @Modifying
    @Query(value = "DELETE FROM event_participant WHERE event_id = :event_id AND user_id = :user_id", nativeQuery = true)
    void deleteByEventIdAndUserId(@Param("event_id") UUID eventId, @Param("user_id") UUID userId);
}