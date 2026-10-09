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
public interface EventMatchLayerVoteRepository extends JpaRepository<EventMatchMapVote, UUID> {
    @Query(value = "SELECT * FROM event_match_layer_vote WHERE match_id = :match_id AND user_id = :user_id", nativeQuery = true)
    Optional<EventMatchMapVote> findByMatchIdAndUserId(@Param("match_id") UUID matchId, @Param("user_id") UUID userId);

    @Modifying
    @Query(value = """
            INSERT INTO event_match_layer_vote (id, match_id, user_id, layer_name)
            VALUES (:id, :match_id, :user_id, :layer_name)
            ON CONFLICT (match_id, user_id)
            DO UPDATE SET
                layer_name = EXCLUDED.layer_name,
                updated_at = CURRENT_TIMESTAMP
            """, nativeQuery = true)
    void upsertVote(
            @Param("id") UUID id,
            @Param("match_id") UUID matchId,
            @Param("user_id") UUID userId,
            @Param("layer_name") String layerName
    );

    @Query(value = """
            SELECT layer_name FROM event_match_layer_vote
            WHERE match_id = :match_id
            GROUP BY layer_name ORDER BY COUNT(id) DESC, layer_name ASC
            LIMIT 1
            """, nativeQuery = true)
    Optional<String> findWinnerLayerByMatchId(@Param("match_id") UUID matchId);
}
