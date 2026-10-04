package com.squad.payment.model.enums;

import lombok.Getter;

@Getter
public enum Hosting {
    AEZA,
    SELECTEL;

    public static Hosting fromString(String hostingName) {
        for(Hosting hosting : Hosting.values()) {
            if(hosting.name().equalsIgnoreCase(hostingName)) {
                return hosting;
            }
        }

        throw new IllegalArgumentException("Invalid hosting name: " + hostingName);
    }
}
