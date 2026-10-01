package com.squad.payment.model.domain;

import com.squad.payment.model.enums.PaymentStatus;
import com.squad.payment.model.enums.converters.PaymentStatusConverter;
import jakarta.persistence.*;
import lombok.*;

import java.math.BigDecimal;
import java.util.List;

@Entity
@AllArgsConstructor
@NoArgsConstructor
@Builder
public class Payment extends BaseEntity {
    private BigDecimal amount;

    @Convert(converter = PaymentStatusConverter.class)
    private PaymentStatus paymentStatus;

    private String errorMessage;

    @OneToOne
    @JoinColumn(name = "event_id")
    private String eventId;

    @OneToOne
    @JoinColumn(name = "vps_hosting_id")
    private VpsHosting vpsHosting;

    @OneToMany(mappedBy = "payment", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<Refund> refunds;
}
