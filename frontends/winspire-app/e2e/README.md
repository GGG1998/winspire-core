# E2E Tests - Winspire App

Testy E2E (end-to-end) dla aplikacji Winspire, zbudowane przy użyciu Playwright.

## Uruchamianie testów

### Podstawowe komendy

```bash
# Uruchom wszystkie testy E2E
yarn test:e2e

# Uruchom testy w trybie UI (interaktywny)
yarn test:e2e:ui

# Uruchom testy z widoczną przeglądarką
yarn test:e2e:headed

# Uruchom testy w trybie debugowania
yarn test:e2e:debug

# Zobacz raport z ostatniego uruchomienia
yarn test:e2e:report
```

### Uruchamianie konkretnych testów

```bash
# Tylko testy rejestracji użytkownika
yarn test:e2e user-registration

# Tylko testy rejestracji streamera
yarn test:e2e streamer-registration

# Konkretny test
yarn test:e2e -g "should successfully register a new user"
```

### Uruchamianie na konkretnej przeglądarce

```bash
# Tylko Chrome
yarn test:e2e --project=chromium

# Tylko Firefox
yarn test:e2e --project=firefox

# Tylko Safari
yarn test:e2e --project=webkit
```

## Struktura testów

```
e2e/
├── user-registration.spec.ts      # Testy rejestracji użytkownika
├── streamer-registration.spec.ts  # Testy rejestracji streamera
└── README.md                      # Ten plik
```

## Testy rejestracji użytkownika (`user-registration.spec.ts`)

### Scenariusze testowe:

1. **Wyświetlanie formularza** - Sprawdza czy wszystkie pola są widoczne
2. **Walidacja pustego formularza** - Sprawdza czy wyświetlają się błędy walidacji
3. **Niezgodność haseł** - Testuje komunikat o błędzie przy różnych hasłach
4. **Sukces rejestracji** - Pełny proces rejestracji nowego użytkownika
5. **Rejestracja z opcjonalnymi polami** - Testuje wypełnienie kraju i miasta
6. **Duplikat email** - Sprawdza obsługę już zarejestrowanego emaila
7. **Nawigacja do logowania** - Testuje linki nawigacyjne

## Testy rejestracji streamera (`streamer-registration.spec.ts`)

### Scenariusze testowe:

1. **Wyświetlanie formularza** - Sprawdza czy wszystkie pola są widoczne
2. **Walidacja pustego formularza** - Sprawdza czy wyświetlają się błędy walidacji
3. **Niezgodność haseł** - Testuje komunikat o błędzie przy różnych hasłach
4. **Sukces rejestracji** - Pełny proces rejestracji nowego streamera
5. **Rejestracja z opcjonalnymi polami** - Testuje wypełnienie kraju i miasta
6. **Duplikat email** - Sprawdza obsługę już zarejestrowanego emaila
7. **Nawigacja do logowania** - Testuje linki nawigacyjne
8. **Rozróżnienie streamer vs user** - Weryfikuje specyficzne elementy dla streamera
9. **Typ profilu** - Sprawdza czy tworzony jest prawidłowy typ profilu

## Wymagania

- Aplikacja musi być uruchomiona na `http://localhost:5173` (lub skonfiguruj w `playwright.config.ts`)
- Backend Supabase musi być dostępny i skonfigurowany
- Baza danych musi być w stanie przyjmować nowe rejestracje

## Konfiguracja

Edytuj `playwright.config.ts` aby dostosować:
- Bazowy URL aplikacji
- Timeout dla testów
- Przeglądarki do testowania
- Reporter wyników
- Konfigurację webServer

## Dobre praktyki

1. **Unikalne dane testowe** - Testy używają `Date.now()` do generowania unikalnych emaili i nicków
2. **Izolacja testów** - Każdy test jest niezależny i nie polega na stanie z innych testów
3. **Czyszczenie** - Testy logują się i wylogowują, aby nie wpływać na kolejne testy
4. **Timeout** - Dostosowane timeouty dla operacji backendowych (rejestracja, redirect)
5. **Asercje** - Używamy oczekiwań (expect) z odpowiednimi timeoutami

## Debugowanie

### Tryb debug

```bash
yarn test:e2e:debug
```

To otworzy Playwright Inspector, gdzie możesz:
- Krokować przez test
- Zobacz DOM w czasie rzeczywistym
- Sprawdzić selektory
- Zobacz timeline zdarzeń

### Tryb UI

```bash
yarn test:e2e:ui
```

Interaktywny interfejs pokazujący:
- Wszystkie testy
- Wyniki w czasie rzeczywistym
- Możliwość uruchomienia pojedynczych testów
- Time travel debugging

### Trace viewer

Jeśli test się nie powiedzie, trace jest automatycznie zapisywany. Zobacz go:

```bash
yarn test:e2e:report
```

## Troubleshooting

### Testy timeout

- Zwiększ timeout w `playwright.config.ts`
- Sprawdź czy backend odpowiada
- Sprawdź logi w konsoli przeglądarki

### Elementy nie są znalezione

- Użyj Playwright Inspector do sprawdzenia selektorów
- Sprawdź czy element jest widoczny w odpowiednim czasie
- Dodaj `await page.waitForSelector()` jeśli potrzeba

### Baza danych

- Upewnij się że masz uprawnienia do tworzenia użytkowników
- Sprawdź czy triggery bazodanowe działają poprawnie
- Zobacz logi Supabase dla szczegółów błędów

## CI/CD

Testy są skonfigurowane do uruchomienia w CI z:
- Retry na niepowodzenie (2 razy)
- Sekwencyjne uruchomienie (1 worker)
- Automatyczne generowanie raportów HTML

Konfiguracja CI w `playwright.config.ts` aktywuje się gdy `process.env.CI` jest ustawiony.


