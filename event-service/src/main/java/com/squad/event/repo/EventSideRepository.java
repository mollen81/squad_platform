package com.squad.event.repo;

import com.squad.event.model.domain.EventSide;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.UUID;

@Repository
public interface EventSideRepository extends JpaRepository<EventSide, UUID> {
}