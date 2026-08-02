package com.squad.clan.repository;

import com.squad.clan.entity.Clan;
import org.springframework.data.jpa.repository.EntityGraph;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.Optional;
import java.util.UUID;

@Repository
public interface ClanRepository extends JpaRepository<Clan, UUID> {

    boolean existsByName(String name);

    @Query("SELECT c FROM Clan c LEFT JOIN FETCH c.members WHERE c.id = :clan_id")
    Optional<Clan> findByIdWithAllMembers(@Param("clan_id") UUID clanId);

    @EntityGraph(attributePaths = {"members"})
    Optional<Clan> findWithMembersById(UUID id);
}
