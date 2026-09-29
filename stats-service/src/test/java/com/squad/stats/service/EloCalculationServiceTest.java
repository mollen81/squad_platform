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

    // ROLE TESTS

    // INFANTRY: Rifleman, Ambusher, Raider, Automatic Rifleman, Machine Gunner
    @ParameterizedTest
    @CsvSource({
            // positive elo
            "1000, true, Rifleman, 10, 0, 5, 0, 1044",
            "1000, true, Ambusher, 10, 0, 5, 0, 1044",
            "1000, true, Raider, 10, 0, 5, 0, 1044",
            "1000, true, Automatic Rifleman, 10, 0, 5, 0, 1044",
            "1000, true, Machine Gunner, 10, 0, 5, 0, 1044",
            // negative elo
            "1000, false, Rifleman, 2, 10, 2, 0, 961",
            "1000, false, Ambusher, 2, 10, 2, 0, 961",
            "1000, false, Raider, 2, 10, 2, 0, 961",
            "1000, false, Automatic Rifleman, 2, 10, 2, 0, 961",
            "1000, false, Machine Gunner, 2, 10, 2, 0, 961",
    })
    void calculateMatchElo_ShouldReturnCorrectBracketsInfantry(int currentElo, boolean isWin, String role,
                                                               int kills, int deaths, int revives,
                                                               int vehiclesDestroyed, int expectedElo) {
        int result = eloService.calculateMatchElo(currentElo, isWin, role, kills, deaths, revives, vehiclesDestroyed);
        assertThat(result).isEqualTo(expectedElo);
    }

    // MEDIC
    @ParameterizedTest
    @CsvSource({
            // positive test
            "1000, true, Medic, 10, 0, 5, 0, 1036",
            // negative test
            "1000, false, Medic, 0, 6, 0, 0, 955"
    })
    void calculateMatchElo_ShouldReturnCorrectBracketsMedic(int currentElo, boolean isWin, String role,
                                                               int kills, int deaths, int revives,
                                                               int vehiclesDestroyed, int expectedElo) {
        int result = eloService.calculateMatchElo(currentElo, isWin, role, kills, deaths, revives, vehiclesDestroyed);
        assertThat(result).isEqualTo(expectedElo);
    }

    // SNIPERS: Sniper, Marksman
    @ParameterizedTest
    @CsvSource({
            // positive test
            "1000, true, Sniper, 10, 0, 5, 0, 1046",
            // negative test
            "1000, false, Marksman, 0, 6, 0, 0, 950"
    })
    void calculateMatchElo_ShouldReturnCorrectBracketsSniper(int currentElo, boolean isWin, String role,
                                                            int kills, int deaths, int revives,
                                                            int vehiclesDestroyed, int expectedElo) {
        int result = eloService.calculateMatchElo(currentElo, isWin, role, kills, deaths, revives, vehiclesDestroyed);
        assertThat(result).isEqualTo(expectedElo);
    }


    // STORMTROOPS: Grenadier, Scout
    @ParameterizedTest
    @CsvSource({
            // positive test
            "1000, true, Grenadier, 10, 2, 5, 0, 1040",
            // negative test
            "1000, false, Scout, 3, 6, 0, 0, 955"
    })
    void calculateMatchElo_ShouldReturnCorrectBracketsStormTroops(int currentElo, boolean isWin, String role,
                                                             int kills, int deaths, int revives,
                                                             int vehiclesDestroyed, int expectedElo) {
        int result = eloService.calculateMatchElo(currentElo, isWin, role, kills, deaths, revives, vehiclesDestroyed);
        assertThat(result).isEqualTo(expectedElo);
    }

    // TODO: Anti-Tanks, Crewman, Pilots, Engineers, Infiltrator tests
}
