package com.squad.battlemetrics;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication(scanBasePackages = "com.squad")
public class BattleMetricsServiceApplication {
    public static void main(String[] args){
        SpringApplication.run(BattleMetricsServiceApplication.class, args);
    }
}
