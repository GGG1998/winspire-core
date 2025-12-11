# Tasks: Szybki powrót do meczu (me.match)

**Input**: plan/ustalenia z tego pliku (002-add-me.md)  
**Prerequisites**: aktualny kod usług matchmaking/competition i frontend host (winspire-app)  
**Tests**: uwzględnione tylko tam, gdzie plan wymaga

**Format**: `- [ ] T001 [P?] [Story?] Opis z dokładną ścieżką`

## Phase 1: Setup (Shared Infrastructure)
**Purpose**: Szybkie sanity check konfiguracji runtimu

- [x] T001 [P] Zweryfikuj konfigurację JWT host w `services/matchmaking/internal/config/config.go` (sekret/issuer/audience) dla nowego endpointu `/v1/matchmaking/me`

---

## Phase 2: Foundational (Blocking Prerequisites)
**Purpose**: Repozytorium i serwisowe wsparcie wyszukiwania meczów po userId

- [x] T002 [P] Dodaj zapytanie repozytorium do pobierania aktywnego/ostatniego meczu po userId w `services/matchmaking/internal/repository/match_repository.go`
- [x] T003 Zaimplementuj metodę serwisową `GetCurrentMatchForUser` w `services/matchmaking/internal/application/match_service.go` bazującą na nowym repo

---

## Phase 3: User Story 1 - Powrót do bieżącego meczu (Priority: P1) 🎯 MVP
**Goal**: Użytkownik hosta widzi w UI przyciski/linki „Wróć do meczu” oparte o `me.match` z matchmakingu  
**Independent Test**: Dla usera z aktywnym meczem `/v1/matchmaking/me` zwraca dane meczu; w winspire-app header i sidebar pokazują link do meczu; dla usera bez meczu link znika.

### Tests for User Story 1 (zgodnie z planem)
- [x] T004 [P] [US1] Dodaj testy handlera `/v1/matchmaking/me` (scenariusze: aktywny mecz, brak meczu) w `services/matchmaking/internal/http/me_handler_test.go`

### Implementation for User Story 1
- [x] T005 [P] [US1] Utwórz handler `/v1/matchmaking/me` zwracający `matchId`, `tournamentId`, `status`, `round`, `table`, `startedAt/updatedAt` w `services/matchmaking/internal/http/me_handler.go`
- [x] T006 [US1] Zarejestruj trasę GET `/v1/matchmaking/me` z autoryzacją JWT w `services/matchmaking/cmd/matchmaking/main.go`
- [x] T007 [US1] Rozszerz kontrakt `TournamentMeInfo` o `match` (matchId, tournamentId, status, round, table, startedAt, updatedAt) w `services/competition/internal/http/handlers/types.go`
- [x] T008 [US1] Pobierz dane z matchmaking `/me` i wypełnij `detail.Me.Match` w `services/competition/internal/http/handlers/tournament_host.go` (wybór aktywnego/ostatniego meczu)
- [x] T009 [P] [US1] Rozszerz typy hosta o `me.match` w `frontends/winspire-app/src/features/host/types.ts`
- [x] T010 [P] [US1] Zmapuj `me.match` z API do modelu UI w `frontends/winspire-app/src/features/host/api/tournamentApi.ts`
- [x] T011 [US1] Oblicz `canJoinMatch` i wyrenderuj przycisk „Wróć do meczu” w `frontends/winspire-app/src/features/host/components/TournamentHeader.tsx` (link do ścieżki meczu)
- [x] T012 [US1] Dodaj warunkowy link „Wróć do meczu” w sidebarze `frontends/winspire-app/src/features/host/layouts/HostLayout.tsx` wykorzystujący `me.match`
- [x] T013 [P] [US1] Dodaj test UI/VT dla widoczności przycisków (aktywny vs brak meczu) w `frontends/winspire-app/src/features/host/__tests__/return-to-match.test.tsx`

---


## Dependencies & Execution Order
- Setup (Phase 1) → Foundational (Phase 2) → User Story 1 (Phase 3) → Polish
- Wewnątrz US1: T005/T006 zależą od T002–T003; T008 zależy od T005–T006; front (T009–T012) zależy od dostępności pola `me.match`; test T004/T013 po implementacji.

## Parallel Opportunities
- [P] repo + handler test (T002, T004, T005, T009, T010, T013, T014) mogą być równolegle po zakończeniu zależności.

## Report
- Output path: `.specify/darft/002-add-me.md`
- Total tasks: 15
- Task count per story: US1 = 10 (T004–T013)
- Parallel opportunities: oznaczone [P]
- Independent test criteria: patrz US1 opis
- Suggested MVP scope: Phase 1–2 + Phase 3 (US1) bez Polish
- Format validation: wszystkie zadania mają `- [ ] TXXX [P?] [Story?] Opis z ścieżką`
Jesteś Agentem Weryfikacji Spójności. Twoim zadaniem jest sprawdzenie, czy każdy z wymienionych kroków z planu pracy został poprawnie odzwierciedlony w dostarczonym fragmencie kodu oraz jasno wskazanie braków.

Cele weryfikacji:
- Zweryfikuj każdy krok z planu względem kontekstu kodu.
- Zidentyfikuj brakujące lub niespójne elementy i opisz, co trzeba dodać/poprawić.
- Uwzględnij kontekst DDD i bounded context (competition vs matchmaking vs game-management).

Dla każdego wymaganego kroku z listy: 
1. Zlokalizuj odpowiednie miejsce w kontekście (jeśli trzeba, zajrzyj do plików).
2. Oceń spójność: Czy krok jest zgodny, niezgodny, czy brak go w kodzie.
3. Jeśli jest niezgodny lub brakuje go, wyjaśnij precyzyjnie, co jest nie tak i co należy zmienić.

- Wskazówka pracujemy z metodologią DDD i mikroserwisów, chcemy aby funkcjnonalność była zgodna
z bounded context
- competition, traktuj jak tournament-management dla CRUD
- matchmaking - system do zarządzania match, rundami, drabinką oraz streamingiem danych między userami
- game-management - serwowanie gry, dodawanie, usuwanie


```
## Cel
Umożliwić graczowi szybki powrót do swojego meczu (przycisk w headerze i link w sidebarze), opierając się o dane „me” z matchmakingu.

## Zakres
- Backend (matchmaking): nowe dane o bieżących meczach użytkownika.
- Frontend (winspire-app host): pobranie i wykorzystanie `me.match` do renderowania linków „Wróć do meczu”.

## Plan techniczny

### 1. Backend: endpoint /me w matchmaking
- Dodać endpoint REST `/me` (opcjonalnie `/tournaments/{id}/me` jeżeli potrzebny kontekst turnieju) zwracający listę lub bieżący mecz użytkownika.
- Pola zwracane (MVP): `matchId`, `tournamentId`, `status`, `round`, `table`, `startedAt/updatedAt` (jeśli dostępne).
- Implementacja: użyć istniejących serwisów wyszukiwania meczów po userId; respektować kontekst autoryzacji/tenanta.
- Dodane testy handlera i serwisu (scenariusze: mecz aktywny, brak meczu).

### 2. Kontrakt danych dla host app
- Rozszerzyć kontrakt `me` w odpowiedzi turnieju: dodać sekcję `match` z polami z pkt 1.
- Jeśli backend zwraca listę meczów, po stronie API wybieramy aktualny/ostatni aktywny do `me.match`.

### 3. Frontend: typy i API
- Zaktualizować `frontends/winspire-app/src/features/host/types.ts`: dodać `me.match` (matchId, status, round/table).
- W `frontends/winspire-app/src/features/host/api/tournamentApi.ts` podciągnąć dane `me.match` z nowej odpowiedzi; ewentualnie dołożyć fetch do `/me` jeśli nie jest częścią detail.

### 4. UI: przyciski powrotu
- `TournamentHeader`: wyliczyć `canJoinMatch` na bazie `me.match` i wyrenderować guzik „Wróć do meczu” kierujący do trasy meczu (np. `/host/matches/{matchId}`).
- `HostLayout` sidebar: warunkowo dodać link „Wróć do meczu” prowadzący do tego samego URL.
- Widoczność: tylko gdy `me.match` istnieje; ukryte w innych stanach.

### 5. Weryfikacja
- Manualnie: użytkownik z aktywnym meczem widzi guzik/link i przechodzi do meczu; bez meczu brak linku.
- Automatycznie: testy backendowego endpointu; ewentualnie testy jednostkowe selektorów/front-end utils jeśli istnieją.

```

2. Wynik: 
- Wypisz serwis, co robi, dlaczego może tu pasować i zagrożnia
- dla danego serwisu wypisz liste plików, które pasują lub wymagają modyfikacji

