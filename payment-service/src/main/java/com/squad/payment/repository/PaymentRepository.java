package com.squad.payment.repository;

import com.squad.payment.model.domain.Payment;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface PaymentRepository extends JpaRepository<Payment, String> {
    Optional<Payment> findByPaymentId(String paymentId);
    Optional<Payment> findByEventId(String eventId);
}
