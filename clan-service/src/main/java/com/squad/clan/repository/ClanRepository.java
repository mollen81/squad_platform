package com.squad.clan.repository;

import com.squad.clan.entity.Clan;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.UUID;

@Repository
public interface ClanRepository extends JpaRepository<Clan, UUID> {
    boolean existsByName(String name);
}
