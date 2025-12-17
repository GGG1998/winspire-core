# Refactoring
**Cel**: Szablon do sledzenia procesu refaktoryzacji architekturalnej
**Format**: Oparty na `.specify/templates/tasks-template.md`

# Refactoring Tasks: [NAZWA REFAKTORU]

**Branch**: `refactor/[nazwa]` | **Data**: [DATA]
**Cel**: 
- Zmodernizowanie struktury PreLobby dla backendu i zaktualizować frontend
- User MUST USE only latest version
- Zamiast robić UPDATE, będziemy robić INSERT z nowym stanem pre-lobby oraz
**Powod**: [Dlaczego refaktoryzujemy - problemy z obecnym kodem]

## Zakres zmian

### Pliki do modyfikacji
```
[SCIEZKA_1]  # [krotki opis co zmieniac]
[SCIEZKA_2]  # [krotki opis co zmieniac]
[SCIEZKA_3]  # [krotki opis co zmieniac]
```

### Pliki do usuniecia (jezeli dotyczy)
```
[SCIEZKA_DO_USUNIECIA_1]  # [powod usuniecia]
```

### Nowe pliki do stworzenia (jezeli dotyczy)
```
[NOWA_SCIEZKA_1]  # [przeznaczenie nowego pliku]
```

---

## Phase 1: Analiza i przygotowanie

**Cel**: Zrozumienie obecnego kodu i przygotowanie do zmian

- [ ] R001 Przeanalizuj obecna strukture w `[GLOWNY_PLIK]`
- [ ] R002 Zidentyfikuj wszystkie zaleznosci (importy, referencje)
- [ ] R003 [P] Sprawdz testy pokrywajace refaktoryzowany kod
- [ ] R004 [P] Utworz branch `refactor/[nazwa]`

**Checkpoint**: Mamy pelne zrozumienie obecnego kodu i jego zaleznosci

---

## Phase 2: Ekstrakcja / Podzial

**Cel**: Wydzielenie logiki do nowych modulow/plikow

- [ ] R005 Wydziel [NAZWA_MODULU_1] do `[NOWA_SCIEZKA_1]`
- [ ] R006 Wydziel [NAZWA_MODULU_2] do `[NOWA_SCIEZKA_2]`
- [ ] R007 [P] Dodaj eksporty w nowych plikach
- [ ] R008 Zaktualizuj importy w oryginalnym pliku

**Checkpoint**: Nowe moduly istnieja i sa eksportowane

---

## Phase 3: Migracja zaleznosci

**Cel**: Aktualizacja wszystkich plikow uzywajacych refaktoryzowanego kodu

- [ ] R009 Zaktualizuj importy w `[PLIK_ZALEZNY_1]`
- [ ] R010 Zaktualizuj importy w `[PLIK_ZALEZNY_2]`
- [ ] R011 [P] Zaktualizuj importy w testach
- [ ] R012 Usun nieuzywany kod z oryginalnego pliku

**Checkpoint**: Wszystkie zaleznosci uzywaja nowych sciezek importu

---

## Phase 4: Cleanup i walidacja

**Cel**: Upewnienie sie ze refaktor nie wprowadzil bledow

- [ ] R013 Usun puste/nieuzywane pliki
- [ ] R014 [P] Uruchom linter (`make lint` lub `yarn lint`)
- [ ] R015 [P] Uruchom testy (`make test` lub `yarn test`)
- [ ] R016 [P] Uruchom build (`make build` lub `yarn build`)
- [ ] R017 Przetestuj manualnie glowna funkcjonalnosc

**Checkpoint**: Kod dziala identycznie jak przed refaktorem

---

## Zaleznosci miedzy taskami

```
R001-R004 (Phase 1) --> R005-R008 (Phase 2) --> R009-R012 (Phase 3) --> R013-R017 (Phase 4)

Rownolegle [P]:
- R003 || R004 (w Phase 1)
- R007 moze byc rownolegly z R005, R006
- R014 || R015 || R016 (w Phase 4)
```

---

## Notatki

- `[P]` = mozna wykonac rownolegle z innymi taskami [P] w tej samej fazie
- Commituj po kazdej fazie lub logicznej grupie zmian
- Jezeli test lub build failuje, napraw przed przejsciem dalej
- Checkpoint = moment walidacji przed kolejna faza

---

## Przyklad wypelnionego szablonu

```markdown
# Refactoring Tasks: useLobby Hook Split

**Branch**: `refactor/use-lobby-split` | **Data**: 2025-01-15
**Cel**: Podzial monolitycznego hooka useLobby na mniejsze, wyspecjalizowane hooki
**Powod**: useLobby ma 500+ linii, trudny do testowania i utrzymania

## Zakres zmian

### Pliki do modyfikacji
frontends/winspire-app/src/features/lobby/hooks/useLobby.ts  # glowny plik do podzialu
frontends/winspire-app/src/features/lobby/components/LobbyView.tsx  # aktualizacja importow

### Nowe pliki do stworzenia
frontends/winspire-app/src/features/lobby/hooks/useLobbyConnection.ts  # logika WebSocket
frontends/winspire-app/src/features/lobby/hooks/useLobbyState.ts  # zarzadzanie stanem
frontends/winspire-app/src/features/lobby/hooks/useLobbyActions.ts  # akcje uzytkownika
```

---

**Nastepny krok**: Skopiuj ten szablon, wypelnij placeholdery swoimi plikami, i uruchom Claude ponownie z wypelnionym planem.
