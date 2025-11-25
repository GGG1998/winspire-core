# Feature Specification: System turniejowy do gier web

**Feature Branch**: `[###-tournament-system]`  
**Created**: 2025-11-22  
**Status**: Draft  
**Input**: User description: "Tworzę system turniejowy do gier web z pokojem, leaderboardem, kontrolą ról i formularzem konfiguracji."

## User Scenarios & Testing

### User Story 1 - Autoryzowane tworzenie turnieju (Priority: P1)
Organizator z rolą `streamer` lub `admin` tworzy nowy turniej, wypełniając formularz konfiguracji i otrzymuje dedykowany pokój kontrolny.

**Why this priority**: Bez bezpiecznego tworzenia turniejów nie możemy uruchomić żadnej rywalizacji.

**Independent Test**: Wejście na stronę organizatora, utworzenie turnieju jako `streamer`, sprawdzenie że inna rola otrzymuje błąd 403.

**Acceptance Scenarios**:
1. **Given** użytkownik z rolą `streamer`, **When** wypełni formularz i zatwierdzi, **Then** turniej zostaje utworzony i przypisany do pokoju.
2. **Given** użytkownik bez roli `streamer`, **When** próbuje utworzyć turniej, **Then** API zwraca `forbidden`.

---

### User Story 2 - Pokój turniejowy (Priority: P2)
Organizator i gracze widzą w jednym miejscu harmonogram, statusy lineupów oraz akcje (ready, confirm) dla bieżącego turnieju.

**Why this priority**: Pokój jest centralnym widokiem operacyjnym; bez niego gracze nie wiedzą co robić.

**Independent Test**: Utworzenie turnieju, zalogowanie jako uczestnik i potwierdzenie udziału w pokoju bez przechodzenia do gry

**Acceptance Scenarios**:
1. **Given** turniej w stanie `OPEN`, **When** lineup wejdzie do pokoju, **Then** zobaczy aktualne okno rejestracji i przycisk potwierdzenia.
2. **Given** turniej w stanie `RUNNING`, **When** lineup potwierdzi gotowość, **Then** status w pokoju aktualizuje się na „ready”.

---

### User Story 3 - Leaderboard / Ladder (Priority: P3)
Użytkownik przegląda leaderboard (drabinkę) powiązaną z turniejem i analizuje swoje wyniki oraz historię meczów.

**Why this priority**: Ranking dodaje kontekst rywalizacji; możemy go wdrożyć po pokoju i formularzu.

**Independent Test**: Zapytanie o `Ladder` → `placementsPage` → `matchesPage` w oparciu o Challengermode API; walidacja poprawnego renderu.

**Acceptance Scenarios**:
1. **Given** aktywna drabinka, **When** użytkownik otworzy leaderboard, **Then** zobaczy dane z `Ladder` i `LadderPlacement`.
2. **Given** placement użytkownika, **When** rozwinie historię, **Then** zobaczy mecze z `LadderMatch` z flagą `includedInScoring`.

---

### Edge Cases
- Co jeśli użytkownik nie ma wymaganej roli przy próbie tworzenia/edycji turnieju?
- Jak system reaguje, gdy Challengermode zwraca `LadderState = CANCELLED` podczas renderu leaderboardu?
- Co w przypadku, gdy formularz konfiguracji ma niekompletne dane (np. brak harmonogramu)?
- Jak obsłużyć sytuację, gdy mutacje `joinLadder`/`leaveLadder` są niedostępne (deprecated) i musimy użyć alternatywnych endpointów leaderboardowych?

## Requirements

### Functional Requirements
- **FR-001**: System MUST umożliwiać tworzenie turnieju wyłącznie użytkownikom z rolą `streamer` lub wyższą.
- **FR-002**: System MUST zwracać odpowiedź `403 forbidden`, gdy tworzenie turnieju wykonuje użytkownik bez wymaganej roli.
- **FR-003**: Formularz konfiguracji turnieju MUST zawiera pola: nazwa, opis, harmonogram, ograniczenia lineupów, powiązanie z grą.
- **FR-004**: System MUST tworzyć/powiązywać pokój turniejowy, który agreguje status lineupów i akcje (ready/confirm).
- **FR-005**: Widok leaderboardu MUST prezentować dane z Challengermode `Ladder` (`placementsPage`, `matchesPage`) z paginacją.
- **FR-006**: System MUST obsługiwać fallback na leaderboardowe mutacje (`joinLeaderboard`, `leaveLeaderboard`) w przypadku deprecacji `joinLadder`/`leaveLadder` [NEEDS CLARIFICATION: wymagana ścieżka migracji].
- **FR-007**: System MUST logować próby nieautoryzowanego dostępu do funkcji organizatorskich.

### Key Entities
- **TournamentConfig**: dane formularza (gra, harmonogram, limity lineupów, zasady potwierdzania).
- **TournamentRoom**: stan lobby (statusy lineupów, akcje do wykonania, okna czasowe).
- **ChallengermodeLadder**: mapping do typów `Ladder`, `LadderPlacement`, `LadderMatch`, `LadderState`, `PlacementsPageConnection`, `MatchesPageConnection` z `docs-site/docs/core/demo.graphql` oraz [definicji Challengermode](https://www.challengermode.com/developers/docs/challengermode-api/reference/definitions#ladder).

## Success Criteria

### Measurable Outcomes
- **SC-003**: Leaderboard renderuje do 200 pozycji w <2 s przy wykorzystaniu `placementsPage`.
- **SC-004**: <1% błędów API podczas pobierania danych `Ladder`/`Leaderboard` w ciągu tygodnia od wdrożenia.

