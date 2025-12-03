WinspireProductType - 
 - IRL, streamerzy oferują live w wyznaczonym miejscu, z wyznaczonymi zasadami
 - Game Streaming, sponsor oferuje nagrodę, w której streamer przeprowadza zawody

IProductContext - Context podmiotu(entity). Specyficzny dla produktu Winspire. Zobacz WinspireProductType

CAMPAING - Potrzebna jest kampania związania z SponsorOffer

SponsorOffer - Sponsor chcąc współpracować ze streamerami, musi coś zaoferować,
pewien przedmiot gratyfikacji, na ten moment w celach weryfikacyjnych, i trochę
ręcznych taka oferta jest dodawana przez naszych sprzedawców, którzy zakontraktowali
umowe. Kontrakt ten to bardzo elastyczna definicja, która zawiera 
- czas istnienia
- przedmiot gratyfikacji, kapitałowy lub przedmiot, czyli co oferuje
- miejsce, miejsca lub globalnie
- nazwa
- opis
- IContext - wybrany produkt


Cups gather a large number of teams to match them into smaller, pre-configured tournaments with opponents of similar skill."


====== UX General ===
# First Specification
Streamer/User czy inny user widzi w [sidebar](https://catalyst.tailwindui.com/docs/sidebar-layout) następujące opcje:
<List>
<Item> Home </Item>
<Item> Events</Item>
<Item> Tournament</Item>
</List>

<List title="Upcoming Events">
<Item> ... </Item>
</List>

## Streamer User Stories

### Streamer tworzy turniej

#### User Scenario

Gdy Streamer jest w zakładce Tournament widzi tabelę, która przedstawia listę turniejów, co więcej
widzi nad tabelą guzik "Stwórz".

**Flow użytkownika:**

1. Streamer wchodzi w zakładkę "Tournament" w sidebarze
2. Widzi tabelę z listą swoich turniejów (lub pustą tabelę jeśli brak)
3. Klika przycisk "Stwórz" nad tabelą
4. Otwiera się formularz tworzenia turnieju
5. Wypełnia wymagane pola i zatwierdza
6. Po utworzeniu widzi popup z podsumowaniem i linkiem do pokoju turnieju
7. Może przejść do pokoju lub zamknąć popup i wrócić do tabeli

#### Formularz tworzenia turnieju

Po kliknięciu "Stwórz" wyświetla się formularz z następującymi polami:

| Pole | Typ | Wymagane | Opis |
|------|-----|----------|------|
| Nazwa | Text input | Tak | Nazwa turnieju widoczna dla uczestników |
| Czas rozpoczęcia | DateTime picker | Tak | Data i godzina startu turnieju |
| Gra | Select (disabled) | Tak | Wybór gry - mock: "Packman" |

**Przyciski formularza:**
- "Utwórz turniej" - zatwierdza formularz
- "Anuluj" - zamyka formularz bez zapisywania

#### Potwierdzenie utworzenia turnieju

Po pomyślnym utworzeniu turnieju wyświetla się popup/modal zawierający:

**Zawartość:**
- Nagłówek: "Turniej utworzony!"
- Podsumowanie danych:
  - Nazwa turnieju
  - Czas rozpoczęcia
  - Wybrana gra
- Link do pokoju turnieju (z możliwością skopiowania do schowka)

**Przyciski:**
- "Przejdź do pokoju" - przekierowuje do pokoju turnieju
- "Zamknij" - zamyka popup i wraca do tabeli turniejów

#### Stany wizualne wierszy tabeli

Wiersze w tabeli turniejów mają różne kolory tła w zależności od stanu turnieju:

| Stan | Kolor | Warunek |
|------|-------|---------|
| Zaplanowany | 🟠 Pomarańczowy | `startTime > Now()` - turniej jeszcze się nie rozpoczął |
| Aktywny | 🟢 Zielony | `startTime <= Now()` && turniej nie jest zakończony |
| Zakończony | 🔴 Czerwony | Wyłoniono zwycięzcę LUB brak aktywności graczy przez 1h |

#### Functional Requirements

- FR1: Formularz tworzenia turnieju musi zawierać pola: nazwa, czas rozpoczęcia, gra
- FR2: Pole "Gra" jest wyłączone (disabled) z domyślną wartością "Packman"
- FR3: Po utworzeniu turnieju wyświetla się popup z podsumowaniem i linkiem do pokoju
- FR4: Link do pokoju można skopiować do schowka jednym kliknięciem
- FR5: Wiersze tabeli zmieniają kolor automatycznie w zależności od stanu turnieju
- FR6: Turniej kończy się automatycznie po 1h braku aktywności graczy

#### Success Criteria

- Streamer może utworzyć turniej w mniej niż 30 sekund
- Link do pokoju turnieju jest dostępny natychmiast po utworzeniu
- Stany wizualne tabeli są aktualizowane w czasie rzeczywistym
- Użytkownik zawsze wie, w jakim stanie jest każdy turniej (zaplanowany/aktywny/zakończony)