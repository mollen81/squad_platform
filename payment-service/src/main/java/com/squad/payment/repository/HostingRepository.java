package com.squad.payment.repository;

import com.squad.payment.model.domain.VdsHosting;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface HostingRepository extends JpaRepository<VdsHosting, String> {
    Optional<VdsHosting> findByHostingId(String hostingId);
    Optional<VdsHosting> findByHostingName(String hostingName);
}
