# Proponowana architektura (sprawdzony pattern)

❌ Dlaczego NIE Supabase Realtime do gameplayu

- LISTEN/NOTIFY ≠ low-latency game loop
- brak deterministycznego orderingu
- brak kontroli nad tickiem
- brak rollbacków
- brak ochrony przed speed-hackiem
- Realtime DB ≠ realtime game

## 1. Supabase (Postgres) – source of truth

### Odpowiedzialności
- Przechowywanie turniejów
- Przechowywanie drabinek
- Przechowywanie rund
- Przechowywanie meczów
- Przechowywanie wyników
- Historia gier
- Reconnect i recovery
- Audyt rozgrywek

### Założenia
- Supabase Realtime NIE jest używane do gameplayu
- Baza danych jest jedynym źródłem prawdy
- Wszystkie kluczowe zdarzenia gry są zapisywane

---

## 2. Go WebSocket Service – game engine

### Odpowiedzialności
- Obsługa połączeń WebSocket
- Logika gry (server-authoritative)
- Walidacja ruchów graczy
- Broadcast stanu gry do klientów
- Zarządzanie cyklem życia gry

### Model wykonania
- Jedna instancja serwisu obsługuje wiele gier
- Jedna gra = jeden goroutine
- Stan gry trzymany w pamięci
- Model tick-based lub event-based

### Przepływ
Client → WebSocket → Go Game Server → (async) → Postgres

---

## 3. Redis – koordynacja (nie source of truth)

### Odpowiedzialności
- Matchmaking
- Kolejki oczekujących (do 250 graczy)
- Presence
- Locki turniejowe
- Routing reconnectów
- Pub/Sub między node’ami

### Ograniczenia
- Redis NIE przechowuje trwałego stanu gry
- Redis NIE jest jedynym źródłem danych

---

## 4. Przebieg gry (2–4 graczy)

### Kroki
1. Utworzenie game_id w bazie danych
2. Go server ładuje minimalny stan gry
3. Stan gry utrzymywany w RAM
4. Każdy ruch:
   - walidacja po stronie serwera
   - broadcast do graczy
   - zapis append-only do bazy danych

### Recovery
- Po crashu serwera:
  - odtworzenie stanu gry przez replay eventów

---

## 5. Turniej

### Odpowiedzialności
- Supabase:
  - generowanie drabinek
  - przechowywanie rund
- Worker (Go lub Edge Function):
  - obserwuje zakończenie gier
  - awansuje zwycięzców
  - tworzy kolejne mecze

---

## 6. Kolejka (do 250 graczy)

### Implementacja
- Redis List lub Sorted Set

### Proces
1. Gracze trafiają do kolejki
2. Worker dobiera 2–4 graczy
3. Tworzenie game_id
4. Przypisanie gry do konkretnego game servera

---

## 7. Reconnect & crash handling

### Reconnect gracza
1. Autoryzacja
2. Pobranie game_id
3. Odtworzenie stanu gry z bazy danych

### Crash serwera
1. Redis wykrywa brak heartbeat
2. Nowy node przejmuje grę
3. Replay eventów z bazy danych

---

## 8. Skalowanie

### Poziome
- Wiele instancji Go WebSocket Service
- Sticky sessions (AWS ALB)
- Routing przez Redis

### Wydajność
- Jedna instancja obsługuje tysiące gier
- Małe pokoje (2–4 graczy)

---

## 9. Minimalny stack produkcyjny

- WebSocket: Go
- Load Balancer: AWS ALB
- Queue / Coordination: Redis
- Database: Supabase Postgres
- Auth: Supabase Auth
- Worker: Go lub Supabase Edge Functions
- Metrics: Prometheus + Grafana
