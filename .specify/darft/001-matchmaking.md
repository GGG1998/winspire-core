/speckit.specify Budujemy matchmakera do naszego serwisu. Nie mamy jeszcze nic dlatego musimy rozplanować działania takie jak:

# Match Making Database

Design tabel, wstępnie wymyśliłem, chodź nie mam pojęcia czy dobrze, dlatego potrzebuję w tej sekcji roli Senior Database Engineera. Turnieje będą miały tryb mode, ale domyślnie jest single elimination 1 vs 1. Chcemy brać przykład z najlepszych z e-sportu. Na przykład możesz się wzorować rozwiązaniami z slither.io

- tabela tournament_brackets - logika drabinki
- tabela tournament_rounds - logiczny układ rund
- tabela tournament_matches - logika meczy, w tym stan meczu

# Eventy
```json
{
  "events": [
    {
      "Notes": "Faza tworzenia turnieju",
      "EventName": "TournamentCreated",
      "Actor": "Host",
      "Command": "CreateTournament",
      "Domain": "Competition",
      "BoundedContext": "TournamentManagement",
      "AggregationName": "Tournament",
      "Invariants": [
        "Host musi mieć aktywne konto",
        "Nazwa turnieju nie może być pusta",
        "minimum_team_count > 1",
        "team_size >= 2"
      ]
    },
    {
      "Notes": "Publikacja turnieju - przejście z draft do scheduled",
      "EventName": "TournamentPublished",
      "Actor": "Host",
      "Command": "PublishTournament",
      "Domain": "Competition",
      "BoundedContext": "TournamentManagement",
      "AggregationName": "Tournament",
      "Invariants": [
        "Turniej musi być w stanie 'draft'",
        "scheduled_start_time musi być ustawiony",
        "scheduled_start_time musi być w przyszłości"
      ]
    },
    {
      "Notes": "Aktualizacja ustawień turnieju",
      "EventName": "TournamentSettingsUpdated",
      "Actor": "Host",
      "Command": "UpdateTournamentSettings",
      "Domain": "Competition",
      "BoundedContext": "TournamentManagement",
      "AggregationName": "Tournament",
      "Invariants": [
        "Turniej musi być w stanie 'draft' lub 'scheduled'",
        "Nie można zmieniać ustawień po starcie"
      ]
    },
    {
      "Notes": "Otwarcie rejestracji na turniej",
      "EventName": "RegistrationOpened",
      "Actor": "Host|System",
      "Command": "OpenRegistration",
      "Domain": "Competition",
      "BoundedContext": "TournamentManagement",
      "AggregationName": "Tournament",
      "Invariants": [
        "Turniej musi być w stanie 'scheduled'"
      ]
    },
    {
      "Notes": "Zamknięcie rejestracji",
      "EventName": "RegistrationClosed",
      "Actor": "Host|System",
      "Command": "CloseRegistration",
      "Domain": "Competition",
      "BoundedContext": "TournamentManagement",
      "AggregationName": "Tournament",
      "Invariants": [
        "Turniej musi być w stanie 'registration_open'"
      ]
    },
    {
      "Notes": "Automatyczne zamknięcie rejestracji gdy wszystkie miejsca zajęte",
      "EventName": "RegistrationAutoClosedDueToCapacity",
      "Actor": "System",
      "Command": "AutoCloseRegistration",
      "Domain": "Competition",
      "BoundedContext": "TournamentManagement",
      "AggregationName": "Tournament",
      "Invariants": [
        "participant_count >= maximum_team_count",
        "Turniej musi być w stanie 'registration_open'"
      ]
    },
    {
      "Notes": "Gracz zarejestrował się na turniej",
      "EventName": "ParticipantRegistered",
      "Actor": "User",
      "Command": "RegisterForTournament",
      "Domain": "Competition",
      "BoundedContext": "TournamentManagement",
      "AggregationName": "TournamentParticipant",
      "Invariants": [
        "Turniej musi być w stanie 'registration_open' lub 'scheduled'",
        "Gracz nie może być już zarejestrowany",
        "participant_count < maximum_team_count (jeśli ustawione)"
      ]
    },
    {
      "Notes": "Gracz potwierdził udział",
      "EventName": "ParticipantConfirmed",
      "Actor": "User",
      "Command": "ConfirmParticipation",
      "Domain": "Competition",
      "BoundedContext": "TournamentManagement",
      "AggregationName": "TournamentParticipant",
      "Invariants": [
        "Status uczestnika musi być 'pending' lub 'registered'",
        "Turniej musi być w stanie 'registration_open' lub 'scheduled'"
      ]
    },
    {
      "Notes": "Gracz zameldował się (check-in, 15 min przed startem)",
      "EventName": "ParticipantCheckedIn",
      "Actor": "User",
      "Command": "CheckInToTournament",
      "Domain": "Competition",
      "BoundedContext": "TournamentManagement",
      "AggregationName": "TournamentParticipant",
      "Invariants": [
        "Status uczestnika musi być 'confirmed'",
        "Czas do startu <= 15 minut",
        "Turniej jeszcze nie wystartował"
      ]
    },
    {
      "Notes": "Gracz wycofał się z turnieju",
      "EventName": "ParticipantWithdrew",
      "Actor": "User",
      "Command": "WithdrawFromTournament",
      "Domain": "Competition",
      "BoundedContext": "TournamentManagement",
      "AggregationName": "TournamentParticipant",
      "Invariants": [
        "Turniej nie może być w stanie 'started' lub 'completed'",
        "Gracz musi być zarejestrowany"
      ]
    },
    {
      "Notes": "Błąd - nie można dołączyć bo miejsca zajęte (z listy użytkownika)",
      "EventName": "RegistrationRejected",
      "Actor": "System",
      "Command": "RegisterForTournament",
      "Domain": "Competition",
      "BoundedContext": "TournamentManagement",
      "AggregationName": "TournamentParticipant",
      "Invariants": [
        "participant_count >= maximum_team_count"
      ]
    },
    {
      "Notes": "Dołączył do turnieju - pokazuje lobby (z listy użytkownika)",
      "EventName": "ParticipantJoinedLobby",
      "Actor": "User|Host",
      "Command": "JoinTournamentLobby",
      "Domain": "Competition",
      "BoundedContext": "Matchmaking",
      "AggregationName": "TournamentLobby",
      "Invariants": [
        "Uczestnik musi być 'confirmed' lub 'checked_in'",
        "Turniej musi być w oknie dołączania (15 min przed startem)",
        "Turniej nie może być już zakończony"
      ]
    },
    {
      "Notes": "Wczytał grę (z listy użytkownika)",
      "EventName": "GameLoaded",
      "Actor": "User",
      "Command": "LoadGame",
      "Domain": "Competition",
      "BoundedContext": "Matchmaking",
      "AggregationName": "TournamentLobby",
      "Invariants": [
        "Uczestnik musi być w lobby",
        "Gra musi być przypisana do turnieju"
      ]
    },
    {
      "Notes": "Host wystartował turniej (z listy użytkownika)",
      "EventName": "TournamentStarted",
      "Actor": "Host",
      "Command": "StartTournament",
      "Domain": "Competition",
      "BoundedContext": "TournamentManagement",
      "AggregationName": "Tournament",
      "Invariants": [
        "Turniej musi być w stanie 'registration_open' lub 'registration_closed'",
        "participant_count >= minimum_team_count",
        "Wszyscy uczestnicy muszą być 'confirmed' lub 'checked_in' (jeśli auto_force_ready=false)"
      ]
    },
    {
      "Notes": "Stworzono drabinkę (z listy użytkownika)",
      "EventName": "BracketGenerated",
      "Actor": "System",
      "Command": "GenerateBracket",
      "Domain": "Competition",
      "BoundedContext": "Matchmaking",
      "AggregationName": "Bracket",
      "Invariants": [
        "Turniej musi być w stanie 'started'",
        "participant_count >= minimum_team_count",
        "Drabinka nie może być już wygenerowana"
      ]
    },
    {
      "Notes": "Stworzono rundę (z listy użytkownika)",
      "EventName": "RoundCreated",
      "Actor": "System",
      "Command": "CreateRound",
      "Domain": "Competition",
      "BoundedContext": "Matchmaking",
      "AggregationName": "Bracket",
      "Invariants": [
        "Drabinka musi istnieć",
        "Poprzednia runda musi być zakończona (jeśli nie jest to pierwsza runda)"
      ]
    },
    {
      "Notes": "Stworzono mecz (z listy użytkownika)",
      "EventName": "MatchCreated",
      "Actor": "System",
      "Command": "CreateMatch",
      "Domain": "Competition",
      "BoundedContext": "Matchmaking",
      "AggregationName": "Match",
      "Invariants": [
        "Runda musi istnieć",
        "Obaj uczestnicy muszą być aktywni w turnieju"
      ]
    },
    {
      "Notes": "Rozpoczęto mecz",
      "EventName": "MatchStarted",
      "Actor": "System|Host",
      "Command": "StartMatch",
      "Domain": "Competition",
      "BoundedContext": "Matchmaking",
      "AggregationName": "Match",
      "Invariants": [
        "Mecz musi być w stanie 'pending' lub 'ready'",
        "Obaj gracze muszą być gotowi (lub auto_force_ready=true)"
      ]
    },
    {
      "Notes": "Zgłoszono wynik meczu",
      "EventName": "ScoreSubmitted",
      "Actor": "User|System|Host",
      "Command": "SubmitScore",
      "Domain": "Competition",
      "BoundedContext": "Matchmaking",
      "AggregationName": "Match",
      "Invariants": [
        "Mecz musi być w stanie 'started'",
        "Wynik musi być prawidłowy (np. score >= 0)"
      ]
    },
    {
      "Notes": "Zakończono mecz z wynikiem",
      "EventName": "MatchCompleted",
      "Actor": "System",
      "Command": "CompleteMatch",
      "Domain": "Competition",
      "BoundedContext": "Matchmaking",
      "AggregationName": "Match",
      "Invariants": [
        "Mecz musi być w stanie 'started'",
        "Wynik musi być zgłoszony i zaakceptowany"
      ]
    },
    {
      "Notes": "Przegrał mecz (z listy użytkownika)",
      "EventName": "ParticipantEliminated",
      "Actor": "System",
      "Command": "EliminateParticipant",
      "Domain": "Competition",
      "BoundedContext": "Matchmaking",
      "AggregationName": "TournamentParticipant",
      "Invariants": [
        "Uczestnik musi być aktywny w turnieju",
        "Mecz musi być zakończony",
        "Uczestnik przegrał mecz"
      ]
    },
    {
      "Notes": "Przeszedł do kolejnej rundy (z listy użytkownika)",
      "EventName": "ParticipantAdvanced",
      "Actor": "System",
      "Command": "AdvanceParticipant",
      "Domain": "Competition",
      "BoundedContext": "Matchmaking",
      "AggregationName": "TournamentParticipant",
      "Invariants": [
        "Uczestnik musi być aktywny w turnieju",
        "Mecz musi być zakończony",
        "Uczestnik wygrał mecz"
      ]
    },
    {
      "Notes": "Zakwestionowano wynik meczu",
      "EventName": "MatchDisputed",
      "Actor": "User",
      "Command": "DisputeMatch",
      "Domain": "Competition",
      "BoundedContext": "Matchmaking",
      "AggregationName": "Match",
      "Invariants": [
        "Mecz musi być zakończony",
        "Czas na dispute nie minął",
        "Gracz musi być uczestnikiem meczu"
      ]
    },
    {
      "Notes": "Rozstrzygnięto spór",
      "EventName": "DisputeResolved",
      "Actor": "Host|Admin",
      "Command": "ResolveDispute",
      "Domain": "Competition",
      "BoundedContext": "Matchmaking",
      "AggregationName": "Match",
      "Invariants": [
        "Dispute musi istnieć",
        "Decyzja musi być podjęta przez uprawnioną osobę"
      ]
    },
    {
      "Notes": "Przyznano walkower",
      "EventName": "WalkoverGranted",
      "Actor": "System|Host",
      "Command": "GrantWalkover",
      "Domain": "Competition",
      "BoundedContext": "Matchmaking",
      "AggregationName": "Match",
      "Invariants": [
        "Mecz musi być w stanie 'pending' lub 'started'",
        "Jeden z graczy nie stawił się lub zrezygnował"
      ]
    },
    {
      "Notes": "Wygrał turniej (z listy użytkownika)",
      "EventName": "TournamentWinnerDeclared",
      "Actor": "System",
      "Command": "DeclareTournamentWinner",
      "Domain": "Competition",
      "BoundedContext": "TournamentManagement",
      "AggregationName": "Tournament",
      "Invariants": [
        "Turniej musi być w stanie 'started'",
        "Finałowy mecz musi być zakończony",
        "Zwycięzca musi być uczestnikiem turnieju"
      ]
    },
    {
      "Notes": "Zakończył turniej (z listy użytkownika)",
      "EventName": "TournamentCompleted",
      "Actor": "System|Host",
      "Command": "CompleteTournament",
      "Domain": "Competition",
      "BoundedContext": "TournamentManagement",
      "AggregationName": "Tournament",
      "Invariants": [
        "Turniej musi być w stanie 'started'",
        "Wszystkie mecze muszą być zakończone",
        "Zwycięzca musi być ogłoszony"
      ]
    },
    {
      "Notes": "Anulował turniej (z listy użytkownika)",
      "EventName": "TournamentCancelled",
      "Actor": "Host",
      "Command": "CancelTournament",
      "Domain": "Competition",
      "BoundedContext": "TournamentManagement",
      "AggregationName": "Tournament",
      "Invariants": [
        "Turniej nie może być w stanie 'completed'",
        "Host musi być właścicielem turnieju"
        "Jeśli turniej trwa, i są aktywne mecze, nie można anulować turnieju", 
      ]
    },
    {
      "Notes": "Przyznano nagrody",
      "EventName": "PrizesDistributed",
      "Actor": "System|Host",
      "Command": "DistributePrizes",
      "Domain": "Competition",
      "BoundedContext": "Rewards",
      "AggregationName": "Tournament",
      "Invariants": [
        "Turniej musi być w stanie 'completed'",
        "Nagrody muszą być skonfigurowane",
        "Zwycięzcy muszą być ogłoszeni"
      ]
    },
    {
      "Notes": "Utworzono drużynę (jeśli team_size > 1)",
      "EventName": "TeamFormed",
      "Actor": "User|System",
      "Command": "FormTeam",
      "Domain": "Competition",
      "BoundedContext": "Matchmaking",
      "AggregationName": "Team",
      "Invariants": [
        "Liczba członków == team_size",
        "Wszyscy członkowie muszą być zarejestrowani na turniej"
      ]
    },
    {
      "Notes": "Przydzielono sędziego do meczu",
      "EventName": "RefereeAssigned",
      "Actor": "Host|System",
      "Command": "AssignReferee",
      "Domain": "Competition",
      "BoundedContext": "Matchmaking",
      "AggregationName": "Match",
      "Invariants": [
        "Mecz musi istnieć",
        "Sędzia musi mieć uprawnienia"
      ]
    },
    {
    "Notes": "Stracił połączenie w trakcie meczu",
    "EventName": "PlayerConnectionLost",
    "Actor": "System",
    "Command": "MarkPlayerDisconnected",
    "Domain": "Competition",
    "BoundedContext": "Matchmaking",
    "AggregationName": "Match",
    "Invariants": [
      "Mecz musi istnieć",
      "Gracz musi być uczestnikiem meczu",
      "Mecz jest w stanie 'started'"
    ]
  },
  {
    "Notes": "Odzyskał połączenie w trakcie meczu",
    "EventName": "PlayerConnectionRestored",
    "Actor": "System",
    "Command": "MarkPlayerReconnected",
    "Domain": "Competition",
    "BoundedContext": "Matchmaking",
    "AggregationName": "Match",
    "Invariants": [
      "Poprzednio musi wystąpić PlayerConnectionLost",
      "Mecz nie może być zakończony",
      "Gracz musi być nadal przypisany do meczu"
    ]
  },
  {
    "Notes": "Wygrał rundę, ale nie ma jeszcze przeciwnika w kolejnej (czeka w kolejce/bracket). Nie może też utknąć",
    "EventName": "ParticipantWaitingForOpponent",
    "Actor": "System",
    "Command": "MarkParticipantWaiting",
    "Domain": "Competition",
    "BoundedContext": "Matchmaking",
    "AggregationName": "Bracket",
    "Invariants": [
      "Uczestnik musi wygrać aktualny mecz",
      "Runda kolejna istnieje lub jest planowana",
      "Brak przypisanego przeciwnika do kolejnego meczu"
    ]
  },
  {
    "Notes": "Streamuje punkty (np. do Redisa) w trakcie meczu/turnieju",
    "EventName": "ScoreStreamedToCache",
    "Actor": "System",
    "Command": "StreamScoreUpdate",
    "Domain": "Competition",
    "BoundedContext": "Telemetry",
    "AggregationName": "Match",
    "Invariants": [
      "Mecz musi istnieć",
      "Wynik musi być w poprawnym formacie",
      "Kanał/cache (np. Redis) musi być dostępny"
    ]
  }
  ]
}
```

# Match Making services backend

## Przepływ zdarzeń
- Czekanie na start turnieju
    - Aplikacja realtime nasłuchuje na event:
        - TournamentStarted (BC: TournamentManagement, Agg: Tournament)
    - Źródło: scheduler lub webhook z serwisu competition (dwa mechanizmy jako fallback).
    -  Redis Pub/Sub konfigurujemy publikację eventu TournamentStarted do kanału, który subskrybuje Matchmaking Realtime.
- Start turnieju → generowanie drabinki i meczów
    - Po TournamentStarted wykonywane są komendy w serwisie competition:
        - GenerateBracket → event BracketGenerated
        - CreateRound → event RoundCreated
        - CreateMatch → event MatchCreated (z przypisanymi losowo graczami)
    - Te eventy zapisujemy w bazie (źródło prawdy) i publikujemy do Matchmaking Realtime.
- Informowanie graczy o meczach i lobby
    - Na podstawie MatchCreated aplikacja realtime:
        - wysyła do graczy (po ich ID) komunikat: „z kim grasz”,
        - emituje event ParticipantJoinedLobby (gracz widzi lobby nadchodzącego meczu).
    -REST API udostępnia komendę JoinTournamentLobby, która prowadzi do eventu ParticipantJoinedLobby.
- Gotowość gracza i wczytanie gry
    - Gdy gracz kliknie „Gotowy”:
        - REST: ConfirmReadiness / CheckInToMatch → event ParticipantCheckedIn (Agg: TournamentParticipant lub Match),
        - Po spełnieniu warunków (obaj gotowi / auto_force_ready) emitujemy:
        - MatchStarted – mecz może się rozpocząć,
        - GameLoaded – jeśli gry wymagają jawnego „wczytania” (event z Twojej listy).
- Jeśli gracz nie kliknie „Gotowy”:
    - w lobby pojawia się licznik 2 min,
    - po przekroczeniu czasu można:
        - emitować PlayerConnectionLost / WalkoverGranted albo
        - ponownie wystawić gracza w stanie ParticipantWaitingForOpponent.

# Match Making frontend

W tej sekcji jesteś UX/UI Developerem oraz znasz się na frontend.


### Wczytywanie Gry

#### Architektura: Static Serve + Script Injection

Gry są serwowane jako statyczne pliki, Game Developer wywołuje polecenia REST i WebSocket

```
[S3/CDN: rozpakowane gry]
         │
         ▼
[Game Proxy Service (Go)]
   │
   ├── GET /g/{tournament_id}/index.html
   │     └── Wstrzykuje <script src="/winspire-sdk.js"> przed </head>
   │     └── Wstrzykuje <script>window.__WINSPIRE_CONTEXT__ = {...}</script>
   │
   └── GET /g/{tournament_id}/**  
         └── Proxy do S3 (CSS, JS, assets, WASM)
```