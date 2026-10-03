package com.squad.payment.model.enums;

import lombok.Getter;

@Getter
public enum PaymentStatus {
    PROCESSING,
    SUCCESS,
    FAILED,
    CLOSED;

    public static PaymentStatus fromString(String stringStatus) {
        for(PaymentStatus paymentStatus : PaymentStatus.values()) {
            if(paymentStatus.toString().equalsIgnoreCase(stringStatus)) {
                return paymentStatus;
            }
        }

        throw new IllegalArgumentException("Invalid payment status: " + stringStatus);
    }
}
