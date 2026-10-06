package com.squad.event.repo;

import com.squad.event.model.domain.EventMatchMapVote;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.Optional;
import java.util.UUID;

@Repository
public interface EventMatchMapVoteRepository extends JpaRepository<EventMatchMapVote, UUID> {
    @Query(value = "SELECT * FROM event_match_map_vote WHERE match_id = :match_id AND user_id = :user_id", nativeQuery = true)
    Optional<EventMatchMapVote> findByMatchIdAndUserId(@Param("match_id") UUID matchId, @Param("user_id") UUID userId);

    @Modifying
    @Query(value = """
            INSERT INTO event_match_map_vote (id, match_id, user_id, map)
            VALUES (:id, :match_id, :user_id, :map)
            ON CONFLICT (match_id, user_id)
            DO UPDATE SET
                map = EXCLUDED.map,
                updated_at = CURRENT_TIMESTAMP
            """, nativeQuery = true)
    void upsertVote(
            @Param("id") UUID id,
            @Param("match_id") UUID matchId,
            @Param("user_id") UUID userId,
            @Param("map") String map
    );
}
