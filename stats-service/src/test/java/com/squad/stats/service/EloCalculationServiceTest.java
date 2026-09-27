package com.squad.stats.service;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;

import static org.assertj.core.api.Assertions.assertThat;

public class EloCalculationServiceTest {

    private final EloCalculationService eloService = new EloCalculationService();

    @ParameterizedTest
    // hours, expectedElo (0 hours = 900 elo)
    @CsvSource({
            "0, 900",
            "10, 900",
            "150, 1000",
            "750, 1150",
            "800, 1300",
            "1600, 1450"
    })
    void calculateInitialElo_ShouldReturnCorrectBrackets(int hours, int expectedElo) {
        int result = eloService.calculateInitialElo(hours);
        assertThat(result).isEqualTo(expectedElo);
    }

    @Test
    void calculateMatchElo_ShouldNotDropBelowMinElo() {
        int newElo = eloService.calculateMatchElo(110, false, "Rifleman", 0, 15, 0, 0);
        assertThat(newElo).isEqualTo(100);
    }

    @Test
    void calculateMatchElo_ShouldNotIncreaseHigherMaxMatchElo() {
        int currentElo = 1000;
        int newElo = eloService.calculateMatchElo(currentElo, true, "Rifleman", 50, 0, 50, 50);
        assertThat(newElo).isEqualTo(currentElo + 50);
    }

    @Test
    void calculateMatchElo_ShouldNotDropBelowMinMatchElo() {
        int currentElo = 1000;
        int newElo = eloService.calculateMatchElo(currentElo, true, "Rifleman", 0, 50, 0, 0);
        assertThat(newElo).isEqualTo(currentElo - 50);
    }

    @ParameterizedTest
    @CsvSource({
            "1000, true, Rifleman, ",
            "",

    })
    void calculateMatchElo_ShouldReturnCorrectBracketsInfantry() {

    }
}
